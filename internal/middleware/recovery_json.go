package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Quineeryn/go-backend-101/internal/apperr"
	"github.com/Quineeryn/go-backend-101/internal/logger"
)

func RecoveryJSON() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, rec any) {
		err := apperr.E(apperr.Internal, "unexpected server error", nil)
		traceID := c.GetString("request_id")
		logger.L.Error("panic",
			zap.Any("recover", rec),
			zap.String("request_id", traceID),
			zap.String("path", c.FullPath()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":     http.StatusInternalServerError,
			"error":    http.StatusText(http.StatusInternalServerError),
			"message":  err.Msg,
			"trace_id": traceID,
			"time":     time.Now().UTC().Format(time.RFC3339),
		})
		c.Abort()
	})
}
