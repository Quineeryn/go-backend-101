package httpx

import (
	"errors"
	"net/http"

	"github.com/Quineeryn/go-backend-101/internal/apperr"
	"github.com/gin-gonic/gin"
)

// ErrorMiddleware menangkap apperr.AppError dari handler dan menulis JSON response yang konsisten.
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		var ae *apperr.AppError
		for i := len(c.Errors) - 1; i >= 0; i-- {
			if errors.As(c.Errors[i].Err, &ae) {
				status := http.StatusInternalServerError
				switch ae.Kind {
				case apperr.Validation:
					status = http.StatusBadRequest
				case apperr.Conflict:
					status = http.StatusConflict
				case apperr.NotFound:
					status = http.StatusNotFound
				case apperr.Internal:
					status = http.StatusInternalServerError
				}
				WriteError(c.Writer, status, ae.Error(), ae.Unwrap())
				// Hentikan middleware chain karena error sudah di-handle
				c.Abort()
				return
			}
		}
	}
}
