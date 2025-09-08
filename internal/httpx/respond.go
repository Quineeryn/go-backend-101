package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type JSON = map[string]any

func WriteJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, code int, msg string, details any) {
	WriteJSON(w, code, JSON{
		"error":   http.StatusText(code),
		"message": msg,
		"code":    code,
		"time":    time.Now().UTC().Format(time.RFC3339),
		"details": details,
	})
}

func DecodeJSON(r *http.Request, dst any) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return errors.New("content type must be application/json")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
