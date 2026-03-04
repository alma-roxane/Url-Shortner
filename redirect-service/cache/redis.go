package cache

import (
	"sync"
	"time"
)

type entry struct {
	longURL   string
	expiresAt *time.Time
	cachedAt  time.Time
}

// Store is a simple in-memory cache used as a fast redirect lookup layer.
type Store struct {
	mu   sync.RWMutex
	data map[string]entry
	ttl  time.Duration
}

func NewStore(ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Store{data: make(map[string]entry), ttl: ttl}
}

func (s *Store) Get(code string) (string, *time.Time, bool) {
	s.mu.RLock()
	item, ok := s.data[code]
	s.mu.RUnlock()
	if !ok {
		return "", nil, false
	}
	if time.Since(item.cachedAt) > s.ttl {
		s.mu.Lock()
		delete(s.data, code)
		s.mu.Unlock()
		return "", nil, false
	}
	if item.expiresAt != nil && item.expiresAt.Before(time.Now().UTC()) {
		s.mu.Lock()
		delete(s.data, code)
		s.mu.Unlock()
		return "", nil, false
	}
	return item.longURL, item.expiresAt, true
}

func (s *Store) Set(code, longURL string, expiresAt *time.Time) {
	s.mu.Lock()
	s.data[code] = entry{longURL: longURL, expiresAt: expiresAt, cachedAt: time.Now().UTC()}
	s.mu.Unlock()
}
