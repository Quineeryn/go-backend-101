package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	C *redis.Client
}

func NewRedis(addr, pass string, db int) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DB:           db,
		MinIdleConns: 3,
		PoolSize:     20,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		DialTimeout:  500 * time.Millisecond,
	})
	return &Redis{C: rdb}
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.C.Ping(ctx).Err()
}
