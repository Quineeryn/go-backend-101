package users

import (
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Quineeryn/go-backend-101/internal/httpx"
)

var emailRx = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func NewRouter(store *Store) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		list := store.List()
		res := make([]UserResponse, 0, len(list))
		for _, u := range list {
			res = append(res, toResponse(u))
		}
		httpx.WriteJSON(w, http.StatusOK, res)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		if err := httpx.DecodeJSON(r, &req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid json", err.Error())
			return
		}
		req.Normalize()
		if req.Name == "" || req.Email == "" || !emailRx.MatchString(req.Email) {
			httpx.WriteError(w, http.StatusBadRequest, "invalid name or email", nil)
			return
		}
		u := User{ID: uuid.NewString(), Name: req.Name, Email: req.Email}
		httpx.WriteJSON(w, http.StatusCreated, toResponse(store.Create(u)))
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		u, err := store.Get(id)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, "user not found", nil)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, toResponse(u))
	})

	r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req UpdateUserRequest
		if err := httpx.DecodeJSON(r, &req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid json", err.Error())
			return
		}
		req.Normalize()
		if req.Name == "" || req.Email == "" || !emailRx.MatchString(req.Email) {
			httpx.WriteError(w, http.StatusBadRequest, "invalid name or email", nil)
			return
		}
		upd, err := store.Update(id, User{Name: req.Name, Email: req.Email})
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, "user not found", nil)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, toResponse(upd))
	})

	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := store.Delete(id); err != nil {
			httpx.WriteError(w, http.StatusNotFound, "user not found", nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
