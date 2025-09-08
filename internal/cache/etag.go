package cache

import (
	"crypto/sha256"
	"encoding/hex"
)

func WeakETag(body []byte) string {
	sum := sha256.Sum256(body)
	return `W/"` + hex.EncodeToString(sum[:]) + `"`
}
