// cmd/api/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	gormpg "gorm.io/driver/postgres"
	gormsqlite "gorm.io/driver/sqlite"

	"gorm.io/gorm"
	_ "modernc.org/sqlite" // penting: register driver "sqlite" (pure Go, no CGO)

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Quineeryn/go-backend-101/internal/auth"
	"github.com/Quineeryn/go-backend-101/internal/cache"
	"github.com/Quineeryn/go-backend-101/internal/config"
	"github.com/Quineeryn/go-backend-101/internal/docs"
	"github.com/Quineeryn/go-backend-101/internal/httpx"
	"github.com/Quineeryn/go-backend-101/internal/logger"
	"github.com/Quineeryn/go-backend-101/internal/middleware"
	"github.com/Quineeryn/go-backend-101/internal/ratelimit"
	"github.com/Quineeryn/go-backend-101/internal/users"
)

func main() {
	_ = godotenv.Load()

	// === config & logger ===
	cfg := config.FromEnv()

	_ = logger.Init(logger.Config{
		Env:        cfg.Env,
		FilePath:   cfg.LogFilePath,
		MaxSizeMB:  cfg.LogMaxSizeMB,
		MaxBackups: cfg.LogMaxBackups,
		MaxAgeDays: cfg.LogMaxAgeDays,
	})
	defer logger.L.Sync()

	appLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(appLogger)

	// === DB (GORM: Postgres atau SQLite) ===
	dsn := cfg.DBDSN
	if dsn == "" {
		dsn = "var/app.db" // fallback dev
	}
	dialect := "sqlite"
	if strings.HasPrefix(dsn, "postgres://") || strings.Contains(dsn, "host=") {
		dialect = "postgres"
	}

	var db *gorm.DB
	var err error

	if dialect == "postgres" {
		db, err = gorm.Open(gormpg.Open(dsn), &gorm.Config{})
		if err != nil {
			slog.Error("db.open.failed", "dialect", "postgres", "err", err)
			os.Exit(1)
		}
		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(30)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	} else {
		ensureDirFor(dsn) // hanya untuk SQLite (file)
		dial := gormsqlite.Dialector{DriverName: "sqlite", DSN: stripSQLiteURI(dsn)}
		db, err = gorm.Open(dial, &gorm.Config{})
		if err != nil {
			slog.Error("db.open.failed", "dialect", "sqlite", "err", err)
			os.Exit(1)
		}
		_ = db.Exec("PRAGMA foreign_keys = ON;").Error
		_ = db.Exec("PRAGMA journal_mode = WAL;").Error
		_ = db.Exec("PRAGMA busy_timeout = 5000;").Error
	}

	// === migrate (DEV only) ===
	if getEnv("AUTO_MIGRATE", "false") == "true" {
		if dialect == "sqlite" {
			if err := db.AutoMigrate(&users.User{}, &auth.RefreshToken{}); err != nil {
				slog.Error("migrate.failed", "err", err)
				os.Exit(1)
			}
		} else {
			slog.Warn("AUTO_MIGRATE ignored on Postgres; run `make migrate-pg` instead")
		}
	}

	// === deps ===
	userStore := users.NewStore(db)
	tokenStore := auth.NewStore(db)
	jwtMgr := &auth.Manager{
		Secret:     []byte(getEnv("JWT_SECRET", "dev-secret-change-me")),
		AccessTTL:  mustParseDur(getEnv("JWT_ACCESS_TTL", "15m")),
		RefreshTTL: mustParseDur(getEnv("JWT_REFRESH_TTL", "168h")),
	}

	// === CP12: Redis client (global) ===
	redisCli := cache.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err := redisCli.Ping(context.Background()); err != nil {
		slog.Warn("redis.ping.failed", "err", err, "addr", cfg.RedisAddr)
	} else {
		slog.Info("redis.connected", "addr", cfg.RedisAddr, "db", cfg.RedisDB)
	}

	// === CP12: users repo dibungkus cache-aside (fallback ke DB-only jika Redis bermasalah) ===
	var usersRepo users.Repo = userStore
	if err := redisCli.Ping(context.Background()); err == nil {
		usersRepo = users.NewCachedStore(userStore, redisCli.C, mustParseDur(getEnv("USERS_CACHE_TTL", "5m")))
	}

	// === HTTP server ===
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// CP11 middlewares
	r.Use(httpx.RequestID())
	r.Use(httpx.AccessLog())
	r.Use(httpx.Metrics())

	// timeout context duluan
	r.Use(timeoutMiddleware(60 * time.Second))

	// middleware CP-06 (urutannya penting)
	r.Use(
		middleware.EnsureCorrelationID(),
		middleware.RequestLogger(),
		middleware.ErrorEnvelope(),
		middleware.RecoveryJSON(),
	)

	// In-memory cache kecil untuk /v1/users/me
	cstore := cache.NewMemory(5 * time.Minute)

	// === CP13: Distributed Rate Limiting (Redis) ===
	// default per-IP-per-route
	{
		rlDefault := ratelimit.NewRedisLimiter(
			redisCli.C,
			mustParseFloat(getEnv("RATE_LIMIT_DEFAULT_RPS", "2")),
			mustParseInt(getEnv("RATE_LIMIT_DEFAULT_BURST", "10")),
			60*time.Second,
		)
		r.Use(ratelimit.MiddlewareRedis(rlDefault, ratelimit.KeyPerIPRoute))
	}

	// health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// docs
	r.GET("/openapi.yaml", gin.WrapF(docs.OpenAPISpec))
	r.GET("/docs", gin.WrapF(docs.Redoc))

	// users routes (handler menerima Repo: store atau cached store)
	users.RegisterRoutes(r, users.NewHandler(usersRepo))

	// auth routes (rate limit login lebih ketat)
	authH := auth.Handler{Users: userStore, Tokens: tokenStore, JWT: jwtMgr}
	v1 := r.Group("/v1")
	{
		v1.POST("/auth/register", authH.Register)

		rlLogin := ratelimit.NewRedisLimiter(
			redisCli.C,
			mustParseFloat(getEnv("RATE_LIMIT_AUTH_RPS", "0.2")), // ~12/min
			mustParseInt(getEnv("RATE_LIMIT_AUTH_BURST", "5")),
			60*time.Second,
		)
		v1.POST("/auth/login",
			ratelimit.MiddlewareRedis(rlLogin, ratelimit.KeyLogin),
			authH.Login,
		)

		v1.POST("/auth/refresh", authH.Refresh)
		v1.POST("/auth/logout", authH.Logout)

		// contoh protected
		v1.GET("/users/me", auth.RequireAuth(jwtMgr), func(c *gin.Context) {
			uid := c.GetString("user_id")
			key := "me:" + uid

			if b, ok := cstore.Get(key); ok {
				etag := cache.WeakETag(b)
				if inm := c.GetHeader("If-None-Match"); inm != "" && inm == etag {
					c.Header("ETag", etag)
					c.Header("X-Cache", "HIT")
					c.Status(http.StatusNotModified)
					return
				}
				c.Header("ETag", etag)
				c.Header("X-Cache", "HIT")
				c.Data(http.StatusOK, "application/json; charset=utf-8", b)
				return
			}

			u, err := userStore.FindByID(c, uid)
			if err != nil {
				c.Status(http.StatusNotFound)
				c.Error(err)
				return
			}
			type mePayload struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			payload := mePayload{ID: u.ID, Name: u.Name, Email: u.Email}
			body, err := json.Marshal(payload)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				c.Error(err)
				return
			}

			cstore.Set(key, body, 30*time.Second)

			etag := cache.WeakETag(body)
			if inm := c.GetHeader("If-None-Match"); inm != "" && inm == etag {
				c.Header("ETag", etag)
				c.Header("X-Cache", "MISS-ETAG-NOTMOD")
				c.Status(http.StatusNotModified)
				return
			}
			c.Header("ETag", etag)
			c.Header("X-Cache", "MISS")
			c.Data(http.StatusOK, "application/json; charset=utf-8", body)
		})

		// Admin-only sample
		v1.GET("/admin/ping",
			auth.RequireAuth(jwtMgr),
			auth.RequireRole("admin"),
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"ok":   true,
					"role": c.GetString("role"),
				})
			},
		)
	}

	// metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// start async
	go func() {
		appLogger.Info("server.starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("server.error", "err", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	cstore.Close()
	appLogger.Info("server.stopped")
}

// timeoutMiddleware: tambah context timeout ke setiap request
func timeoutMiddleware(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// ==== helpers ====
func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustParseDur(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func stripSQLiteURI(dsn string) string {
	if strings.HasPrefix(dsn, "file:") {
		return strings.TrimPrefix(dsn, "file:")
	}
	return dsn
}

func ensureDirFor(dsn string) {
	p := stripSQLiteURI(dsn)
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	if p == "" || p == ":memory:" {
		return
	}
	dir := filepath.Dir(p)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
}

func schemaGuard(db *gorm.DB) error {
	m := db.Migrator()
	type pair struct {
		model any
		col   string
	}
	checks := []pair{
		{&users.User{}, "password_hash"},
		{&users.User{}, "role"},
	}
	if !m.HasTable(&auth.RefreshToken{}) {
		return fmt.Errorf("missing table refresh_tokens")
	}
	for _, c := range checks {
		if !m.HasColumn(c.model, c.col) {
			return fmt.Errorf("missing column %T.%s", c.model, c.col)
		}
	}
	return nil
}

func mustParseFloat(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 1.0
	}
	return v
}

func mustParseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return v
}
