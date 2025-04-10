package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimiter_Allow(t *testing.T) {
	tests := []struct {
		name     string
		rate     float64
		capacity int64
		requests int
		wait     time.Duration
		want     bool
	}{
		{
			name:     "single request allowed",
			rate:     1.0,
			capacity: 1,
			requests: 1,
			want:     true,
		},
		{
			name:     "burst allowed",
			rate:     1.0,
			capacity: 5,
			requests: 5,
			want:     true,
		},
		{
			name:     "exceed burst",
			rate:     1.0,
			capacity: 5,
			requests: 6,
			want:     false,
		},
		{
			name:     "rate limit",
			rate:     2.0,
			capacity: 2,
			requests: 3,
			wait:     500 * time.Millisecond,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewLimiter(tt.rate, tt.capacity)

			// Make initial requests
			for i := 0; i < tt.requests-1; i++ {
				assert.True(t, limiter.Allow(), "Request %d should be allowed", i+1)
			}

			// Wait if specified
			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}

			// Test the last request
			got := limiter.Allow()
			assert.Equal(t, tt.want, got, "Allow() = %v, want %v", got, tt.want)
		})
	}
}

func TestLimiter_Concurrent(t *testing.T) {
	limiter := NewLimiter(100.0, 100)
	done := make(chan bool)
	allowed := 0

	// Start 200 goroutines trying to get tokens
	for i := 0; i < 200; i++ {
		go func() {
			if limiter.Allow() {
				allowed++
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 200; i++ {
		<-done
	}

	// We should have exactly 100 allowed requests
	assert.Equal(t, 100, allowed, "Expected 100 allowed requests, got %d", allowed)
}

func TestLimiter_Refill(t *testing.T) {
	limiter := NewLimiter(10.0, 10)

	// Use all tokens
	for i := 0; i < 10; i++ {
		require.True(t, limiter.Allow(), "Request %d should be allowed", i+1)
	}

	// Should be rate limited
	require.False(t, limiter.Allow(), "Should be rate limited")

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)

	// Should be able to make requests again
	assert.True(t, limiter.Allow(), "Should be able to make requests after refill")
}
