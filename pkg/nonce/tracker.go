package nonce

import (
	"sync"
	"time"
)

type Tracker struct {
	mu     sync.Mutex
	nonces map[uint64]time.Time
	window time.Duration
}

func NewTracker(window time.Duration) *Tracker {
	return &Tracker{
		nonces: make(map[uint64]time.Time),
		window: window,
	}
}

func (t *Tracker) IsValid(nonce uint64, timestamp int64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if nonce was already used and not expired
	if ts, exists := t.nonces[nonce]; exists {
		if time.Since(ts) <= t.window {
			return false
		}
		// If nonce exists but expired, remove it
		delete(t.nonces, nonce)
	}

	// Record the nonce
	t.nonces[nonce] = time.Now()
	return true
}
