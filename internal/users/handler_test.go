package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Quineeryn/go-backend-101/internal/apperr"
	"github.com/Quineeryn/go-backend-101/internal/httpx"
	appLogger "github.com/Quineeryn/go-backend-101/internal/logger"
)

/***************
 * Test-only middleware & helpers
 ***************/

// Tangkap *apperr.AppError dari Gin, map ke status code, lalu tulis JSON.
func testErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		var ae *apperr.AppError
		for i := len(c.Errors) - 1; i >= 0; i-- {
			if errors.As(c.Errors[i].Err, &ae) {
				status := http.StatusInternalServerError
				switch ae.Kind {
				case apperr.Validation:
					status = http.StatusBadRequest
				case apperr.Conflict:
					status = http.StatusConflict
				case apperr.NotFound:
					status = http.StatusNotFound
				case apperr.Internal:
					status = http.StatusInternalServerError
				}
				// pakai ae.Error() agar aman kalau field internal berubah
				httpx.WriteError(c.Writer, status, ae.Error(), ae.Unwrap())
				return
			}
		}
	}
}

// DB helper tanpa CGO, 1 koneksi, dan unique index email
func newHTTPTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uix_users_email ON users(email);`).Error; err != nil {
		t.Fatalf("create unique index: %v", err)
	}
	return db
}

func newHTTP(t *testing.T) (*gin.Engine, *Store) {
	t.Helper()

	// Matikan logger global supaya httpx.Audit tidak panic.
	appLogger.L = zap.NewNop()

	gin.SetMode(gin.TestMode)
	db := newHTTPTestDB(t)
	store := NewStore(db)
	h := NewHandler(store)

	r := gin.New()
	r.Use(gin.Recovery())

	// Penting: pasang middleware error agar AbortError → status code yang benar.
	r.Use(testErrorMiddleware())

	// Routes yang dites
	r.POST("/v1/users", h.Create)
	r.GET("/v1/users", h.List)
	r.GET("/v1/users/:id", h.Get)
	r.PUT("/v1/users/:id", h.Update)
	r.DELETE("/v1/users/:id", h.Delete)

	return r, store
}

func doJSON(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

/***************
 * Simple envelopes (opsional untuk decode)
 ***************/
type dataEnvelope struct {
	Data any `json:"data"`
}
type errorEnvelope struct {
	Code    int         `json:"code"`
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Time    time.Time   `json:"time"`
}

/***************
 * TESTS
 ***************/

func TestUsers_Create_201(t *testing.T) {
	r, _ := newHTTP(t)

	w := doJSON(r, http.MethodPost, "/v1/users", map[string]any{
		"name":  "Alea",
		"email": "alea@example.com",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d body=%s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"email"`)) ||
		!bytes.Contains(w.Body.Bytes(), []byte(`"alea@example.com"`)) {
		t.Fatalf("response should contain email field: %s", w.Body.String())
	}
}

func TestUsers_Create_409_Duplicate(t *testing.T) {
	r, _ := newHTTP(t)

	_ = doJSON(r, http.MethodPost, "/v1/users", map[string]any{
		"name":  "X",
		"email": "dup@example.com",
	})
	w := doJSON(r, http.MethodPost, "/v1/users", map[string]any{
		"name":  "Y",
		"email": "dup@example.com",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestUsers_Create_400_InvalidPayload(t *testing.T) {
	r, _ := newHTTP(t)

	w := doJSON(r, http.MethodPost, "/v1/users", map[string]any{
		"name":  "NoEmail",
		"email": "",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestUsers_Get_200_Then_404_AfterDelete(t *testing.T) {
	r, store := newHTTP(t)

	// create
	res := doJSON(r, http.MethodPost, "/v1/users", map[string]any{
		"name":  "Temp",
		"email": "temp@example.com",
	})
	if res.Code != http.StatusCreated {
		t.Fatalf("create want 201, got %d body=%s", res.Code, res.Body.String())
	}

	// ambil id dari DB (paling stabil)
	u, err := store.FindByEmail(context.Background(), "temp@example.com")
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}
	id := u.ID

	// GET 200
	w := doJSON(r, http.MethodGet, "/v1/users/"+id, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}

	// DELETE 204
	del := doJSON(r, http.MethodDelete, "/v1/users/"+id, nil)
	if del.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d body=%s", del.Code, del.Body.String())
	}

	// GET lagi → 404
	w2 := doJSON(r, http.MethodGet, "/v1/users/"+id, nil)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("want 404 after delete, got %d body=%s", w2.Code, w2.Body.String())
	}
}

func TestUsers_Update_200_And_409_OnDuplicate(t *testing.T) {
	r, store := newHTTP(t)

	// create dua user
	_ = doJSON(r, http.MethodPost, "/v1/users", map[string]any{
		"name":  "A",
		"email": "a@example.com",
	})
	_ = doJSON(r, http.MethodPost, "/v1/users", map[string]any{
		"name":  "B",
		"email": "b@example.com",
	})

	u2, err := store.FindByEmail(context.Background(), "b@example.com")
	if err != nil {
		t.Fatalf("find u2: %v", err)
	}

	// update sukses
	upd := doJSON(r, http.MethodPut, "/v1/users/"+u2.ID, map[string]any{
		"name":  "B New",
		"email": "bnew@example.com",
	})
	if upd.Code != http.StatusOK {
		t.Fatalf("update want 200, got %d body=%s", upd.Code, upd.Body.String())
	}
	if !bytes.Contains(upd.Body.Bytes(), []byte(`"email":"bnew@example.com"`)) {
		t.Fatalf("updated email not found in body: %s", upd.Body.String())
	}

	// update duplicate → 409
	updDup := doJSON(r, http.MethodPut, "/v1/users/"+u2.ID, map[string]any{
		"name":  "Whatever",
		"email": "a@example.com",
	})
	if updDup.Code != http.StatusConflict {
		t.Fatalf("want 409 duplicate, got %d body=%s", updDup.Code, updDup.Body.String())
	}
}

func TestUsers_List_200_ReturnsArray(t *testing.T) {
	r, _ := newHTTP(t)

	w := doJSON(r, http.MethodGet, "/v1/users", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
	// Minimal check: respons berbentuk array JSON
	if len(w.Body.Bytes()) > 0 && w.Body.Bytes()[0] != '[' {
		t.Fatalf("expected JSON array response, got: %s", w.Body.String())
	}
}
