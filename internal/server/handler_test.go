package server

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/pow"
	"github.com/clevertechru/server_pow/pkg/quotes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
	// Create a temporary quotes file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "quotes.yml")
	quotesList := []string{
		"Test quote 1",
		"Test quote 2",
		"Test quote 3",
	}

	data := struct {
		Quotes []string `yaml:"quotes"`
	}{
		Quotes: quotesList,
	}

	file, err := os.Create(configPath)
	require.NoError(t, err, "Failed to create test quotes file")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close test quotes file: %v", err)
		}
	}()

	if err := yaml.NewEncoder(file).Encode(data); err != nil {
		t.Fatalf("Failed to write test quotes: %v", err)
	}

	// Initialize quotes storage
	require.NoError(t, quotes.Init(configPath), "Failed to initialize quotes storage")

	cfg := &config.ServerSettings{
		ChallengeDifficulty: 2, // 2 bytes = 16 leading zeros
		RateLimit:           1000,
		BurstLimit:          1000,
		MaxConnections:      1000,
		WorkerPoolSize:      10,
	}
	handler, err := NewHandler(cfg)
	require.NoError(t, err, "Failed to create handler")

	// Test invalid nonce first
	conn := &mockConn{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}

	handler.ProcessConnection(conn)

	time.Sleep(100 * time.Millisecond)

	challengeStr, err := conn.writeBuf.ReadString('\n')
	require.NoError(t, err, "Failed to read challenge")
	challengeStr = strings.TrimSpace(challengeStr)

	parts := strings.Split(challengeStr, "|")
	require.Len(t, parts, 3, "Invalid challenge format")

	// Test invalid nonce
	buf := bytes.NewBuffer(nil)
	if _, err := fmt.Fprintf(buf, "0\n"); err != nil {
		t.Fatalf("Error writing nonce: %v", err)
	}
	if _, err := conn.readBuf.Write(buf.Bytes()); err != nil {
		t.Fatalf("Error writing nonce: %v", err)
	}

	// Wait for the nonce to be processed
	time.Sleep(100 * time.Millisecond)

	response, err := conn.writeBuf.ReadString('\n')
	require.NoError(t, err, "Failed to read response")
	response = strings.TrimSpace(response)
	assert.Equal(t, "Invalid proof of work", response, "Expected invalid PoW response")
	assert.True(t, conn.closed, "Expected connection to be closed after invalid PoW")

	// Test valid nonce with new connection
	conn = &mockConn{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}

	handler.ProcessConnection(conn)

	time.Sleep(100 * time.Millisecond)

	challengeStr, err = conn.writeBuf.ReadString('\n')
	require.NoError(t, err, "Failed to read challenge")
	challengeStr = strings.TrimSpace(challengeStr)

	parts = strings.Split(challengeStr, "|")
	require.Len(t, parts, 3, "Invalid challenge format")

	nonce, err := pow.SolvePoW(challengeStr)
	require.NoError(t, err, "Failed to solve PoW")

	buf = bytes.NewBuffer(nil)
	if _, err := fmt.Fprintf(buf, "%d\n", nonce); err != nil {
		t.Fatalf("Error writing nonce: %v", err)
	}
	if _, err := conn.readBuf.Write(buf.Bytes()); err != nil {
		t.Fatalf("Error writing nonce: %v", err)
	}

	// Wait for the nonce to be processed
	time.Sleep(100 * time.Millisecond)

	quote, err := conn.writeBuf.ReadString('\n')
	require.NoError(t, err, "Failed to read quote")
	quote = strings.TrimSpace(quote)

	assert.NotEmpty(t, quote, "Expected non-empty quote")
	assert.NotEqual(t, "Invalid proof of work", quote, "Got invalid proof of work response")
	assert.True(t, conn.closed, "Expected connection to be closed")
}
