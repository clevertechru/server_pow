package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/clevertechru/server_pow/pkg/backoff"
	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/connlimit"
	"github.com/clevertechru/server_pow/pkg/nonce"
	"github.com/clevertechru/server_pow/pkg/pow"
	"github.com/clevertechru/server_pow/pkg/quotes"
	"github.com/clevertechru/server_pow/pkg/ratelimit"
	"github.com/clevertechru/server_pow/pkg/workerpool"
)

type Handler struct {
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
	h := &Handler{
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

	// Start queue processor
	go h.processQueue()

	return h
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
		if _, err := conn.Write([]byte("Connection limit exceeded\n")); err != nil {
			log.Printf("Error writing to connection: %v", err)
		}
		return
	}
	defer h.connLimiter.Release()

	if !h.rateLimiter.Allow() {
		log.Printf("Rate limit exceeded for %s", conn.RemoteAddr())
		if _, err := conn.Write([]byte("Rate limit exceeded\n")); err != nil {
			log.Printf("Error writing to connection: %v", err)
		}
		return
	}

	if err := conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout)); err != nil {
		log.Printf("Error setting read deadline: %v", err)
		return
	}
	if err := conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout)); err != nil {
		log.Printf("Error setting write deadline: %v", err)
		return
	}

	challenge := pow.GenerateChallenge(h.config.ChallengeDifficulty)
	challengeStr := fmt.Sprintf("%s|%d|%d", challenge.Data, challenge.Target, challenge.Timestamp)
	log.Printf("Sending challenge: %s", challengeStr)
	if _, err := conn.Write([]byte(challengeStr + "\n")); err != nil {
		log.Printf("Error writing challenge: %v", err)
		return
	}

	// Read nonce
	bufferPtr := h.pool.Get().(*[]byte)
	buffer := *bufferPtr
	defer func() {
		if bufferPtr != nil {
			h.pool.Put(bufferPtr)
		}
	}()

	// Read nonce with retries
	var nonce string
	for {
		if err := conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout)); err != nil {
			log.Printf("Error setting read deadline: %v", err)
			return
		}
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Read timeout: %v", err)
				continue
			}
			log.Printf("Error reading nonce: %v", err)
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

	log.Printf("Received nonce: %s", nonce)
	var nonceInt int64
	if _, err := fmt.Sscanf(nonce, "%d", &nonceInt); err != nil {
		log.Printf("Error parsing nonce: %v", err)
		return
	}

	// Parse challenge string
	parts := strings.Split(challengeStr, "|")
	if len(parts) != 3 {
		log.Printf("Invalid challenge format")
		return
	}

	target, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Invalid target: %v", err)
		return
	}

	timestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		log.Printf("Invalid timestamp: %v", err)
		return
	}

	verifyChallenge := pow.Challenge{
		Data:      parts[0],
		Target:    target,
		Timestamp: timestamp,
	}
	log.Printf("Verifying challenge: %+v with nonce: %d", verifyChallenge, nonceInt)

	if !pow.VerifyPoW(verifyChallenge, nonceInt) {
		log.Printf("Invalid proof of work")
		if _, err := conn.Write([]byte("Invalid proof of work\n")); err != nil {
			log.Printf("Error writing to connection: %v", err)
		}
		return
	}

	if !h.nonceTracker.IsValid(uint64(nonceInt)) {
		log.Printf("Replay attack detected for nonce %d", nonceInt)
		if _, err := conn.Write([]byte("Replay attack detected\n")); err != nil {
			log.Printf("Error writing to connection: %v", err)
		}
		return
	}

	// Only send quote after successful PoW validation
	quote := quotes.GetRandomQuote()
	log.Printf("Sending quote: %s", quote)
	if _, err := conn.Write([]byte(quote + "\n")); err != nil {
		log.Printf("Error writing quote: %v", err)
	}
}
