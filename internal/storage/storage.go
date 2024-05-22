package storage

import (
	"context"
	"sync"
	"time"
)

type Storage interface {
	Set(key, value string, ttl time.Duration)
	Get(key string) (value string, ok bool)
	Delete(key string) (ok bool)

	GcLoop(ctx context.Context)
}

type entry struct {
	expiresAt time.Time
	value     string
}

type storage struct {
	kvMap map[string]entry
	mu    sync.RWMutex
}

func NewStorage() Storage {
	return &storage{
		kvMap: make(map[string]entry),
	}
}

func (s *storage) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.kvMap[key] = entry{
		expiresAt: time.Now().Add(ttl),
		value:     value,
	}
}

func (s *storage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.kvMap[key]
	if ok && time.Now().After(value.expiresAt) {
		delete(s.kvMap, key)
		return "", false
	}
	return value.value, ok
}

func (s *storage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, found := s.kvMap[key]
	if found {
		delete(s.kvMap, key)
	}
	return found
}
