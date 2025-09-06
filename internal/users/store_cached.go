package users

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Quineeryn/go-backend-101/internal/httpx"
)

type CachedStore struct {
	inner *Store
	rdb   *redis.Client
	ttl   time.Duration
}

func NewCachedStore(inner *Store, rdb *redis.Client, ttl time.Duration) *CachedStore {
	return &CachedStore{inner: inner, rdb: rdb, ttl: ttl}
}

func keyUser(id string) string { return "app:users:" + id }

// Get: cache-aside
func (s *CachedStore) Get(ctx context.Context, id string) (User, error) {
	k := keyUser(id)
	if s.rdb != nil {
		if b, err := s.rdb.Get(ctx, k).Bytes(); err == nil && len(b) > 0 {
			var u User
			if json.Unmarshal(b, &u) == nil {
				httpx.CacheHit.WithLabelValues("user").Inc()
				return u, nil
			}
		} else {
			httpx.CacheMiss.WithLabelValues("user").Inc()
		}
	}

	u, err := s.inner.Get(ctx, id)
	if err != nil {
		return u, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(u); err == nil {
			_ = s.rdb.Set(ctx, k, b, s.ttl).Err()
		}
	}
	return u, nil
}

// List: (opsional) tidak di-cache dulu
func (s *CachedStore) List(ctx context.Context) ([]User, error) {
	return s.inner.List(ctx)
}

// Create: tulis DB, lalu pre-warm cache
func (s *CachedStore) Create(ctx context.Context, u User) (User, error) {
	created, err := s.inner.Create(ctx, u)
	if err != nil {
		return created, err
	}
	if s.rdb != nil {
		k := keyUser(created.ID)
		if b, err := json.Marshal(created); err == nil {
			_ = s.rdb.Set(ctx, k, b, s.ttl).Err()
		}
	}
	return created, nil
}

// Update: tulis DB, lalu SET ulang cache
func (s *CachedStore) Update(ctx context.Context, id string, data User) (User, error) {
	updated, err := s.inner.Update(ctx, id, data)
	if err != nil {
		return updated, err
	}
	if s.rdb != nil {
		k := keyUser(id)
		if b, err := json.Marshal(updated); err == nil {
			_ = s.rdb.Set(ctx, k, b, s.ttl).Err()
		}
	}
	return updated, nil
}

// Delete: hapus DB, lalu DEL cache
func (s *CachedStore) Delete(ctx context.Context, id string) error {
	if err := s.inner.Delete(ctx, id); err != nil {
		return err
	}
	if s.rdb != nil {
		_ = s.rdb.Del(ctx, keyUser(id)).Err()
	}
	return nil
}
