package service

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTimeoutConn is a mock connection that always returns timeout errors
type mockTimeoutConn struct {
	readCalls int
	data      string
}

func (m *mockTimeoutConn) Read(b []byte) (n int, err error) {
	m.readCalls++
	if m.readCalls == 1 {
		return 0, &net.OpError{
			Op:     "read",
			Net:    "tcp",
			Source: nil,
			Addr:   nil,
			Err:    &timeoutError{},
		}
	}
	copy(b, m.data)
	return len(m.data), nil
}

func (m *mockTimeoutConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m *mockTimeoutConn) Close() error {
	return nil
}

func (m *mockTimeoutConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

func (m *mockTimeoutConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

func (m *mockTimeoutConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockTimeoutConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockTimeoutConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestConnectionManager_New(t *testing.T) {
	manager := NewConnectionManager(time.Second, time.Second)
	assert.Equal(t, time.Second, manager.readTimeout, "Expected readTimeout %v", time.Second)
	assert.Equal(t, time.Second, manager.writeTimeout, "Expected writeTimeout %v", time.Second)
	assert.NotNil(t, manager.pool, "Expected non-nil pool")
}

func TestConnectionManager_SetTimeouts(t *testing.T) {
	manager := NewConnectionManager(time.Second, time.Second)

	// Create a test connection
	server, client := net.Pipe()
	defer func() {
		require.NoError(t, server.Close())
		require.NoError(t, client.Close())
	}()

	err := manager.SetTimeouts(client)
	assert.NoError(t, err)
}

func TestConnectionManager_Write(t *testing.T) {
	manager := NewConnectionManager(time.Second, time.Second)

	// Create a test connection
	server, client := net.Pipe()
	defer func() {
		require.NoError(t, server.Close())
		require.NoError(t, client.Close())
	}()

	// Use WaitGroup to synchronize read and write operations
	var wg sync.WaitGroup
	wg.Add(1)

	// Start reader goroutine
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		n, err := server.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, "test message\n", string(buf[:n]))
	}()

	// Test writing data
	err := manager.Write(client, "test message")
	assert.NoError(t, err)

	// Wait for read to complete
	wg.Wait()
}

func TestConnectionManager_ReadWithRetry(t *testing.T) {
	t.Run("successful read after timeout", func(t *testing.T) {
		mockConn := &mockTimeoutConn{data: "test data"}
		manager := NewConnectionManager(100*time.Millisecond, 100*time.Millisecond)

		result, err := manager.ReadWithRetry(mockConn)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "test data" {
			t.Errorf("expected 'test data', got '%s'", result)
		}
		if mockConn.readCalls != 2 {
			t.Errorf("expected 2 read calls, got %d", mockConn.readCalls)
		}
	})
}
