package service

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/pow"
)

type RequestHandler struct {
	cfg          *config.ServerConfig
	quoteService *QuoteService
	proxyService *ProxyService
	powService   *PoWService
	challenges   map[string]pow.Challenge // Store challenges for each client
}

func NewRequestHandler(cfg *config.ServerConfig) (*RequestHandler, error) {
	h := &RequestHandler{
		cfg:        cfg,
		challenges: make(map[string]pow.Challenge),
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

	powService, err := NewPoWService(cfg.Server.ChallengeDifficulty, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	h.powService = powService

	return h, nil
}

func (h *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID := r.RemoteAddr

	switch r.Method {
	case http.MethodGet:
		// Generate and send challenge
		challenge := h.powService.GenerateChallenge()
		h.challenges[clientID] = challenge
		challengeStr := h.powService.FormatChallenge(challenge)

		if _, err := w.Write([]byte(challengeStr)); err != nil {
			log.Printf("Error writing challenge: %v", err)
		}
	case http.MethodPost:
		// Verify nonce and send quote
		nonceStr, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read nonce", http.StatusBadRequest)
			return
		}

		nonce, err := strconv.ParseInt(strings.TrimSpace(string(nonceStr)), 10, 64)
		if err != nil {
			http.Error(w, "Invalid nonce format", http.StatusBadRequest)
			return
		}

		// Get the challenge for this client
		challenge, ok := h.challenges[clientID]
		if !ok {
			http.Error(w, "No challenge found", http.StatusBadRequest)
			return
		}

		if !h.powService.VerifyPoW(challenge, nonce) {
			http.Error(w, "Invalid proof of work", http.StatusForbidden)
			return
		}

		// Clear the challenge after verification
		delete(h.challenges, clientID)

		if h.cfg.Server.Mode == "quotes" {
			quote := h.quoteService.GetRandomQuote()
			if _, err := w.Write([]byte(quote)); err != nil {
				log.Printf("Error writing quote response: %v", err)
			}
		} else {
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
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
