package service

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"
)

type HTTPHandler struct {
	cfg          *config.ServerConfig
	powService   *PoWService
	quoteService *QuoteService
	proxyService *ProxyService
}

func NewHTTPHandler(cfg *config.ServerConfig) (*HTTPHandler, error) {
	powService, err := NewPoWService(cfg.Server.ChallengeDifficulty, 5*time.Minute)
	if err != nil {
		return nil, err
	}

	h := &HTTPHandler{
		cfg:        cfg,
		powService: powService,
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

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get challenge from header or generate new one
	challenge := r.Header.Get("X-Challenge")
	if challenge == "" {
		challenge = h.powService.FormatChallenge(h.powService.GenerateChallenge())
		w.Header().Set("X-Challenge", challenge)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get nonce from header
	nonceStr := r.Header.Get("X-Nonce")
	if nonceStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse challenge and nonce
	parts := strings.Split(challenge, "|")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	challengeTimestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nonce, err := strconv.ParseInt(nonceStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse and verify PoW
	powChallenge, err := h.powService.ParseChallenge(challenge)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !h.powService.VerifyPoW(powChallenge, nonce) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check challenge expiration
	if time.Now().Unix()-challengeTimestamp > 300 { // 5 minutes
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Handle request based on mode
	if h.cfg.Server.Mode == "quotes" {
		quote := h.quoteService.GetRandomQuote()
		if _, err := w.Write([]byte(quote)); err != nil {
			log.Printf("Error writing quote response: %v", err)
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
		log.Printf("Error writing proxy response: %v", err)
	}
}
