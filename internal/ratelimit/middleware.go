package ratelimit

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Middleware: tolak request jika melebihi limit.
// Pasang header: Retry-After, X-RateLimit-Limit, X-RateLimit-Remaining (approx).
func Middleware(store *Store, keyFn func(*gin.Context) string, rps rate.Limit, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFn(c)
		lim := store.Get(key, rps, burst)

		// ✅ konsumsi 1 token kalau boleh sekarang
		if lim.Allow() {
			c.Next()
			return
		}

		// ❌ tidak boleh sekarang → hitung delay berikutnya tanpa mengkonsumsi token
		res := lim.ReserveN(time.Now(), 1)
		if !res.OK() {
			c.Header("Retry-After", "1")
			c.Status(http.StatusTooManyRequests)
			_ = c.Error(errors.New("rate limit exceeded"))
			c.Abort()
			return
		}
		delay := res.DelayFrom(time.Now())
		res.CancelAt(time.Now()) // jangan konsumsi—kita hanya mau tau delay

		sec := int(delay / time.Second)
		if sec < 1 {
			sec = 1
		}
		c.Header("Retry-After", itoa(sec))
		c.Status(http.StatusTooManyRequests)
		_ = c.Error(errors.New("rate limit exceeded"))
		c.Abort()
	}
}

func formatRetryAfter(d time.Duration) string {
	// detik bulat
	sec := int(d.Round(time.Second) / time.Second)
	if sec < 1 {
		sec = 1
	}
	return itoa(sec)
}

func itoa(i int) string {
	// kecil, hindari import strconv
	if i == 0 {
		return "0"
	}
	var b [12]byte
	pos := len(b)
	n := i
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		pos--
		b[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
