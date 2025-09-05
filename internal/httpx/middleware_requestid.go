package httpx

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CtxKeyRequestID = "request_id"
const HeaderRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(CtxKeyRequestID, rid)
		c.Writer.Header().Set(HeaderRequestID, rid)
		c.Next()
	}
}
