package service

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type ConnectionManager struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	pool         *sync.Pool
}

func NewConnectionManager(readTimeout, writeTimeout time.Duration) *ConnectionManager {
	return &ConnectionManager{
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		pool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 1024)
				return &b
			},
		},
	}
}

func (m *ConnectionManager) SetTimeouts(conn net.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(m.readTimeout)); err != nil {
		return fmt.Errorf("error setting read deadline: %w", err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(m.writeTimeout)); err != nil {
		return fmt.Errorf("error setting write deadline: %w", err)
	}
	return nil
}

func (m *ConnectionManager) ReadWithRetry(conn net.Conn) (string, error) {
	bufferPtr := m.pool.Get().(*[]byte)
	buffer := *bufferPtr
	defer func() {
		if bufferPtr != nil {
			m.pool.Put(bufferPtr)
		}
	}()

	for {
		if err := m.SetTimeouts(conn); err != nil {
			return "", err
		}

		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Read timeout: %v", err)
				continue
			}
			return "", fmt.Errorf("error reading: %w", err)
		}

		if n == 0 {
			continue
		}

		return strings.TrimSpace(string(buffer[:n])), nil
	}
}

func (m *ConnectionManager) Write(conn net.Conn, data string) error {
	if err := m.SetTimeouts(conn); err != nil {
		return err
	}

	if _, err := conn.Write([]byte(data + "\n")); err != nil {
		return fmt.Errorf("error writing: %w", err)
	}
	return nil
}
