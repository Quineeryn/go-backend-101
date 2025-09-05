package ratelimit

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func clientIP(c *gin.Context) string {
	// honor X-Forwarded-For if behind proxy (dev: langsung RemoteIP)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.ClientIP()
	}
	return ip
}

// KeyPerIP per route (default)
func KeyPerIP(c *gin.Context) string {
	return "ip:" + clientIP(c) + ":path:" + c.FullPath()
}

// KeyLogin — gabung IP + email (kalau ada di payload) untuk limiter login
func KeyLogin(c *gin.Context) string {
	type body struct {
		Email string `json:"email"`
	}
	var b body
	_ = c.ShouldBindJSON(&b) // ignore error; hanya bantu key
	// reset raw body tidak dibutuhkan karena ShouldBindJSON di login akan jalan lagi (gin cached)
	email := strings.ToLower(strings.TrimSpace(b.Email))
	return "login:" + clientIP(c) + ":" + email
}
