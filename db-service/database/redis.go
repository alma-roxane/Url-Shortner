package database

import "sync"

// VisitCache keeps lightweight rolling counters in-memory.
type VisitCache struct {
	mu     sync.RWMutex
	byCode map[string]int64
}

func NewVisitCache() *VisitCache {
	return &VisitCache{byCode: make(map[string]int64)}
}

func (v *VisitCache) Increment(code string) int64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.byCode[code]++
	return v.byCode[code]
}

func (v *VisitCache) Get(code string) int64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.byCode[code]
}
