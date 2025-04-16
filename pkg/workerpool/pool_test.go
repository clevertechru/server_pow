package workerpool

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockConn struct {
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
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestPool(t *testing.T) {
	t.Run("Basic Functionality", func(t *testing.T) {
		processed := make(chan bool, 1)
		handler := func(conn net.Conn) {
			processed <- true
		}

		pool := NewPool(2, handler)
		defer pool.Shutdown()

		conn := &mockConn{}
		assert.True(t, pool.Submit(conn), "Should accept connection")

		select {
		case <-processed:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Handler was not called")
		}
	})

	t.Run("Multiple Workers", func(t *testing.T) {
		var wg sync.WaitGroup
		processed := make(chan bool, 10)
		handler := func(conn net.Conn) {
			processed <- true
			wg.Done()
		}

		pool := NewPool(5, handler)
		defer pool.Shutdown()

		// Submit more tasks than workers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			conn := &mockConn{}
			if !pool.Submit(conn) {
				wg.Done() // If submission fails, decrement counter
			}
		}

		// Wait for all accepted tasks to complete
		wg.Wait()
		close(processed)

		count := 0
		for range processed {
			count++
		}
		assert.True(t, count > 0, "Some tasks should be processed")
	})

	t.Run("Queue Full", func(t *testing.T) {
		// Create a pool with small buffer
		processed := make(chan bool)
		handler := func(conn net.Conn) {
			<-processed // Block until test is done
		}

		pool := NewPool(1, handler)
		defer func() {
			close(processed) // Unblock handler before shutdown
			pool.Shutdown()
		}()

		// Fill the queue (buffer size = workers)
		conn1 := &mockConn{}
		assert.True(t, pool.Submit(conn1), "Should accept first connection")

		// Try to submit another connection
		conn2 := &mockConn{}
		assert.False(t, pool.Submit(conn2), "Should reject connection when queue is full")
	})

	t.Run("Shutdown", func(t *testing.T) {
		processed := make(chan bool, 1)
		handler := func(conn net.Conn) {
			processed <- true
		}

		pool := NewPool(2, handler)

		// Submit a task
		conn := &mockConn{}
		assert.True(t, pool.Submit(conn), "Should accept connection")

		// Wait for task to be processed
		select {
		case <-processed:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Handler was not called")
		}

		// Shutdown the pool
		pool.Shutdown()

		// Try to submit after shutdown
		assert.False(t, pool.Submit(&mockConn{}), "Should reject connection after shutdown")
	})

	t.Run("Worker Cleanup", func(t *testing.T) {
		handler := func(conn net.Conn) {}

		pool := NewPool(3, handler)
		// Give workers time to start
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(3), pool.activeWorkers, "Should start with 3 workers")

		pool.Shutdown()
		// Give workers time to shut down
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, int32(0), pool.activeWorkers, "All workers should be cleaned up")
	})

	t.Run("Concurrent Submits", func(t *testing.T) {
		var wg sync.WaitGroup
		processed := make(chan bool, 100)
		handler := func(conn net.Conn) {
			processed <- true
		}

		pool := NewPool(10, handler)

		// Submit tasks concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				conn := &mockConn{}
				pool.Submit(conn)
			}()
		}

		wg.Wait()
		pool.Shutdown()
		close(processed)

		count := 0
		for range processed {
			count++
		}
		assert.True(t, count > 0, "At least some tasks should be processed")
	})
}
