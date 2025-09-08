package ratelimit

import (
	"bytes"
	"encoding/json"
	"io"
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

// KeyLogin — gabung IP + email (non-destructive bind, body tetap bisa dipakai handler)
func KeyLogin(c *gin.Context) string {
	var email string

	if c.Request.Body != nil {
		// 1) salin body mentah
		raw, _ := io.ReadAll(c.Request.Body)
		// 2) KEMBALIKAN body supaya handler masih bisa baca
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))

		// 3) parse minimal field email (abaikan error)
		var b struct {
			Email string `json:"email"`
		}
		_ = json.Unmarshal(raw, &b)
		email = b.Email
	}

	email = strings.ToLower(strings.TrimSpace(email))
	return "login:" + clientIP(c) + ":" + email
}
