package httpx

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Quineeryn/go-backend-101/internal/logger"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// skip logging untuk /metrics biar gak spam
		if c.FullPath() == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		dur := time.Since(start)

		rid, _ := c.Get(CtxKeyRequestID)

		logger.L.Info("access",
			zap.Any("request_id", rid),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("bytes_out", int64(c.Writer.Size())),
			zap.Duration("latency", dur),
			zap.String("ip", c.ClientIP()),
			zap.String("ua", c.Request.UserAgent()),
		)
	}
}
