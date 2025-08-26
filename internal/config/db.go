package config

import (
	"context"
	"log"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// OpenDB membuka koneksi Postgres dengan guardrails:
// - Fail-fast bila DSN kosong
// - Tambah sslmode=disable bila target localhost & tidak diset
// - Set connection pool & ping dengan timeout
func OpenDB(dsn string) *gorm.DB {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		log.Fatal("DB_DSN is empty — set it in environment or .env (e.g. postgres://user:pass@localhost:5432/db?sslmode=disable)")
	}

	// Tambahkan sslmode=disable utk localhost jika belum ada
	// (supaya tidak kena TLS refused pada Postgres lokal)
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		if !strings.Contains(dsn, "sslmode=") {
			u, err := url.Parse(dsn)
			if err == nil {
				host := u.Hostname()
				if host == "localhost" || host == "127.0.0.1" || host == "::1" {
					q := u.Query()
					q.Set("sslmode", "disable")
					u.RawQuery = q.Encode()
					dsn = u.String()
				}
			}
		}
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("cannot connect to db (open): %v", err)
	}

	// Set connection pool & ping dengan timeout
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("cannot get sql DB from gorm: %v", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("cannot connect to db (ping): %v", err)
	}

	return db
}
