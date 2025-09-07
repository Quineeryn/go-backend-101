package users

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Quineeryn/go-backend-101/internal/apperr"
	httpx "github.com/Quineeryn/go-backend-101/internal/httpx"
)

// Repo diimplement oleh Store biasa dan CachedStore.
type Repo interface {
	Create(ctx context.Context, u User) (User, error)
	List(ctx context.Context) ([]User, error)
	Get(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, id string, data User) (User, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	store Repo
}

func NewHandler(s Repo) *Handler { return &Handler{store: s} }

// (Opsional) Masih dipertahankan kalau suatu saat mau dipakai untuk non-AppError path.
// Tapi pada versi ini kita tidak memanggilnya lagi.
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
	uid := httpx.CurrentUserID(c)
	action := httpx.ActionUserCreate
	success := false
	msg := ""
	resource := ""

	defer func() {
		httpx.Audit(c, httpx.AuditEvent{
			UserID:   uid,
			Action:   action,
			Resource: resource,
			Success:  success,
			Message:  msg,
		})
	}()

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.AbortError(c, "users.create", apperr.E(apperr.Validation, "invalid request body", err))
		return
	}
	req.Normalize()
	if req.Name == "" || req.Email == "" {
		httpx.AbortError(c, "users.create", apperr.E(apperr.Validation, "name and email are required", nil))
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
			httpx.AbortError(c, "users.create", apperr.E(apperr.Conflict, "email already exists", err))
			return
		}
		httpx.AbortError(c, "users.create", apperr.E(apperr.Internal, "failed to create user", err))
		return
	}

	success = true
	msg = "ok"
	resource = created.ID
	c.JSON(http.StatusCreated, toResponse(created))
}

// GET /v1/users
func (h *Handler) List(c *gin.Context) {
	usersList, err := h.store.List(c.Request.Context())
	if err != nil {
		httpx.AbortError(c, "users.list", apperr.E(apperr.Internal, "failed to list users", err))
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
	uid := httpx.CurrentUserID(c)
	action := httpx.ActionUserView
	success := false
	msg := ""
	id := c.Param("id")
	resource := id

	defer func() {
		httpx.Audit(c, httpx.AuditEvent{
			UserID:   uid,
			Action:   action,
			Resource: resource,
			Success:  success,
			Message:  msg,
		})
	}()

	u, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			httpx.AbortError(c, "users.get", apperr.E(apperr.NotFound, "user not found", err))
			return
		}
		httpx.AbortError(c, "users.get", apperr.E(apperr.Internal, "failed to get user", err))
		return
	}

	success = true
	msg = "ok"
	c.JSON(http.StatusOK, toResponse(u))
}

// PUT /v1/users/:id
func (h *Handler) Update(c *gin.Context) {
	uid := httpx.CurrentUserID(c)
	action := httpx.ActionUserUpdate
	success := false
	msg := ""
	id := c.Param("id")
	resource := id

	defer func() {
		httpx.Audit(c, httpx.AuditEvent{
			UserID:   uid,
			Action:   action,
			Resource: resource,
			Success:  success,
			Message:  msg,
		})
	}()

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.AbortError(c, "users.update", apperr.E(apperr.Validation, "invalid request body", err))
		return
	}
	req.Normalize()
	if req.Name == "" || req.Email == "" {
		httpx.AbortError(c, "users.update", apperr.E(apperr.Validation, "name and email are required", nil))
		return
	}

	data := User{Name: req.Name, Email: req.Email}
	updated, err := h.store.Update(c.Request.Context(), id, data)
	if err != nil {
		switch err {
		case ErrNotFound:
			httpx.AbortError(c, "users.update", apperr.E(apperr.NotFound, "user not found", err))
			return
		case ErrDuplicate:
			httpx.AbortError(c, "users.update", apperr.E(apperr.Conflict, "email already exists", err))
			return
		default:
			httpx.AbortError(c, "users.update", apperr.E(apperr.Internal, "failed to update user", err))
			return
		}
	}

	success = true
	msg = "ok"
	c.JSON(http.StatusOK, toResponse(updated))
}

// DELETE /v1/users/:id
func (h *Handler) Delete(c *gin.Context) {
	uid := httpx.CurrentUserID(c)
	action := httpx.ActionUserDelete
	success := false
	msg := ""
	id := c.Param("id")
	resource := id

	defer func() {
		httpx.Audit(c, httpx.AuditEvent{
			UserID:   uid,
			Action:   action,
			Resource: resource,
			Success:  success,
			Message:  msg,
		})
	}()

	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		if err == ErrNotFound {
			httpx.AbortError(c, "users.delete", apperr.E(apperr.NotFound, "user not found", err))
			return
		}
		httpx.AbortError(c, "users.delete", apperr.E(apperr.Internal, "failed to delete user", err))
		return
	}

	success = true
	msg = "ok"
	c.Status(http.StatusNoContent)
}
