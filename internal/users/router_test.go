package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"

	appLogger "github.com/Quineeryn/go-backend-101/internal/logger"
)

func newDBForRouterTest(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uix_users_email ON users(email);`).Error; err != nil {
		t.Fatalf("create unique index: %v", err)
	}
	return db
}

func TestRegisterRoutes_PostUser(t *testing.T) {
	// siapkan logger nop agar Audit tidak panic
	appLogger.L = zap.NewNop()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// inisialisasi handler (pakai Store biasa)
	db := newDBForRouterTest(t)
	store := NewStore(db)
	h := NewHandler(store)

	// panggil RegisterRoutes yang lagi kita cover
	RegisterRoutes(r, h) // pastikan signature cocok: func RegisterRoutes(r *gin.Engine, h *Handler)

	// hit POST /v1/users
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(map[string]any{
		"name":  "RR",
		"email": "rr@example.com",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/users", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d body=%s", w.Code, w.Body.String())
	}
}
