package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRole(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "" {
			c.Status(http.StatusUnauthorized)
			c.Error(errors.New("missing role in token"))
			c.Abort()
			return
		}
		for _, a := range allowed {
			if role == a {
				c.Next()
				return
			}
		}
		c.Status(http.StatusForbidden)
		c.Error(errors.New("forbidden for role: " + role))
		c.Abort()
	}
}
