package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	TraceID string    `json:"trace_id"`
	Code    int       `json:"code"`
	Error   string    `json:"error"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Time    time.Time `json:"time"`
}

// ErrorEnvelope membungkus semua error Gin ke JSON standar.
// RecoveryJSON mengubah panic jadi 500 JSON (bukan HTML).
func ErrorEnvelope() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		status := c.Writer.Status()
		if status < 400 {
			status = http.StatusInternalServerError
			c.Status(status)
		}

		traceID, _ := c.Get(ContextTraceID)
		err := c.Errors.Last() // pesan terakhir yang paling relevan

		c.AbortWithStatusJSON(status, ErrorResponse{
			TraceID: toString(traceID),
			Code:    status,
			Error:   http.StatusText(status),
			Message: err.Error(),
			Time:    time.Now().UTC(),
		})
	}
}

func RecoveryJSON() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		traceID, _ := c.Get(ContextTraceID)

		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
			TraceID: toString(traceID),
			Code:    http.StatusInternalServerError,
			Error:   http.StatusText(http.StatusInternalServerError),
			Message: "internal server error",
			Time:    time.Now().UTC(),
			// Details sengaja tidak ditampilkan biar aman (tidak bocor stacktrace)
		})
	})
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
