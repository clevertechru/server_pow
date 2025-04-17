package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyService_ForwardRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("test response")); err != nil {
			t.Errorf("Error writing test response: %v", err)
		}
	}))
	defer server.Close()

	service := NewProxyService(server.URL, time.Second)

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Forward the request
	resp, err := service.ForwardRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "test response", string(resp))
}

func TestProxyService_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		if _, err := w.Write([]byte("test response")); err != nil {
			t.Errorf("Error writing test response: %v", err)
		}
	}))
	defer server.Close()

	service := NewProxyService(server.URL, 100*time.Millisecond)

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	_, err = service.ForwardRequest(context.Background(), req)
	assert.Error(t, err)
}
