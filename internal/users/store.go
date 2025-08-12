package users

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("user not found")

type Store struct {
	mu    sync.RWMutex
	items map[string]User
}

func NewStore() *Store {
	return &Store{
		items: make(map[string]User),
	}
}

func (s *Store) Create(user User) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[user.ID] = user
	return user
}

func (s *Store) Get(id string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, exists := s.items[id]
	if !exists {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (s *Store) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]User, 0, len(s.items))
	for _, user := range s.items {
		out = append(out, user)
	}
	return out
}

func (s *Store) Update(id string, user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return User{}, ErrNotFound
	}
	user.ID = id
	s.items[id] = user
	return user, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}
