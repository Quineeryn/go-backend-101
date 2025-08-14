package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Quineeryn/go-backend-101/internal/testdb"
	"github.com/Quineeryn/go-backend-101/internal/users"
)

func newRouter(t *testing.T) *gin.Engine {
	t.Helper()

	db := testdb.Open(t)
	if err := users.AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	store := users.NewStore(db)
	h := users.NewHandler(store)

	r := gin.New()
	r.Use(gin.Recovery())
	users.RegisterRoutes(r, h)
	return r
}

func TestCreateAndList(t *testing.T) {
	r := newRouter(t)

	// create
	body := []byte(`{"name":"Alea","email":"alea@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rec.Code)
	}

	// list
	req2 := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec2.Code)
	}
	var got []map[string]any
	if err := json.Unmarshal(rec2.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 user, got %d", len(got))
	}
}

func TestCreate_DuplicateEmail(t *testing.T) {
	r := newRouter(t)

	// create pertama
	body := []byte(`{"name":"Alea","email":"alea@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rec.Code)
	}

	// create kedua (email sama) → 409
	req2 := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d (body: %s)", rec2.Code, rec2.Body.String())
	}
}
