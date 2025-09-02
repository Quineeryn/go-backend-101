//go:build integration

package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegration_Create_List_Duplicate(t *testing.T) {
	r := newRouterIT(t)

	// create
	body := []byte(`{"name":"Alea","email":"alea@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d (%s)", rec.Code, rec.Body.String())
	}

	// duplicate -> 409
	req2 := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("duplicate: want 409, got %d (%s)", rec2.Code, rec2.Body.String())
	}

	// list -> 200 & len=1
	req3 := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d (%s)", rec3.Code, rec3.Body.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(rec3.Body.Bytes(), &got); err != nil {
		t.Fatalf("list: invalid json: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("list: want 1 user, got %d", len(got))
	}
}

func TestIntegration_Get_Update_Delete(t *testing.T) {
	r := newRouterIT(t)

	// create
	body := []byte(`{"name":"Alea","email":"alea@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d (%s)", rec.Code, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	id := created["id"].(string)

	// get -> 200
	reqG := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	recG := httptest.NewRecorder()
	r.ServeHTTP(recG, reqG)
	if recG.Code != http.StatusOK {
		t.Fatalf("get: want 200, got %d (%s)", recG.Code, recG.Body.String())
	}

	// update name & email -> 200
	up := []byte(`{"name":"Alea New","email":"alea.new@example.com"}`)
	reqU := httptest.NewRequest(http.MethodPut, "/v1/users/"+id, bytes.NewReader(up))
	reqU.Header.Set("Content-Type", "application/json")
	recU := httptest.NewRecorder()
	r.ServeHTTP(recU, reqU)
	if recU.Code != http.StatusOK {
		t.Fatalf("update: want 200, got %d (%s)", recU.Code, recU.Body.String())
	}

	// delete -> 204
	reqD := httptest.NewRequest(http.MethodDelete, "/v1/users/"+id, nil)
	recD := httptest.NewRecorder()
	r.ServeHTTP(recD, reqD)
	if recD.Code != http.StatusNoContent {
		t.Fatalf("delete: want 204, got %d (%s)", recD.Code, recD.Body.String())
	}

	// get again -> 404
	reqG2 := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
	recG2 := httptest.NewRecorder()
	r.ServeHTTP(recG2, reqG2)
	if recG2.Code != http.StatusNotFound {
		t.Fatalf("get after delete: want 404, got %d (%s)", recG2.Code, recG2.Body.String())
	}
}
