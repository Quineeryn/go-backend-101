package users

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	store *Store
}

func NewHandler(s *Store) *Handler { return &Handler{store: s} }

type errorResponse struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
	Time    string `json:"time"`
}

func writeError(c *gin.Context, status int, msg, details string) {
	c.JSON(status, errorResponse{
		Code:    status,
		Error:   http.StatusText(status),
		Message: msg,
		Details: details,
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}

// POST /v1/users
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	req.Normalize()
	if req.Name == "" || req.Email == "" {
		writeError(c, http.StatusBadRequest, "name and email are required", "")
		return
	}

	u := User{
		ID:    uuid.NewString(),
		Name:  req.Name,
		Email: req.Email,
	}
	created, err := h.store.Create(c.Request.Context(), u)
	if err != nil {
		if err == ErrDuplicate {
			writeError(c, http.StatusConflict, "duplicate email", "email already exists")
			return
		}
		writeError(c, http.StatusInternalServerError, "failed to create user", err.Error())
		return
	}

	c.JSON(http.StatusCreated, toResponse(created))
}

// GET /v1/users
func (h *Handler) List(c *gin.Context) {
	usersList, err := h.store.List(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to list users", err.Error())
		return
	}
	out := make([]UserResponse, 0, len(usersList))
	for _, u := range usersList {
		out = append(out, toResponse(u))
	}
	c.JSON(http.StatusOK, out)
}

// GET /v1/users/:id
func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	u, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			writeError(c, http.StatusNotFound, "user not found", "")
			return
		}
		writeError(c, http.StatusInternalServerError, "failed to get user", err.Error())
		return
	}
	c.JSON(http.StatusOK, toResponse(u))
}

// PUT /v1/users/:id
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	req.Normalize()
	if req.Name == "" || req.Email == "" {
		writeError(c, http.StatusBadRequest, "name and email are required", "")
		return
	}

	data := User{Name: req.Name, Email: req.Email}
	updated, err := h.store.Update(c.Request.Context(), id, data)
	if err != nil {
		switch err {
		case ErrNotFound:
			writeError(c, http.StatusNotFound, "user not found", "")
			return
		case ErrDuplicate:
			writeError(c, http.StatusConflict, "duplicate email", "email already exists")
			return
		default:
			writeError(c, http.StatusInternalServerError, "failed to update user", err.Error())
			return
		}
	}
	c.JSON(http.StatusOK, toResponse(updated))
}

// DELETE /v1/users/:id
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		if err == ErrNotFound {
			writeError(c, http.StatusNotFound, "user not found", "")
			return
		}
		writeError(c, http.StatusInternalServerError, "failed to delete user", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
