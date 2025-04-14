package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/clevertechru/server_pow/pkg/backoff"
	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/connctx"
	"github.com/clevertechru/server_pow/pkg/connlimit"
	"github.com/clevertechru/server_pow/pkg/nonce"
	"github.com/clevertechru/server_pow/pkg/pow"
	"github.com/clevertechru/server_pow/pkg/quotes"
	"github.com/clevertechru/server_pow/pkg/ratelimit"
	"github.com/clevertechru/server_pow/pkg/workerpool"
)

type Handler struct {
	ctx          context.Context
	cancel       context.CancelFunc
	config       *config.ServerSettings
	pool         *sync.Pool
	rateLimiter  *ratelimit.Limiter
	connLimiter  *connlimit.Limiter
	nonceTracker *nonce.Tracker
	workerPool   *workerpool.Pool
	backoffQueue *backoff.Queue
}

const nonceWindow = 5 * time.Minute // 5-minute window for nonces

func NewHandler(config *config.ServerSettings) *Handler {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Handler{
		ctx:    ctx,
		cancel: cancel,
		config: config,
		pool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 1024)
				return &b
			},
		},
		rateLimiter:  ratelimit.NewLimiter(float64(config.RateLimit), int64(config.BurstLimit)),
		connLimiter:  connlimit.NewLimiter(config.MaxConnections),
		nonceTracker: nonce.NewTracker(nonceWindow),
		backoffQueue: backoff.NewQueue(config.QueueSize, config.BaseBackoff, config.MaxBackoff),
	}

	h.workerPool = workerpool.NewPool(config.WorkerPoolSize, h.handleConnection)
	go h.processQueue()

	return h
}

func (h *Handler) processQueue() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			if conn := h.backoffQueue.Get(); conn != nil {
				if !h.workerPool.Submit(conn) {
					if err := conn.Close(); err != nil {
						log.Printf("Error closing queued connection: %v", err)
					}
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
	h.cancel()
	h.backoffQueue.Clear()
	h.workerPool.Shutdown()
}

func (h *Handler) handleConnection(conn net.Conn) {
	connCtx := connctx.NewConnContext(h.ctx, conn, h.config.ReadTimeout)
	defer connCtx.Close()

	log.Printf("[%s] New connection from %s", connCtx.ID(), connCtx.RemoteAddr())
	defer log.Printf("[%s] Connection closed after %v", connCtx.ID(), connCtx.Duration())

	if !h.connLimiter.Acquire() {
		log.Printf("[%s] Connection limit exceeded", connCtx.ID())
		if _, err := connCtx.Write([]byte("Connection limit exceeded\n")); err != nil {
			log.Printf("[%s] Error writing to connection: %v", connCtx.ID(), err)
		}
		return
	}
	defer h.connLimiter.Release()

	if !h.rateLimiter.Allow() {
		log.Printf("[%s] Rate limit exceeded", connCtx.ID())
		if _, err := connCtx.Write([]byte("Rate limit exceeded\n")); err != nil {
			log.Printf("[%s] Error writing to connection: %v", connCtx.ID(), err)
		}
		return
	}

	challenge := pow.GenerateChallenge(h.config.ChallengeDifficulty)
	challengeStr := fmt.Sprintf("%s|%s|%d", challenge.Data, challenge.Target, challenge.Timestamp)
	log.Printf("[%s] Sending challenge: %s", connCtx.ID(), challengeStr)

	if _, err := connCtx.Write([]byte(challengeStr + "\n")); err != nil {
		log.Printf("[%s] Error writing challenge: %v", connCtx.ID(), err)
		return
	}

	bufferPtr := h.pool.Get().(*[]byte)
	buffer := *bufferPtr
	defer h.pool.Put(bufferPtr)

	var nonce string
	for {
		select {
		case <-connCtx.Done():
			log.Printf("[%s] Connection context canceled", connCtx.ID())
			return
		default:
			n, err := connCtx.Read(buffer)
			if err != nil {
				if err == context.DeadlineExceeded {
					log.Printf("[%s] Read timeout", connCtx.ID())
				} else {
					log.Printf("[%s] Error reading nonce: %v", connCtx.ID(), err)
				}
				return
			}
			if n == 0 {
				continue
			}
			nonce = strings.TrimSpace(string(buffer[:n]))
			if nonce != "" {
				break
			}
		}
	}

	log.Printf("[%s] Received nonce: %s", connCtx.ID(), nonce)
	var nonceInt int64
	if _, err := fmt.Sscanf(nonce, "%d", &nonceInt); err != nil {
		log.Printf("[%s] Error parsing nonce: %v", connCtx.ID(), err)
		return
	}

	parts := strings.Split(challengeStr, "|")
	verifyChallenge := pow.Challenge{
		Data:      parts[0],
		Target:    parts[1],
		Timestamp: challenge.Timestamp,
	}
	log.Printf("[%s] Verifying challenge: %+v with nonce: %d", connCtx.ID(), verifyChallenge, nonceInt)

	if !pow.VerifyPoW(verifyChallenge, nonceInt) {
		log.Printf("[%s] Invalid proof of work", connCtx.ID())
		if _, err := connCtx.Write([]byte("Invalid proof of work\n")); err != nil {
			log.Printf("[%s] Error writing to connection: %v", connCtx.ID(), err)
		}
		return
	}

	if !h.nonceTracker.IsValid(uint64(nonceInt)) {
		log.Printf("[%s] Replay attack detected for nonce %d", connCtx.ID(), nonceInt)
		if _, err := connCtx.Write([]byte("Replay attack detected\n")); err != nil {
			log.Printf("[%s] Error writing to connection: %v", connCtx.ID(), err)
		}
		return
	}

	quote := quotes.GetRandomQuote()
	log.Printf("[%s] Sending quote: %s", connCtx.ID(), quote)
	if _, err := connCtx.Write([]byte(quote + "\n")); err != nil {
		log.Printf("[%s] Error writing quote: %v", connCtx.ID(), err)
	}
}
