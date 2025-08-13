package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openapiYAML []byte

//go:embed redoc.html
var redocHTML []byte

func OpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openapiYAML)
}

func Redoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(redocHTML)
}
