package server

import (
	"fmt"
	"log"
	"net"
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
	config       *config.ServerSettings
	rateLimiter  *ratelimit.Limiter
	connLimiter  *connlimit.Limiter
	workerPool   *workerpool.Pool
	backoffQueue *backoff.Queue
	powService   *service.PoWService
	quoteService *service.QuoteService
	connManager  *service.ConnectionManager
}

func NewHandler(config *config.ServerSettings) (*Handler, error) {
	powService, err := service.NewPoWService(config.ChallengeDifficulty, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to create PoW service: %w", err)
	}

	h := &Handler{
		config:       config,
		rateLimiter:  ratelimit.NewLimiter(float64(config.RateLimit), int64(config.BurstLimit)),
		connLimiter:  connlimit.NewLimiter(config.MaxConnections),
		backoffQueue: backoff.NewQueue(config.QueueSize, config.BaseBackoff, config.MaxBackoff),
		powService:   powService,
		quoteService: service.NewQuoteService(),
		connManager:  service.NewConnectionManager(config.ReadTimeout, config.WriteTimeout),
	}

	h.workerPool = workerpool.NewPool(config.WorkerPoolSize, h.handleConnection)

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
