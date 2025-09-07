package httpx

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Quineeryn/go-backend-101/internal/apperr"
)

func AbortError(c *gin.Context, op string, err error) {
	if err == nil {
		return
	}
	status := apperr.StatusFor(err)
	ae := apperr.Op(op, err)

	// simpan error ke gin.Context → ditangani ErrorEnvelope
	c.Error(ae) //nolint:errcheck
	c.AbortWithStatusJSON(status, gin.H{
		"code":     status,
		"error":    http.StatusText(status),
		"message":  ae.Msg,
		"trace_id": c.GetString("request_id"),
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}
