package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Quineeryn/go-backend-101/internal/middleware"
	"github.com/gin-gonic/gin"
)

type errResp struct {
	TraceID string `json:"trace_id"`
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(
		middleware.EnsureCorrelationID(),
		middleware.RequestLogger(),
		middleware.ErrorEnvelope(),
		middleware.RecoveryJSON(),
	)

	// endpoint yang melempar error 404
	r.GET("/v1/debug/notfound", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
		c.Error(gin.Error{
			Err:  http.ErrNoLocation,
			Type: gin.ErrorTypePublic,
		})
	})

	// endpoint yang panic
	r.GET("/v1/debug/panic", func(c *gin.Context) {
		panic("boom")
	})

	return r
}

func decode(t *testing.T, rr *httptest.ResponseRecorder) errResp {
	t.Helper()
	var got errResp
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v, body=%s", err, rr.Body.String())
	}
	return got
}

func TestErrorEnvelope_NotFound(t *testing.T) {
	r := newTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/v1/debug/notfound", nil)
	req.Header.Set("X-Request-ID", "test-trace-123")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("want %d got %d", http.StatusNotFound, rr.Code)
	}

	got := decode(t, rr)

	if got.TraceID != "test-trace-123" {
		t.Fatalf("trace_id not propagated, got=%q", got.TraceID)
	}
	if got.Code != http.StatusNotFound || got.Error != "Not Found" {
		t.Fatalf("unexpected envelope: %+v", got)
	}
	if got.Message == "" {
		t.Fatalf("message should not be empty")
	}
}

func TestRecoveryJSON_Panic(t *testing.T) {
	r := newTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/v1/debug/panic", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("want %d got %d", http.StatusInternalServerError, rr.Code)
	}

	got := decode(t, rr)

	if got.TraceID == "" {
		t.Fatalf("trace_id should be present")
	}
	if got.Error != "Internal Server Error" {
		t.Fatalf("unexpected error text: %s", got.Error)
	}
	if got.Message == "" {
		t.Fatalf("message should not be empty")
	}
}
