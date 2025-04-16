package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/pow"
)

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

type defaultDialer struct{}

func (d *defaultDialer) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

type Handler struct {
	config *config.ClientConfig
	dialer Dialer
}

func NewHandler(config *config.ClientConfig) *Handler {
	return &Handler{
		config: config,
		dialer: &defaultDialer{},
	}
}

func (h *Handler) MakeRequest() error {
	addr := net.JoinHostPort(h.config.Client.ServerHost, h.config.Client.ServerPort)
	conn, err := h.dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	readTimeout, err := h.config.GetReadTimeout()
	if err != nil {
		return fmt.Errorf("invalid read timeout: %v", err)
	}
	writeTimeout, err := h.config.GetWriteTimeout()
	if err != nil {
		return fmt.Errorf("invalid write timeout: %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %v", err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %v", err)
	}

	reader := bufio.NewReader(conn)
	challenge, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read challenge: %v", err)
	}

	challenge = strings.TrimSpace(challenge)
	nonce, err := pow.SolvePoW(challenge)
	if err != nil {
		return fmt.Errorf("failed to solve PoW: %v", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %v", err)
	}
	if _, err := fmt.Fprintf(conn, "%d\n", nonce); err != nil {
		return fmt.Errorf("failed to write nonce: %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %v", err)
	}
	quote, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read quote: %v", err)
	}

	log.Printf("Received quote: %s", strings.TrimSpace(quote))
	return nil
}
