package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Quineeryn/go-backend-101/internal/config"
	"github.com/Quineeryn/go-backend-101/internal/users"
	"github.com/go-chi/chi/v5"
)

func newRouter() *chi.Mux {
	cfg := config.FromEnv()
	db := config.OpenDB(cfg.DBDSN)
	store := users.NewStore(db)
	r := chi.NewRouter()
	r.Mount("/v1/users", users.NewRouter(store))
	return r
}

func TestCreateAndList(t *testing.T) {
	r := newRouter()

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
	r := newRouter()

	// create pertama
	body := []byte(`{"name":"Alea","email":"alea@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rec.Code)
	}

	// create kedua dengan email sama -> 409
	req2 := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d (body: %s)", rec2.Code, rec2.Body.String())
	}
}
