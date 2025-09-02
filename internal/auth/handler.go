package auth

import (
	"net/http"
	"time"

	"github.com/Quineeryn/go-backend-101/internal/password"
	"github.com/Quineeryn/go-backend-101/internal/users"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Users  *users.Store
	Tokens *Store
	JWT    *Manager
}

func (h *Handler) Register(c *gin.Context) {
	var in struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(err)
		return
	}
	ph, err := password.Hash(in.Password)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(err)
		return
	}

	u, err := h.Users.Create(c, users.User{
		ID:           uuid.New().String(),
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: &ph,
	})
	if err != nil {
		c.Status(http.StatusConflict)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "name": u.Name, "email": u.Email})
}

func (h *Handler) Login(c *gin.Context) {
	var in struct{ Email, Password string }
	if err := c.ShouldBindJSON(&in); err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(err)
		return
	}
	u, err := h.Users.FindByEmail(c, in.Email)
	if err != nil || u.PasswordHash == nil || !password.Verify(*u.PasswordHash, in.Password) {
		c.Status(http.StatusUnauthorized)
		c.Error(err)
		return
	}

	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	access, err := h.JWT.SignAccess(u.ID, accessJTI)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(err)
		return
	}

	refresh, err := h.JWT.SignRefresh(u.ID, refreshJTI)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(err)
		return
	}

	// simpan refresh aktif
	if err := h.Tokens.Save(c, &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    u.ID,
		JTI:       refreshJTI,
		ExpiresAt: time.Now().UTC().Add(h.JWT.RefreshTTL),
	}); err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": access, "refresh_token": refresh})
}

func (h *Handler) Refresh(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(err)
		return
	}
	claims, err := h.JWT.Parse(in.RefreshToken)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.Error(err)
		return
	}

	// cek refresh masih aktif
	active, err := h.Tokens.IsActive(c, claims.ID)
	if err != nil || !active {
		c.Status(http.StatusUnauthorized)
		c.Error(err)
		return
	}

	// rotate: revoke jti lama, issue yang baru
	_ = h.Tokens.RevokeByJTI(c, claims.ID)

	newJTI := uuid.New().String()
	newAccess, err := h.JWT.SignAccess(claims.UserID, uuid.New().String())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(err)
		return
	}
	newRefresh, err := h.JWT.SignRefresh(claims.UserID, newJTI)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(err)
		return
	}

	_ = h.Tokens.Save(c, &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    claims.UserID,
		JTI:       newJTI,
		ExpiresAt: time.Now().UTC().Add(h.JWT.RefreshTTL),
	})

	c.JSON(http.StatusOK, gin.H{"access_token": newAccess, "refresh_token": newRefresh})
}

func (h *Handler) Logout(c *gin.Context) {
	var in struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(err)
		return
	}
	claims, err := h.JWT.Parse(in.RefreshToken)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.Error(err)
		return
	}
	if err := h.Tokens.RevokeByJTI(c, claims.ID); err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
