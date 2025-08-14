package users

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Quineeryn/go-backend-101/internal/httpx"
)

// simple email regex (cukup untuk kebanyakan kasus)
var emailRx = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// NewRouter mendaftarkan semua route /v1/users
func NewRouter(store *Store) chi.Router {
	r := chi.NewRouter()

	// GET /v1/users
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		list := store.List()
		res := make([]UserResponse, 0, len(list))
		for _, u := range list {
			res = append(res, toResponse(u))
		}
		httpx.WriteJSON(w, http.StatusOK, res)
	})

	// POST /v1/users
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

		u := User{
			ID:    uuid.NewString(),
			Name:  req.Name,
			Email: req.Email,
		}

		created, err := store.Create(u)
		if err != nil {
			// ➜ map duplicate → 409 Conflict
			if errors.Is(err, ErrDuplicate) {
				httpx.WriteError(w, http.StatusConflict, "email already exists", nil)
				return
			}
			// error lain → 500
			httpx.WriteError(w, http.StatusInternalServerError, "failed to create user", err.Error())
			return
		}

		httpx.WriteJSON(w, http.StatusCreated, toResponse(created))
	})

	// GET /v1/users/{id}
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		u, err := store.Get(id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				httpx.WriteError(w, http.StatusNotFound, "user not found", nil)
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, "failed to get user", err.Error())
			return
		}

		httpx.WriteJSON(w, http.StatusOK, toResponse(u))
	})

	// PUT /v1/users/{id}
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
			switch {
			case errors.Is(err, ErrNotFound):
				httpx.WriteError(w, http.StatusNotFound, "user not found", nil)
				return
			case errors.Is(err, ErrDuplicate):
				httpx.WriteError(w, http.StatusConflict, "email already exists", nil)
				return
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "failed to update user", err.Error())
				return
			}
		}

		httpx.WriteJSON(w, http.StatusOK, toResponse(upd))
	})

	// DELETE /v1/users/{id}
	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if err := store.Delete(id); err != nil {
			if errors.Is(err, ErrNotFound) {
				httpx.WriteError(w, http.StatusNotFound, "user not found", nil)
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, "failed to delete user", err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
