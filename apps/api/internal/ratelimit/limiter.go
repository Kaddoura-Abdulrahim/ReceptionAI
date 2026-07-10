package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	attempts map[string][]time.Time
}

func New(limit int, window time.Duration) *Limiter {
	return &Limiter{
		limit:    limit,
		window:   window,
		attempts: make(map[string][]time.Time),
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	cutoff := now.Add(-l.window)
	current := l.attempts[key]
	kept := current[:0]
	for _, at := range current {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	if len(kept) >= l.limit {
		l.attempts[key] = kept
		return false
	}
	l.attempts[key] = append(kept, now)
	return true
}
