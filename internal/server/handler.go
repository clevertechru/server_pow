package server

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clevertechru/server_pow/internal/server/service"
	"github.com/clevertechru/server_pow/pkg/backoff"
	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/connlimit"
	"github.com/clevertechru/server_pow/pkg/metrics"
	"github.com/clevertechru/server_pow/pkg/ratelimit"
	"github.com/clevertechru/server_pow/pkg/workerpool"
)

type Handler struct {
	config       *config.ServerConfig
	rateLimiter  *ratelimit.Limiter
	connLimiter  *connlimit.Limiter
	workerPool   *workerpool.Pool
	backoffQueue *backoff.Queue
	powService   *service.PoWService
	quoteService *service.QuoteService
	proxyService *service.ProxyService
	connManager  *service.ConnectionManager
}

func NewHandler(config *config.ServerConfig) (*Handler, error) {
	powService, err := service.NewPoWService(config.Server.ChallengeDifficulty, 5*time.Minute)
	if err != nil {
		return nil, err
	}

	readTimeout, err := time.ParseDuration(config.Server.Connection.ReadTimeout)
	if err != nil {
		return nil, err
	}

	writeTimeout, err := time.ParseDuration(config.Server.Connection.WriteTimeout)
	if err != nil {
		return nil, err
	}

	baseBackoff, err := time.ParseDuration(config.Server.Connection.BaseBackoff)
	if err != nil {
		return nil, err
	}

	maxBackoff, err := time.ParseDuration(config.Server.Connection.MaxBackoff)
	if err != nil {
		return nil, err
	}

	rateLimiter := ratelimit.NewLimiter(config.Server.Connection.RateLimit, config.Server.Connection.BurstLimit)
	connLimiter := connlimit.NewLimiter(config.Server.Connection.MaxConnections)

	var quoteService *service.QuoteService
	var proxyService *service.ProxyService

	if config.Server.Mode == "quotes" {
		quoteService = service.NewQuoteService()
	} else {
		proxyTimeout, err := time.ParseDuration(config.Server.Proxy.Timeout)
		if err != nil {
			return nil, err
		}
		proxyService = service.NewProxyService(config.Server.Proxy.Target, proxyTimeout)
	}

	h := &Handler{
		config:       config,
		rateLimiter:  rateLimiter,
		connLimiter:  connLimiter,
		backoffQueue: backoff.NewQueue(config.Server.Connection.QueueSize, baseBackoff, maxBackoff),
		powService:   powService,
		quoteService: quoteService,
		proxyService: proxyService,
		connManager:  service.NewConnectionManager(readTimeout, writeTimeout),
	}

	h.workerPool = workerpool.NewPool(config.Server.Connection.WorkerPoolSize, h.handleConnection)

	// Start queue processor
	go h.processQueue()

	return h, nil
}

func (h *Handler) processQueue() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if conn := h.backoffQueue.Get(); conn != nil {
			if !h.workerPool.Submit(conn) {
				if err := conn.Close(); err != nil {
					log.Printf("Error closing queued connection: %v", err)
				}
			}
		}
	}
}

func (h *Handler) ProcessConnection(conn net.Conn) {
	if !h.workerPool.Submit(conn) {
		if h.backoffQueue.Add(conn) {
			if _, err := conn.Write([]byte("Server is busy, connection queued\n")); err != nil {
				log.Printf("Error writing to connection: %v", err)
				if err := conn.Close(); err != nil {
					log.Printf("Error closing connection: %v", err)
				}
			}
		} else {
			if _, err := conn.Write([]byte("Server is busy, please try again later\n")); err != nil {
				log.Printf("Error writing to connection: %v", err)
			}
			if err := conn.Close(); err != nil {
				log.Printf("Error closing connection: %v", err)
			}
		}
	}
}

func (h *Handler) Shutdown() {
	h.backoffQueue.Clear()
	h.workerPool.Shutdown()
}

func (h *Handler) handleConnection(conn net.Conn) {
	start := time.Now()
	defer func() {
		metrics.ResponseTime.Observe(time.Since(start).Seconds())
	}()

	// Check rate limit
	if !h.rateLimiter.Allow() {
		metrics.RateLimitHits.Inc()
		log.Printf("Rate limit exceeded for %s", conn.RemoteAddr())
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
		return
	}

	// Check connection limit
	if !h.connLimiter.Acquire() {
		log.Printf("Connection limit exceeded for %s", conn.RemoteAddr())
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
		return
	}
	defer h.connLimiter.Release()

	metrics.ActiveConnections.Inc()
	metrics.TotalConnections.Inc()
	defer metrics.ActiveConnections.Dec()

	// Set timeouts
	if err := h.connManager.SetTimeouts(conn); err != nil {
		log.Printf("Error setting timeouts: %v", err)
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
		return
	}

	// Generate challenge
	challenge := h.powService.GenerateChallenge()
	metrics.PoWChallengesGenerated.Inc()

	// Send challenge
	challengeStr := h.powService.FormatChallenge(challenge)
	if err := h.connManager.Write(conn, challengeStr); err != nil {
		log.Printf("Error sending challenge: %v", err)
		return
	}

	// Read nonce
	nonceStr, err := h.connManager.ReadWithRetry(conn)
	if err != nil {
		log.Printf("Error reading nonce: %v", err)
		return
	}

	nonce, err := strconv.ParseInt(strings.TrimSpace(nonceStr), 10, 64)
	if err != nil {
		log.Printf("Error parsing nonce: %v", err)
		return
	}

	// Verify PoW
	metrics.PoWChallengesVerified.Inc()
	if !h.powService.VerifyPoW(challenge, nonce) {
		metrics.PoWVerificationFailures.Inc()
		log.Printf("Invalid PoW from %s", conn.RemoteAddr())
		if err := h.connManager.Write(conn, "Invalid proof of work"); err != nil {
			log.Printf("Error sending invalid PoW response: %v", err)
		}
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
		return
	}

	// Send quote
	quote := h.quoteService.GetRandomQuote()
	if err := h.connManager.Write(conn, quote); err != nil {
		log.Printf("Error sending quote: %v", err)
		return
	}
	if err := conn.Close(); err != nil {
		log.Printf("Error closing connection: %v", err)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.config.Server.Mode == "quotes" {
		quote := h.quoteService.GetRandomQuote()
		if _, err := w.Write([]byte(quote)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
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
	}
}
