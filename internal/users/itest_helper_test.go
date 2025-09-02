//go:build integration

package users_test

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Quineeryn/go-backend-101/internal/users"
)

var itestDB *gorm.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		panic("DB_DSN is empty for integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	itestDB = db

	// Pakai migration file SQL sebagai sumber kebenaran (recommended).
	// Kalau perlu pakai GORM juga (non-destructive), uncomment:
	// _ = users.AutoMigrate(itestDB)

	code := m.Run()
	os.Exit(code)
}

func newRouterIT(t *testing.T) *gin.Engine {
	t.Helper()

	// Bersihkan data antar test (TRUNCATE lebih aman untuk FK; di sini tabel tunggal)
	if err := itestDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY`).Error; err != nil {
		// Kalau pakai TEXT id manual, RESTART IDENTITY nggak ngaruh — tetap aman
	}

	store := users.NewStore(itestDB)
	h := users.NewHandler(store)

	r := gin.New()
	r.Use(gin.Recovery())
	users.RegisterRoutes(r, h)
	return r
}
