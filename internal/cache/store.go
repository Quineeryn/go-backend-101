package cache

import (
	"sync"
	"time"
)

type Store interface {
	Get(key string) (val []byte, ok bool)
	Set(key string, val []byte, ttl time.Duration)
	Close()
}

type memEntry struct {
	val   []byte
	expAt time.Time
}

type Memory struct {
	mu    sync.Mutex
	items map[string]memEntry
	quit  chan struct{}
	ttlGC time.Duration
}

func NewMemory(ttlGC time.Duration) *Memory {
	m := &Memory{
		items: make(map[string]memEntry),
		quit:  make(chan struct{}),
		ttlGC: ttlGC,
	}
	go m.gc()
	return m
}

func (m *Memory) Get(key string) ([]byte, bool) {
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.items[key]
	if !ok || now.After(e.expAt) {
		if ok {
			delete(m.items, key)
		}
		return nil, false
	}
	return e.val, true
}

func (m *Memory) Set(key string, val []byte, ttl time.Duration) {
	m.mu.Lock()
	m.items[key] = memEntry{val: val, expAt: time.Now().Add(ttl)}
	m.mu.Unlock()
}

func (m *Memory) Close() { close(m.quit) }

func (m *Memory) gc() {
	t := time.NewTicker(m.ttlGC)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			now := time.Now()
			m.mu.Lock()
			for k, v := range m.items {
				if now.After(v.expAt) {
					delete(m.items, k)
				}
			}
			m.mu.Unlock()
		case <-m.quit:
			return
		}
	}
}
