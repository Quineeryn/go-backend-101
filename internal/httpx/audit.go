package httpx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Quineeryn/go-backend-101/internal/logger"
)

type AuditEvent struct {
	UserID   string
	Action   string
	Resource string
	Success  bool
	Message  string
}

func Audit(c *gin.Context, ev AuditEvent) {
	rid, _ := c.Get(CtxKeyRequestID)
	route := c.FullPath()
	method := c.Request.Method
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	status := c.Writer.Status()

	logger.L.Info("audit",
		zap.Any("request_id", rid),
		zap.String("route", route),
		zap.String("method", method),
		zap.Int("status", status),
		zap.String("ip", ip),
		zap.String("ua", ua),
		zap.String("user_id", ev.UserID),
		zap.String("action", ev.Action),
		zap.String("resource", ev.Resource),
		zap.Bool("success", ev.Success),
		zap.String("message", ev.Message),
	)
}
