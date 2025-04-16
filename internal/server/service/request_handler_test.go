package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestHandler_HandleRequest_QuotesMode(t *testing.T) {
	cfg := &config.ServerConfig{}
	cfg.Server.Mode = "quotes"

	handler, err := NewRequestHandler(cfg)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
}

func TestRequestHandler_HandleRequest_ProxyMode(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("proxy response"))
		require.NoError(t, err, "Failed to write test response")
	}))
	defer ts.Close()

	cfg := &config.ServerConfig{}
	cfg.Server.Mode = "proxy"
	cfg.Server.Proxy.Target = ts.URL
	cfg.Server.Proxy.Timeout = "5s"

	handler, err := NewRequestHandler(cfg)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "proxy response", w.Body.String())
}
