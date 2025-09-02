package auth

import (
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
			c.Error(gin.Error{Err: http.ErrNoCookie, Type: gin.ErrorTypePublic})
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		claims, err := mgr.Parse(raw)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Error(err)
			return
		}
		// inject ke context
		c.Set("user_id", claims.UserID)
		// korelasikan trace id di header
		if v, ok := c.Get(middleware.ContextTraceID); ok {
			c.Writer.Header().Set(middleware.HeaderRequestID, v.(string))
		}
		c.Next()
	}
}
