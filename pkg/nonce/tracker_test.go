package nonce

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTracker_IsValid(t *testing.T) {
	tracker := NewTracker(1 * time.Second)

	// First use of nonce should be valid
	assert.True(t, tracker.IsValid(123))

	// Reuse of same nonce should be invalid
	assert.False(t, tracker.IsValid(123))

	// Different nonce should be valid
	assert.True(t, tracker.IsValid(456))
}

func TestTracker_Window(t *testing.T) {
	tracker := NewTracker(100 * time.Millisecond)

	// Nonce should be valid initially
	assert.True(t, tracker.IsValid(123))

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Same nonce should be valid again after window expires
	assert.True(t, tracker.IsValid(123))
}

func TestTracker_Concurrent(t *testing.T) {
	tracker := NewTracker(1 * time.Second)
	done := make(chan bool)
	valid := 0

	// Try to use the same nonce concurrently
	for i := 0; i < 10; i++ {
		go func() {
			if tracker.IsValid(123) {
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

	// Add some nonces
	tracker.IsValid(1)
	tracker.IsValid(2)

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Add a new nonce to trigger cleanup
	tracker.IsValid(3)

	// Old nonces should be reusable
	assert.True(t, tracker.IsValid(1))
	assert.True(t, tracker.IsValid(2))
}
