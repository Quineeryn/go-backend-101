package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Quineeryn/go-backend-101/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RequireAuth(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.Status(http.StatusUnauthorized)
			c.Error(errors.New("missing or invalid Authorization header"))
			c.Abort()
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		claims, err := mgr.Parse(raw)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Error(err)
			c.Abort()
			return
		}
		// inject ke context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		// korelasikan trace id di header
		if v, ok := c.Get(middleware.ContextTraceID); ok {
			c.Writer.Header().Set(middleware.HeaderRequestID, v.(string))
		}
		c.Next()
	}
}
