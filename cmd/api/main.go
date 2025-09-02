package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Quineeryn/go-backend-101/internal/config"
	"github.com/Quineeryn/go-backend-101/internal/docs"
	"github.com/Quineeryn/go-backend-101/internal/middleware"
	"github.com/Quineeryn/go-backend-101/internal/users"
)

func main() {
	godotenv.Load() // load .env file if exists
	cfg := config.FromEnv()
	db := config.OpenDB(cfg.DBDSN)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Migrate schema sebelum start server
	// if err := users.AutoMigrate(db); err != nil {
	// 	slog.Error("migrate.failed", "err", err)
	// 	os.Exit(1)
	// }

	// ==== Gin router ====
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(timeoutMiddleware(60 * time.Second)) // hard timeout per request (opsional)

	// Healthcheck
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Users
	store := users.NewStore(db)

	r.Use(
		middleware.RequestLogger(),
		middleware.ErrorEnvelope(),
		middleware.EnsureCorrelationID(),
		middleware.RecoveryJSON(),
	)

	users.RegisterRoutes(r, users.NewHandler(store)) // akan mendaftarkan /v1/users (tanpa trailing slash)

	// Docs (handler kamu kemungkinan http.HandlerFunc → bungkus pakai WrapF)
	r.GET("/openapi.yaml", gin.WrapF(docs.OpenAPISpec))
	r.GET("/docs", gin.WrapF(docs.Redoc))

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start async
	go func() {
		logger.Info("server.starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server.error", "err", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("server.stopped")
}

// timeoutMiddleware: menambahkan context timeout ke setiap request.
func timeoutMiddleware(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
