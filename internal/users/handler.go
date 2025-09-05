package users

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	httpx "github.com/Quineeryn/go-backend-101/internal/httpx"
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
	uid := httpx.CurrentUserID(c)
	action := "USER_CREATE"
	success := false
	msg := ""
	resource := "" // nanti diisi user.ID saat berhasil

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
		msg = "invalid request body"
		writeError(c, http.StatusBadRequest, msg, err.Error())
		return
	}
	req.Normalize()
	if req.Name == "" || req.Email == "" {
		msg = "name and email are required"
		writeError(c, http.StatusBadRequest, msg, "")
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
			msg = "duplicate email"
			writeError(c, http.StatusConflict, msg, "email already exists")
			return
		}
		msg = "failed to create user"
		writeError(c, http.StatusInternalServerError, msg, err.Error())
		return
	}

	// sukses
	success = true
	msg = "ok"
	resource = created.ID // pakai ID (hindari PII)
	c.JSON(http.StatusCreated, toResponse(created))
}

// GET /v1/users
func (h *Handler) List(c *gin.Context) {
	// NOTE: biasanya cukup access log. Kalau mau audit read ops, kamu bisa aktifkan:
	// uid := httpx.CurrentUserID(c)
	// action := "USER_LIST"
	// success := false
	// msg := ""
	// resource := "users:list"
	// defer func() {
	// 	httpx.Audit(c, httpx.AuditEvent{
	// 		UserID: uid, Action: action, Resource: resource, Success: success, Message: msg,
	// 	})
	// }()

	usersList, err := h.store.List(c.Request.Context())
	if err != nil {
		// msg = "failed to list users"; success tetap false kalau pakai audit
		writeError(c, http.StatusInternalServerError, "failed to list users", err.Error())
		return
	}
	out := make([]UserResponse, 0, len(usersList))
	for _, u := range usersList {
		out = append(out, toResponse(u))
	}
	// success = true; msg = "ok" (kalau pakai audit)
	c.JSON(http.StatusOK, out)
}

// GET /v1/users/:id
func (h *Handler) Get(c *gin.Context) {
	// Audit read ini opsional—aktifkan kalau aksesnya sensitif (mis. admin lihat profil orang lain)
	uid := httpx.CurrentUserID(c)
	action := "USER_VIEW"
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
			msg = "user not found"
			writeError(c, http.StatusNotFound, msg, "")
			return
		}
		msg = "failed to get user"
		writeError(c, http.StatusInternalServerError, msg, err.Error())
		return
	}

	success = true
	msg = "ok"
	c.JSON(http.StatusOK, toResponse(u))
}

// PUT /v1/users/:id
func (h *Handler) Update(c *gin.Context) {
	uid := httpx.CurrentUserID(c)
	action := "USER_UPDATE"
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
		msg = "invalid request body"
		writeError(c, http.StatusBadRequest, msg, err.Error())
		return
	}
	req.Normalize()
	if req.Name == "" || req.Email == "" {
		msg = "name and email are required"
		writeError(c, http.StatusBadRequest, msg, "")
		return
	}

	data := User{Name: req.Name, Email: req.Email}
	updated, err := h.store.Update(c.Request.Context(), id, data)
	if err != nil {
		switch err {
		case ErrNotFound:
			msg = "user not found"
			writeError(c, http.StatusNotFound, msg, "")
			return
		case ErrDuplicate:
			msg = "duplicate email"
			writeError(c, http.StatusConflict, msg, "email already exists")
			return
		default:
			msg = "failed to update user"
			writeError(c, http.StatusInternalServerError, msg, err.Error())
			return
		}
	}

	success = true
	msg = "ok"
	// resource sudah id (tetap)
	c.JSON(http.StatusOK, toResponse(updated))
}

// DELETE /v1/users/:id
func (h *Handler) Delete(c *gin.Context) {
	uid := httpx.CurrentUserID(c)
	action := "USER_DELETE"
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
			msg = "user not found"
			writeError(c, http.StatusNotFound, msg, "")
			return
		}
		msg = "failed to delete user"
		writeError(c, http.StatusInternalServerError, msg, err.Error())
		return
	}

	success = true
	msg = "ok"
	c.Status(http.StatusNoContent)
}
