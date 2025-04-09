package client

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConn struct {
	*bytes.Buffer
	closed bool
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1234}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4321}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type mockDialer struct {
	conn net.Conn
}

func (d *mockDialer) Dial(network, address string) (net.Conn, error) {
	return d.conn, nil
}

func TestMakeRequest(t *testing.T) {
	cfg := &config.ClientConfig{
		ServerHost: "localhost",
		ServerPort: "8080",
	}
	handler := NewHandler(cfg)

	conn := &mockConn{Buffer: &bytes.Buffer{}}

	timestamp := time.Now().Unix()
	challengeStr := fmt.Sprintf("test_data|0000|%d\n", timestamp)
	_, err := conn.Write([]byte(challengeStr))
	require.NoError(t, err, "Failed to write challenge")

	handler.dialer = &mockDialer{conn: conn}

	err = handler.MakeRequest()
	require.NoError(t, err, "Failed to make request")

	_, err = conn.Write([]byte("test quote\n"))
	require.NoError(t, err, "Failed to write quote")

	assert.True(t, conn.closed, "Expected connection to be closed")
}
