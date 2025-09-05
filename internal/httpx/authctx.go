// internal/httpx/authctx.go
package httpx

import "github.com/gin-gonic/gin"

const CtxKeyUserID = "user_id" // set ini di middleware JWT-mu

func CurrentUserID(c *gin.Context) string {
	if v, ok := c.Get(CtxKeyUserID); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "anonymous"
}
