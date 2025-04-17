package service

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"
)

type ProxyService struct {
	targetURL string
	timeout   time.Duration
	client    *http.Client
}

func NewProxyService(targetURL string, timeout time.Duration) *ProxyService {
	return &ProxyService{
		targetURL: targetURL,
		timeout:   timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s *ProxyService) ForwardRequest(ctx context.Context, r *http.Request) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, r.Method, s.targetURL, r.Body)
	if err != nil {
		return nil, err
	}

	// Copy headers
	for k, v := range r.Header {
		req.Header[k] = v
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log the error but don't fail the request
			log.Printf("Error closing response body: %v", err)
		}
	}()

	return io.ReadAll(resp.Body)
}
