package server

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/pow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.readBuf.Len() == 0 {
		return 0, nil
	}
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
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

func TestHandleConnection(t *testing.T) {
	cfg := &config.ServerConfig{
		ChallengeDifficulty: "0000",
		RateLimit:           1000,
		BurstLimit:          1000,
	}
	handler := NewHandler(cfg)

	conn := &mockConn{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}

	go handler.HandleConnection(conn)

	time.Sleep(100 * time.Millisecond)

	challengeStr, err := conn.writeBuf.ReadString('\n')
	require.NoError(t, err, "Failed to read challenge")
	challengeStr = strings.TrimSpace(challengeStr)

	parts := strings.Split(challengeStr, "|")
	require.Len(t, parts, 3, "Invalid challenge format")

	nonce, err := pow.SolvePoW(challengeStr)
	require.NoError(t, err, "Failed to solve PoW")

	conn.readBuf.Write([]byte(fmt.Sprintf("%d\n", nonce)))

	time.Sleep(200 * time.Millisecond)

	quote, err := conn.writeBuf.ReadString('\n')
	require.NoError(t, err, "Failed to read quote")
	quote = strings.TrimSpace(quote)

	assert.NotEmpty(t, quote, "Expected non-empty quote")
	assert.NotEqual(t, "Invalid proof of work", quote, "Got invalid proof of work response")
	assert.True(t, conn.closed, "Expected connection to be closed")
}
