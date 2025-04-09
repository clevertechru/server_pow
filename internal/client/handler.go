package client

import (
	"bufio"
	"fmt"
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
	addr := net.JoinHostPort(h.config.ServerHost, h.config.ServerPort)
	conn, err := h.dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))

	reader := bufio.NewReader(conn)
	challenge, err := reader.ReadString('\n')
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("read timeout: %v", err)
		}
		return fmt.Errorf("failed to read challenge: %v", err)
	}
	challenge = strings.TrimSpace(challenge)

	nonce, err := pow.SolvePoW(challenge)
	if err != nil {
		return fmt.Errorf("failed to solve PoW: %v", err)
	}

	// Reset write deadline before sending nonce
	conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))
	_, err = conn.Write([]byte(fmt.Sprintf("%d\n", nonce)))
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("write timeout: %v", err)
		}
		return fmt.Errorf("failed to write nonce: %v", err)
	}

	// Reset read deadline before reading quote
	conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout))
	quote, err := reader.ReadString('\n')
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return fmt.Errorf("read timeout: %v", err)
		}
		return fmt.Errorf("failed to read quote: %v", err)
	}
	fmt.Printf("Received quote: %s", quote)
	return nil
}
