package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Quineeryn/go-backend-101/internal/apperr"
	"github.com/Quineeryn/go-backend-101/internal/logger"
)

func ErrorEnvelope() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Jika sudah ada respons < 400, biarkan.
		if c.Writer.Written() && c.Writer.Status() < http.StatusBadRequest {
			return
		}
		// Tidak ada error yang diset handler → biarkan (bisa 404 route, dll).
		if len(c.Errors) == 0 {
			return
		}

		gerr := c.Errors.Last() // *gin.Error
		status := apperr.StatusFor(gerr.Err)
		msg := safeMessage(gerr.Err)
		traceID := c.GetString("request_id")

		logger.L.Error("request.error",
			zap.Int("status", status),
			zap.String("path", c.FullPath()),
			zap.String("request_id", traceID),
			zap.Error(gerr.Err),
		)

		c.JSON(status, gin.H{
			"code":     status,
			"error":    http.StatusText(status),
			"message":  msg,
			"trace_id": traceID,
			"time":     time.Now().UTC().Format(time.RFC3339),
		})
		c.Abort()
	}
}

// Ambil pesan aman dari AppError; fallback generik.
func safeMessage(err error) string {
	var ae *apperr.AppError
	if errors.As(err, &ae) && ae.Msg != "" {
		return ae.Msg
	}
	return "request failed"
}
