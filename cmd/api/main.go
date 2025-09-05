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

	// <--- tambah
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	gormpg "gorm.io/driver/postgres"
	gormsqlite "gorm.io/driver/sqlite"

	"gorm.io/gorm"
	_ "modernc.org/sqlite" // penting: register driver "sqlite" (pure Go, no CGO)

	"github.com/Quineeryn/go-backend-101/internal/auth"
	"github.com/Quineeryn/go-backend-101/internal/cache"
	"github.com/Quineeryn/go-backend-101/internal/config"
	"github.com/Quineeryn/go-backend-101/internal/docs"
	"github.com/Quineeryn/go-backend-101/internal/middleware"
	"github.com/Quineeryn/go-backend-101/internal/ratelimit"
	"github.com/Quineeryn/go-backend-101/internal/users"
)

func main() {
	_ = godotenv.Load()

	// === config & logger ===
	cfg := config.FromEnv() // PORT, DBDSN, dll
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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
		// (opsional) pool tuning
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
		// PRAGMA khusus SQLite
		_ = db.Exec("PRAGMA foreign_keys = ON;").Error
		_ = db.Exec("PRAGMA journal_mode = WAL;").Error
		_ = db.Exec("PRAGMA busy_timeout = 5000;").Error
	}

	// === migrate (aman untuk dev) ===
	// === migrate (DEV only via toggle) ===
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

	// === HTTP server ===
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Jangan pakai gin.Recovery(); kita pakai RecoveryJSON agar output error selalu JSON
	// r.Use(gin.Logger()) // opsional; kalau aktif bisa dobel dengan RequestLogger

	// timeout context duluan
	r.Use(timeoutMiddleware(60 * time.Second))

	// middleware CP-06 (urutannya penting)
	r.Use(
		middleware.EnsureCorrelationID(),
		middleware.RequestLogger(),
		middleware.ErrorEnvelope(),
		middleware.RecoveryJSON(),
	)

	// --- Cache store (in-memory) ---
	cstore := cache.NewMemory(5 * time.Minute)

	// --- Rate limiting ---
	// store in-memory untuk token bucket; GC tiap 10 menit
	rlStore := ratelimit.NewStore(10 * time.Minute)
	// limiter default per IP per route
	defaultRPS := rate.Limit(mustParseFloat(getEnv("RATE_LIMIT_DEFAULT_RPS", "2")))
	defaultBurst := mustParseInt(getEnv("RATE_LIMIT_DEFAULT_BURST", "10"))
	r.Use(ratelimit.Middleware(rlStore, ratelimit.KeyPerIP, defaultRPS, defaultBurst))

	// health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// docs
	r.GET("/openapi.yaml", gin.WrapF(docs.OpenAPISpec))
	r.GET("/docs", gin.WrapF(docs.Redoc))

	// users (pakai handler kamu)
	users.RegisterRoutes(r, users.NewHandler(userStore))

	authH := auth.Handler{Users: userStore, Tokens: tokenStore, JWT: jwtMgr}
	v1 := r.Group("/v1")
	{
		v1.POST("/auth/register", authH.Register)

		// limiter khusus login: per IP + email (lebih ketat)
		authRPS := rate.Limit(mustParseFloat(getEnv("RATE_LIMIT_AUTH_RPS", "0.2"))) // ~12/min
		authBurst := mustParseInt(getEnv("RATE_LIMIT_AUTH_BURST", "5"))
		v1.POST("/auth/login",
			ratelimit.Middleware(rlStore, ratelimit.KeyLogin, authRPS, authBurst),
			authH.Login,
		)

		v1.POST("/auth/refresh", authH.Refresh)
		v1.POST("/auth/logout", authH.Logout)

		// contoh protected
		v1.GET("/users/me", auth.RequireAuth(jwtMgr), func(c *gin.Context) {
			uid := c.GetString("user_id")
			key := "me:" + uid

			if b, ok := cstore.Get(key); ok {
				// ETag check dari cache
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

			// MISS -> ambil dari DB
			u, err := userStore.FindByID(c, uid)
			if err != nil {
				c.Status(http.StatusNotFound)
				c.Error(err)
				return
			}
			// build body JSON manual (agar bisa cache byte result)
			payload := map[string]any{
				"id":    u.ID,
				"name":  u.Name,
				"email": u.Email,
			}
			body, err := json.Marshal(payload)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				c.Error(err)
				return
			}

			// simpan ke cache (TTL 30s)
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
		logger.Info("server.starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server.error", "err", err)
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
	logger.Info("server.stopped")
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

// buang prefix "file:" (Windows suka rewel dengan URI)
func stripSQLiteURI(dsn string) string {
	if strings.HasPrefix(dsn, "file:") {
		return strings.TrimPrefix(dsn, "file:")
	}
	return dsn
}

// pastikan folder untuk file DB ada
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

// Pastikan kolom/tabel penting sudah ada saat AUTO_MIGRATE=false
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
	// tabel refresh_tokens minimal harus ada
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
