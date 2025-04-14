package connctx

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConn struct {
	readDelay  time.Duration
	writeDelay time.Duration
	readErr    error
	writeErr   error
	closed     bool
	addr       net.Addr
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	time.Sleep(m.readDelay)
	if m.readErr != nil {
		return 0, m.readErr
	}
	return len(b), nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	time.Sleep(m.writeDelay)
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return len(b), nil
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return m.addr }
func (m *mockConn) RemoteAddr() net.Addr               { return m.addr }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

type mockAddr struct{}

func (m *mockAddr) Network() string { return "tcp" }
func (m *mockAddr) String() string  { return "127.0.0.1:8080" }

func TestNewConnContext(t *testing.T) {
	conn := &mockConn{addr: &mockAddr{}}
	timeout := 100 * time.Millisecond
	ctx := context.Background()

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	// Check basic properties
	assert.NotEmpty(t, connCtx.ID())
	assert.Equal(t, conn.RemoteAddr().String(), connCtx.RemoteAddr())
	assert.NotNil(t, connCtx.StartTime())
	assert.NotNil(t, connCtx.Context())
	assert.NotNil(t, connCtx.TimeoutContext())
}

func TestConnContext_ReadTimeout(t *testing.T) {
	conn := &mockConn{
		addr:      &mockAddr{},
		readDelay: 200 * time.Millisecond,
	}
	timeout := 100 * time.Millisecond
	ctx := context.Background()

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	buf := make([]byte, 10)
	_, err := connCtx.Read(buf)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestConnContext_WriteTimeout(t *testing.T) {
	conn := &mockConn{
		addr:       &mockAddr{},
		writeDelay: 200 * time.Millisecond,
	}
	timeout := 100 * time.Millisecond
	ctx := context.Background()

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	_, err := connCtx.Write([]byte("test"))
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestConnContext_Cancel(t *testing.T) {
	conn := &mockConn{
		addr:      &mockAddr{},
		readDelay: 200 * time.Millisecond,
	}
	timeout := 1 * time.Second
	ctx := context.Background()

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	go func() {
		time.Sleep(100 * time.Millisecond)
		connCtx.Cancel()
	}()

	buf := make([]byte, 10)
	_, err := connCtx.Read(buf)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestConnContext_ReadError(t *testing.T) {
	expectedErr := errors.New("read error")
	conn := &mockConn{
		addr:    &mockAddr{},
		readErr: expectedErr,
	}
	timeout := 1 * time.Second
	ctx := context.Background()

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	buf := make([]byte, 10)
	_, err := connCtx.Read(buf)
	assert.ErrorIs(t, err, expectedErr)
}

func TestConnContext_WriteError(t *testing.T) {
	expectedErr := errors.New("write error")
	conn := &mockConn{
		addr:     &mockAddr{},
		writeErr: expectedErr,
	}
	timeout := 1 * time.Second
	ctx := context.Background()

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	_, err := connCtx.Write([]byte("test"))
	assert.ErrorIs(t, err, expectedErr)
}

func TestConnContext_Close(t *testing.T) {
	conn := &mockConn{addr: &mockAddr{}}
	timeout := 1 * time.Second
	ctx := context.Background()

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	err := connCtx.Close()
	assert.NoError(t, err)
	assert.True(t, conn.closed)

	// Check context is canceled
	select {
	case <-connCtx.Done():
		// Expected
	default:
		t.Error("Context should be canceled after Close")
	}
}

func TestConnContext_Duration(t *testing.T) {
	conn := &mockConn{addr: &mockAddr{}}
	timeout := 1 * time.Second
	ctx := context.Background()

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	time.Sleep(100 * time.Millisecond)
	duration := connCtx.Duration()
	assert.True(t, duration >= 100*time.Millisecond)
}

func TestConnContext_ParentCancellation(t *testing.T) {
	conn := &mockConn{
		addr:      &mockAddr{},
		readDelay: 200 * time.Millisecond,
	}
	timeout := 1 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	connCtx := NewConnContext(ctx, conn, timeout)
	require.NotNil(t, connCtx)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	buf := make([]byte, 10)
	_, err := connCtx.Read(buf)
	assert.ErrorIs(t, err, context.Canceled)
}
