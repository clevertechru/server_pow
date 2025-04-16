package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu       sync.Mutex
	rate     int
	capacity int
	tokens   float64
	last     time.Time
}

func NewLimiter(rate int, capacity int) *Limiter {
	return &Limiter{
		rate:     rate,
		capacity: capacity,
		tokens:   float64(capacity),
		last:     time.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.last)
	l.last = now

	l.tokens += float64(elapsed) * float64(l.rate) / float64(time.Second)
	if l.tokens > float64(l.capacity) {
		l.tokens = float64(l.capacity)
	}

	if l.tokens < 1.0 {
		return false
	}

	l.tokens--
	return true
}
