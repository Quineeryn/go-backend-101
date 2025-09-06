package ratelimit

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	rdb   *redis.Client
	rps   float64
	burst int64
	// TTL cadangan untuk key baru/idle
	ttl time.Duration
}

func NewRedisLimiter(rdb *redis.Client, rps float64, burst int, ttl time.Duration) *RedisLimiter {
	return &RedisLimiter{rdb: rdb, rps: rps, burst: int64(burst), ttl: ttl}
}

// Key builder (bisa disesuaikan)
func KeyPerIPRoute(c *gin.Context) string {
	ip := c.ClientIP()
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}
	return "rl:ip:" + ip + ":" + route
}

// Lua script: token-bucket
// KEYS[1]=bucket key
// ARGV[1]=now(ms) ARGV[2]=rate(tokens/sec) ARGV[3]=burst ARGV[4]=ttl(sec)
var lua = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local burst = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

-- fields: tokens, last (ms)
local data = redis.call("HMGET", key, "tokens", "last")
local tokens = tonumber(data[1])
local last = tonumber(data[2])

if not tokens or not last then
  tokens = burst
  last = now
else
  local delta = math.max(0, now - last) / 1000.0
  local filled = tokens + (delta * rate)
  if filled > burst then filled = burst end
  tokens = filled
  last = now
end

local allowed = 0
if tokens >= 1.0 then
  tokens = tokens - 1.0
  allowed = 1
end

redis.call("HMSET", key, "tokens", tokens, "last", last)
redis.call("EXPIRE", key, ttl)
return {allowed, tokens}
`)

func (l *RedisLimiter) Allow(ctx context.Context, key string) (bool, float64, error) {
	now := time.Now().UnixMilli()
	res, err := lua.Run(ctx, l.rdb, []string{key},
		now, l.rps, l.burst, int(l.ttl.Seconds())).Result()
	if err != nil {
		return true, 0, err // fail-open
	}
	arr := res.([]interface{})
	allowed := arr[0].(int64) == 1
	remain := arr[1].(float64)
	return allowed, remain, nil
}

// Gin middleware
func MiddlewareRedis(l *RedisLimiter, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// lewati /metrics biar Prometheus gak kena limit
		if c.FullPath() == "/metrics" {
			c.Next()
			return
		}
		key := keyFn(c)
		ok, _, err := l.Allow(c, key)
		if err != nil {
			// gagal konek redis => fail-open
			c.Next()
			return
		}
		if !ok {
			c.Header("Retry-After", "1")
			c.JSON(429, gin.H{
				"code":    429,
				"error":   "Too Many Requests",
				"message": "rate limit exceeded",
				"time":    time.Now().UTC().Format(time.RFC3339),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
