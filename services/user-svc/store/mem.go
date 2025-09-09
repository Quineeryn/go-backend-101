package store

import (
	"sync"
	"time"
)

type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
}

type MemStore struct {
	mu sync.RWMutex
	m  map[string]User
}

func NewMemStore() *MemStore { return &MemStore{m: map[string]User{}} }

func (s *MemStore) Get(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.m[id]
	return u, ok
}

func (s *MemStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]User, 0, len(s.m))
	for _, u := range s.m {
		out = append(out, u)
	}
	return out
}

func (s *MemStore) Create(u User) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[u.ID] = u
	return u
}
