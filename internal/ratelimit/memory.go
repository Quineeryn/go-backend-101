package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type entry struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

type Store struct {
	mu       sync.Mutex
	items    map[string]*entry
	ttl      time.Duration
	quit     chan struct{}
	capacity int // unused but kept for future metrics
}

func NewStore(ttl time.Duration) *Store {
	s := &Store{
		items: make(map[string]*entry),
		ttl:   ttl,
		quit:  make(chan struct{}),
	}
	go s.gc()
	return s
}

func (s *Store) Get(key string, rps rate.Limit, burst int) *rate.Limiter {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.items[key]; ok {
		e.lastSeen = now
		// kalau setting berubah, update limiter
		if e.lim.Limit() != rps || e.lim.Burst() != burst {
			e.lim.SetLimit(rps)
			e.lim.SetBurst(burst)
		}
		return e.lim
	}
	lim := rate.NewLimiter(rps, burst)
	s.items[key] = &entry{lim: lim, lastSeen: now}
	return lim
}

func (s *Store) gc() {
	t := time.NewTicker(s.ttl / 2)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			now := time.Now()
			s.mu.Lock()
			for k, e := range s.items {
				if now.Sub(e.lastSeen) > s.ttl {
					delete(s.items, k)
				}
			}
			s.mu.Unlock()
		case <-s.quit:
			return
		}
	}
}

func (s *Store) Close() { close(s.quit) }
