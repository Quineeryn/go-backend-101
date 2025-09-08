package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger: logging sederhana (bisa diganti zap/logrus di production).
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		c.Next()

		dur := time.Since(start)
		status := c.Writer.Status()
		traceID, _ := c.Get(ContextTraceID)

		log.Printf("[REQ] %s %s -> %d (%s) trace_id=%v", method, path, status, dur, traceID)
	}
}
