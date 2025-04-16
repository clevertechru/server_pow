package service

import (
	"net/http"

	"github.com/clevertechru/server_pow/pkg/config"
)

type RequestHandler struct {
	cfg          *config.ServerConfig
	quoteService *QuoteService
	proxyService *ProxyService
}

func NewRequestHandler(cfg *config.ServerConfig) (*RequestHandler, error) {
	h := &RequestHandler{
		cfg: cfg,
	}

	if cfg.Server.Mode == "quotes" {
		h.quoteService = NewQuoteService()
	} else {
		timeout, err := cfg.GetProxyTimeout()
		if err != nil {
			return nil, err
		}
		h.proxyService = NewProxyService(cfg.Server.Proxy.Target, timeout)
	}

	return h, nil
}

func (h *RequestHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if h.cfg.Server.Mode == "quotes" {
		quote := h.quoteService.GetRandomQuote()
		if _, err := w.Write([]byte(quote)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		return
	}

	// Proxy mode
	resp, err := h.proxyService.ForwardRequest(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(resp); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
