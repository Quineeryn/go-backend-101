package logger

import "strings"

var redactedKeys = []string{"password", "passwd", "pwd", "token", "authorization", "secret"}

func Sanitize(m map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		lk := strings.ToLower(k)
		redact := false
		for _, rk := range redactedKeys {
			if strings.Contains(lk, rk) {
				redact = true
				break
			}
		}
		if redact {
			out[k] = "***REDACTED***"
		} else {
			out[k] = v
		}
	}
	return out
}
