package main

import (
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/Quineeryn/go-backend-101/internal/auth"
	"github.com/Quineeryn/go-backend-101/internal/config"
	"github.com/Quineeryn/go-backend-101/internal/users"

	gormpg "gorm.io/driver/postgres"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // enable sqlite (modernc) when needed
)

func main() {
	// 1) Load .env
	_ = godotenv.Load()

	// 2) Read DSN from the same config as the server
	cfg := config.FromEnv()
	dsn := cfg.DBDSN
	if dsn == "" {
		log.Fatal("DB_DSN is empty: set it in .env or environment")
	}

	// 3) Detect dialect
	isPG := strings.HasPrefix(dsn, "postgres://") || strings.Contains(dsn, "host=")

	var db *gorm.DB
	var err error

	if isPG {
		// ---- PostgreSQL: run idempotent SQL (no AutoMigrate to avoid constraint drama)
		db, err = gorm.Open(gormpg.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("open pg:", err)
		}
		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(15 * time.Minute)

		stmts := []string{
			// users: add needed columns (safe to rerun)
			`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS password_hash TEXT`,
			`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user'`,
			`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now()`,
			`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`,

			// refresh_tokens table (+ indexes)
			`CREATE TABLE IF NOT EXISTS refresh_tokens (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL,
				jti TEXT NOT NULL,
				expires_at TIMESTAMPTZ NOT NULL,
				revoked_at TIMESTAMPTZ,
				created_at TIMESTAMPTZ NOT NULL DEFAULT now()
			)`,
			`CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens(user_id)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS ux_refresh_jti ON refresh_tokens(jti)`,

			// Add FK if not exists (avoid duplicate_object)
			`DO $$ BEGIN
				ALTER TABLE refresh_tokens
				ADD CONSTRAINT fk_refresh_user
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		}
		for _, s := range stmts {
			if err := db.Exec(s).Error; err != nil {
				log.Fatalf("migration failed at:\n%s\nerr: %v", s, err)
			}
		}
	} else {
		// ---- SQLite (dev): AutoMigrate is fine
		db, err = gorm.Open(gormsqlite.Dialector{DriverName: "sqlite", DSN: dsn}, &gorm.Config{})
		if err != nil {
			log.Fatal("open sqlite:", err)
		}
		if err := db.AutoMigrate(&users.User{}, &auth.RefreshToken{}); err != nil {
			log.Fatal("automigrate sqlite:", err)
		}
	}

	log.Println("migration OK")
}
