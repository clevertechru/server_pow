package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/pow"
)

type Handler struct {
	config *config.ClientConfig
	client *http.Client
}

func NewHandler(config *config.ClientConfig) *Handler {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Client.TLS.InsecureSkipVerify,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &Handler{
		config: config,
		client: client,
	}
}

func (h *Handler) MakeRequest() error {
	protocol := h.config.Client.Protocol
	if protocol == "" {
		protocol = "http"
	}

	url := fmt.Sprintf("%s://%s:%s", protocol, h.config.Client.ServerHost, h.config.Client.ServerPort)

	// First request to get challenge
	resp, err := h.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get challenge: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	challengeBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read challenge: %v", err)
	}

	challenge := strings.TrimSpace(string(challengeBytes))
	nonce, err := pow.SolvePoW(challenge)
	if err != nil {
		return fmt.Errorf("failed to solve PoW: %v", err)
	}

	// Second request to send nonce and get quote
	req, err := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf("%d", nonce)))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err = h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send nonce: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	quoteBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read quote: %v", err)
	}

	log.Printf("Received quote: %s", strings.TrimSpace(string(quoteBytes)))
	return nil
}
