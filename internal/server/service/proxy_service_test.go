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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("test response"))
		require.NoError(t, err, "Failed to write test response")
	}))
	defer ts.Close()

	service := NewProxyService(ts.URL, time.Second)

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Forward the request
	resp, err := service.ForwardRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "test response", string(resp))
}

func TestProxyService_ForwardRequest_Error(t *testing.T) {
	// Create a test server that returns an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("test response"))
		require.NoError(t, err, "Failed to write test response")
	}))
	defer ts.Close()

	service := NewProxyService(ts.URL, time.Second)

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	_, err = service.ForwardRequest(context.Background(), req)
	assert.Error(t, err)
}

func TestProxyService_Timeout(t *testing.T) {
	// Create a test server that sleeps
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		_, err := w.Write([]byte("test response"))
		require.NoError(t, err, "Failed to write test response")
	}))
	defer ts.Close()

	service := NewProxyService(ts.URL, 100*time.Millisecond)

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	_, err = service.ForwardRequest(context.Background(), req)
	assert.Error(t, err)
}
