package connlimit

import (
	"sync"
)

type Limiter struct {
	mu    sync.Mutex
	limit int
	count int
}

func NewLimiter(limit int) *Limiter {
	return &Limiter{
		limit: limit,
	}
}

func (l *Limiter) Acquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.count >= l.limit {
		return false
	}

	l.count++
	return true
}

func (l *Limiter) Release() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.count > 0 {
		l.count--
	}
}

func (l *Limiter) Count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.count
}
