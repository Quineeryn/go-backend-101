package httpx

import (
	"github.com/gin-gonic/gin"

	"github.com/Quineeryn/go-backend-101/internal/apperr"
)

func AbortError(c *gin.Context, op string, err error) {
	if err == nil {
		return
	}

	// simpan error ke gin.Context → ditangani ErrorEnvelope
	c.Error(apperr.Op(op, err)) //nolint:errcheck
	c.Abort()
}
