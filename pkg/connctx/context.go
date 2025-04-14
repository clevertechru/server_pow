package connctx

import (
	"context"
	"net"
	"time"
)

type contextKey string

const (
	connIDKey     = contextKey("conn_id")
	startTimeKey  = contextKey("start_time")
	remoteAddrKey = contextKey("remote_addr")
)

type ConnContext struct {
	ctx        context.Context
	cancel     context.CancelFunc
	conn       net.Conn
	timeoutCtx context.Context
}

func NewConnContext(parent context.Context, conn net.Conn, timeout time.Duration) *ConnContext {
	ctx, cancel := context.WithCancel(parent)
	ctx = context.WithValue(ctx, connIDKey, generateConnID())
	ctx = context.WithValue(ctx, startTimeKey, time.Now())
	ctx = context.WithValue(ctx, remoteAddrKey, conn.RemoteAddr().String())

	timeoutCtx, _ := context.WithTimeout(ctx, timeout)

	return &ConnContext{
		ctx:        ctx,
		cancel:     cancel,
		conn:       conn,
		timeoutCtx: timeoutCtx,
	}
}

func (c *ConnContext) Context() context.Context {
	return c.ctx
}

func (c *ConnContext) TimeoutContext() context.Context {
	return c.timeoutCtx
}

func (c *ConnContext) Cancel() {
	c.cancel()
}

func (c *ConnContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *ConnContext) Conn() net.Conn {
	return c.conn
}

func (c *ConnContext) ID() string {
	return c.ctx.Value(connIDKey).(string)
}

func (c *ConnContext) StartTime() time.Time {
	return c.ctx.Value(startTimeKey).(time.Time)
}

func (c *ConnContext) RemoteAddr() string {
	return c.ctx.Value(remoteAddrKey).(string)
}

func (c *ConnContext) Duration() time.Duration {
	return time.Since(c.StartTime())
}

func (c *ConnContext) Read(b []byte) (n int, err error) {
	done := make(chan struct{})
	var result int
	var readErr error

	go func() {
		result, readErr = c.conn.Read(b)
		close(done)
	}()

	select {
	case <-done:
		return result, readErr
	case <-c.timeoutCtx.Done():
		return 0, context.DeadlineExceeded
	case <-c.ctx.Done():
		return 0, context.Canceled
	}
}

func (c *ConnContext) Write(b []byte) (n int, err error) {
	done := make(chan struct{})
	var result int
	var writeErr error

	go func() {
		result, writeErr = c.conn.Write(b)
		close(done)
	}()

	select {
	case <-done:
		return result, writeErr
	case <-c.timeoutCtx.Done():
		return 0, context.DeadlineExceeded
	case <-c.ctx.Done():
		return 0, context.Canceled
	}
}

func (c *ConnContext) Close() error {
	c.cancel()
	return c.conn.Close()
}

// Helper functions
func generateConnID() string {
	return time.Now().Format("20060102150405.000") + "-" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
