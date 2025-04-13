package nonce

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTracker_IsValid(t *testing.T) {
	tracker := NewTracker(1 * time.Second)
	now := time.Now().Unix()

	// First use of nonce should be valid
	assert.True(t, tracker.IsValid(123, now))

	// Reuse of same nonce should be invalid
	assert.False(t, tracker.IsValid(123, now))

	// Different nonce should be valid
	assert.True(t, tracker.IsValid(456, now))
}

func TestTracker_Window(t *testing.T) {
	tracker := NewTracker(100 * time.Millisecond)
	now := time.Now().Unix()

	// Nonce should be valid initially
	assert.True(t, tracker.IsValid(123, now))

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Same nonce should be valid again after window expires
	assert.True(t, tracker.IsValid(123, now))
}

func TestTracker_Concurrent(t *testing.T) {
	tracker := NewTracker(1 * time.Second)
	now := time.Now().Unix()
	done := make(chan bool)
	valid := 0

	// Try to use the same nonce concurrently
	for i := 0; i < 10; i++ {
		go func() {
			if tracker.IsValid(123, now) {
				valid++
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Only one should be valid
	assert.Equal(t, 1, valid, "Expected exactly one valid nonce")
}

func TestTracker_Cleanup(t *testing.T) {
	tracker := NewTracker(100 * time.Millisecond)
	now := time.Now().Unix()

	// Add some nonces
	tracker.IsValid(1, now)
	tracker.IsValid(2, now)

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Add a new nonce to trigger cleanup
	tracker.IsValid(3, now)

	// Old nonces should be reusable
	assert.True(t, tracker.IsValid(1, now))
	assert.True(t, tracker.IsValid(2, now))
}
