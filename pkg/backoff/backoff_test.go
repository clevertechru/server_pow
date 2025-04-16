package backoff

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConn struct {
	addr   string
	closed bool
	mu     sync.Mutex
}

func (m *mockConn) Read(b []byte) (n int, err error)  { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error) { return 0, nil }
func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
func (m *mockConn) LocalAddr() net.Addr { return nil }
func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
}
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestQueue(t *testing.T) {
	t.Run("Basic Add and Get", func(t *testing.T) {
		queue := NewQueue(5, time.Millisecond, time.Second)
		conn := &mockConn{}

		// Add connection
		assert.True(t, queue.Add(conn), "Should accept connection")
		assert.Equal(t, 1, queue.Size(), "Queue size should be 1")

		// Get connection immediately (no backoff for first attempt)
		retrieved := queue.Get()
		assert.NotNil(t, retrieved, "Should get connection back")
		assert.Equal(t, 0, queue.Size(), "Queue should be empty")
	})

	t.Run("Queue Full", func(t *testing.T) {
		queue := NewQueue(2, time.Millisecond, time.Second)

		// Fill queue
		conn1, conn2 := &mockConn{}, &mockConn{}
		assert.True(t, queue.Add(conn1), "Should accept first connection")
		assert.True(t, queue.Add(conn2), "Should accept second connection")

		// Try to add when full
		conn3 := &mockConn{}
		assert.False(t, queue.Add(conn3), "Should reject when full")
		assert.Equal(t, 2, queue.Size(), "Queue size should be 2")
	})

	t.Run("Backoff Timing", func(t *testing.T) {
		baseDelay := 10 * time.Millisecond
		queue := NewQueue(5, baseDelay, time.Second)
		conn := &mockConn{}

		// First attempt - no delay
		require.True(t, queue.Add(conn))
		retrieved := queue.Get()
		require.NotNil(t, retrieved)

		// Second attempt - should have baseDelay
		require.True(t, queue.Add(conn))
		time.Sleep(baseDelay / 2) // Not enough time
		retrieved = queue.Get()
		require.Nil(t, retrieved, "Should not get connection before backoff")

		time.Sleep(baseDelay) // Now enough time
		retrieved = queue.Get()
		require.NotNil(t, retrieved, "Should get connection after backoff")

		// Third attempt - should have 2*baseDelay
		require.True(t, queue.Add(conn))
		time.Sleep(baseDelay) // Not enough time
		retrieved = queue.Get()
		require.Nil(t, retrieved, "Should not get connection before backoff")

		time.Sleep(baseDelay * 2) // Now enough time
		retrieved = queue.Get()
		require.NotNil(t, retrieved, "Should get connection after backoff")
	})

	t.Run("Max Backoff", func(t *testing.T) {
		baseDelay := time.Millisecond
		maxDelay := 4 * time.Millisecond
		queue := NewQueue(5, baseDelay, maxDelay)
		conn := &mockConn{}

		// Add and retrieve multiple times to increase backoff
		for i := 0; i < 5; i++ {
			require.True(t, queue.Add(conn))
			time.Sleep(maxDelay + time.Millisecond) // Wait max delay plus buffer
			retrieved := queue.Get()
			require.NotNil(t, retrieved, "Should get connection after max backoff")
		}
	})

	t.Run("Clear Queue", func(t *testing.T) {
		queue := NewQueue(5, time.Millisecond, time.Second)

		// Add connections
		conn1, conn2 := &mockConn{}, &mockConn{}
		require.True(t, queue.Add(conn1))
		require.True(t, queue.Add(conn2))
		require.Equal(t, 2, queue.Size())

		// Clear queue
		queue.Clear()
		assert.Equal(t, 0, queue.Size(), "Queue should be empty after clear")
		assert.True(t, conn1.closed, "Connection should be closed")
		assert.True(t, conn2.closed, "Connection should be closed")
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		queue := NewQueue(100, time.Millisecond, time.Second)
		var wg sync.WaitGroup

		// Concurrent adds
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				conn := &mockConn{}
				queue.Add(conn)
			}()
		}

		// Concurrent gets
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				queue.Get()
			}()
		}

		wg.Wait()
		assert.True(t, queue.Size() >= 0, "Queue size should not be negative")
	})

	t.Run("Multiple Connections Same Address", func(t *testing.T) {
		queue := NewQueue(5, time.Millisecond, time.Second)
		conn := &mockConn{}

		// First attempt - no delay
		assert.True(t, queue.Add(conn))
		retrieved := queue.Get()
		require.NotNil(t, retrieved)

		// Second attempt - should have backoff
		assert.True(t, queue.Add(conn))
		retrieved = queue.Get()
		require.Nil(t, retrieved, "Should not get connection before backoff")

		time.Sleep(2 * time.Millisecond) // Wait for backoff
		retrieved = queue.Get()
		require.NotNil(t, retrieved, "Should get connection after backoff")
	})
}

func TestCalculateBackoff(t *testing.T) {
	queue := NewQueue(5, time.Millisecond, 10*time.Millisecond)

	tests := []struct {
		name     string
		attempts int
		expected time.Duration
	}{
		{"No attempts", 0, 0},
		{"First attempt", 1, time.Millisecond},
		{"Second attempt", 2, 2 * time.Millisecond},
		{"Third attempt", 3, 4 * time.Millisecond},
		{"Fourth attempt", 4, 8 * time.Millisecond},
		{"Max backoff", 5, 10 * time.Millisecond},
		{"Beyond max backoff", 6, 10 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := queue.calculateBackoff(tt.attempts)
			assert.Equal(t, tt.expected, backoff)
		})
	}
}
