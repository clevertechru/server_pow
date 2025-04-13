package connlimit

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiter_Acquire(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		acquires int
		want     bool
	}{
		{
			name:     "single acquire allowed",
			limit:    1,
			acquires: 1,
			want:     true,
		},
		{
			name:     "multiple acquires allowed",
			limit:    5,
			acquires: 5,
			want:     true,
		},
		{
			name:     "exceed limit",
			limit:    3,
			acquires: 4,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewLimiter(tt.limit)

			// Make initial acquires
			for i := 0; i < tt.acquires-1; i++ {
				assert.True(t, limiter.Acquire(), "Acquire %d should succeed", i+1)
			}

			// Test the last acquire
			got := limiter.Acquire()
			assert.Equal(t, tt.want, got, "Acquire() = %v, want %v", got, tt.want)
		})
	}
}

func TestLimiter_Release(t *testing.T) {
	limiter := NewLimiter(2)

	// Acquire both slots
	assert.True(t, limiter.Acquire())
	assert.True(t, limiter.Acquire())

	// Should be at limit
	assert.False(t, limiter.Acquire())

	// Release one slot
	limiter.Release()

	// Should be able to acquire again
	assert.True(t, limiter.Acquire())
}

func TestLimiter_Concurrent(t *testing.T) {
	limiter := NewLimiter(10)
	var wg sync.WaitGroup
	active := 0
	mu := sync.Mutex{}

	// Start 20 goroutines trying to acquire
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Acquire() {
				mu.Lock()
				active++
				mu.Unlock()

				// Simulate some work
				time.Sleep(10 * time.Millisecond)

				mu.Lock()
				active--
				mu.Unlock()
				limiter.Release()
			}
		}()
	}

	// Check active connections periodically
	for i := 0; i < 5; i++ {
		time.Sleep(20 * time.Millisecond)
		mu.Lock()
		assert.LessOrEqual(t, active, 10, "Active connections should never exceed limit")
		mu.Unlock()
	}

	wg.Wait()
	assert.Equal(t, 0, active, "All connections should be released")
}

func TestLimiter_Count(t *testing.T) {
	limiter := NewLimiter(3)

	assert.Equal(t, 0, limiter.Count(), "Initial count should be 0")

	limiter.Acquire()
	assert.Equal(t, 1, limiter.Count(), "Count should be 1 after acquire")

	limiter.Acquire()
	assert.Equal(t, 2, limiter.Count(), "Count should be 2 after second acquire")

	limiter.Release()
	assert.Equal(t, 1, limiter.Count(), "Count should be 1 after release")
}
