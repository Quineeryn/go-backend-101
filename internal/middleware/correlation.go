package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderRequestID = "X-Request-ID"
	ContextTraceID  = "trace_id"
)

// EnsureCorrelationID memastikan setiap request punya Trace/Request ID.
// Jika header X-Request-ID ada, dipakai; kalau tidak, generate UUID.
func EnsureCorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(HeaderRequestID)
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set(ContextTraceID, traceID)
		c.Writer.Header().Set(HeaderRequestID, traceID)
		c.Next()
	}
}
