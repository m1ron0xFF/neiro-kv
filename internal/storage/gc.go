package storage

import (
	"context"
	"log"
	"time"
)

const gcInterval = time.Second

func (s *storage) GcLoop(ctx context.Context) {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("GC loop done")
			return
		case <-ticker.C:
			log.Println("GC loop started")
			s.mu.Lock()
			log.Printf("lock acquired, len before: %d\n", len(s.kvMap))

			startTime := time.Now()
			for key := range s.kvMap {
				if time.Now().After(s.kvMap[key].expiresAt) {
					delete(s.kvMap, key)
				}
			}

			log.Printf("len after: %d, elapsed %v\n", len(s.kvMap), time.Since(startTime))
			s.mu.Unlock()
		}
	}
}
