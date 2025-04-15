package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/clevertechru/server_pow/internal/server/service"
	"github.com/clevertechru/server_pow/pkg/backoff"
	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/connlimit"
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
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	if !h.connLimiter.Acquire() {
		log.Printf("Connection limit exceeded for %s", conn.RemoteAddr())
		if err := h.connManager.Write(conn, "Connection limit exceeded"); err != nil {
			log.Printf("Error writing to connection: %v", err)
		}
		return
	}
	defer h.connLimiter.Release()

	if !h.rateLimiter.Allow() {
		log.Printf("Rate limit exceeded for %s", conn.RemoteAddr())
		if err := h.connManager.Write(conn, "Rate limit exceeded"); err != nil {
			log.Printf("Error writing to connection: %v", err)
		}
		return
	}

	challenge := h.powService.GenerateChallenge()
	challengeStr := h.powService.FormatChallenge(challenge)
	log.Printf("Sending challenge: %s", challengeStr)
	if err := h.connManager.Write(conn, challengeStr); err != nil {
		log.Printf("Error writing challenge: %v", err)
		return
	}

	nonceStr, err := h.connManager.ReadWithRetry(conn)
	if err != nil {
		log.Printf("Error reading nonce: %v", err)
		return
	}

	log.Printf("Received nonce: %s", nonceStr)
	nonceInt, err := strconv.ParseInt(nonceStr, 10, 64)
	if err != nil {
		log.Printf("Error parsing nonce: %v", err)
		return
	}

	verifyChallenge, err := h.powService.ParseChallenge(challengeStr)
	if err != nil {
		log.Printf("Error parsing challenge: %v", err)
		return
	}

	log.Printf("Verifying challenge: %+v with nonce: %d", verifyChallenge, nonceInt)
	if !h.powService.VerifyPoW(verifyChallenge, nonceInt) {
		log.Printf("Invalid proof of work")
		if err := h.connManager.Write(conn, "Invalid proof of work"); err != nil {
			log.Printf("Error writing to connection: %v", err)
		}
		return
	}

	if !h.powService.ValidateNonce(uint64(nonceInt)) {
		log.Printf("Replay attack detected for nonce %d", nonceInt)
		if err := h.connManager.Write(conn, "Replay attack detected"); err != nil {
			log.Printf("Error writing to connection: %v", err)
		}
		return
	}

	quote := h.quoteService.GetRandomQuote()
	log.Printf("Sending quote: %s", quote)
	if err := h.connManager.Write(conn, quote); err != nil {
		log.Printf("Error writing quote: %v", err)
	}
}
