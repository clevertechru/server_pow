package server

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"
	"github.com/clevertechru/server_pow/pkg/pow"
	"github.com/clevertechru/server_pow/pkg/quotes"
	"github.com/clevertechru/server_pow/pkg/ratelimit"
)

type Handler struct {
	config  *config.ServerConfig
	pool    *sync.Pool
	limiter *ratelimit.Limiter
}

func NewHandler(config *config.ServerConfig) *Handler {
	return &Handler{
		config: config,
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024)
			},
		},
		limiter: ratelimit.NewLimiter(float64(config.RateLimit), int64(config.BurstLimit)),
	}
}

func (h *Handler) HandleConnection(conn net.Conn) {
	defer conn.Close()

	if !h.limiter.Allow() {
		log.Printf("Rate limit exceeded for %s", conn.RemoteAddr())
		conn.Write([]byte("Rate limit exceeded\n"))
		return
	}

	conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))

	challenge := pow.GenerateChallenge(h.config.ChallengeDifficulty)
	challengeStr := fmt.Sprintf("%s|%s|%d", challenge.Data, challenge.Target, challenge.Timestamp)
	log.Printf("Sending challenge: %s", challengeStr)
	conn.Write([]byte(challengeStr + "\n"))

	buffer := h.pool.Get().([]byte)
	defer h.pool.Put(buffer)

	var nonce string
	for {
		// Reset read deadline for each read attempt
		conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout))
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Read timeout: %v", err)
			} else {
				log.Printf("Error reading nonce: %v", err)
			}
			return
		}
		if n == 0 {
			continue
		}
		nonce = string(buffer[:n])
		nonce = strings.TrimSpace(nonce)
		if nonce != "" {
			break
		}
	}

	log.Printf("Received nonce: %s", nonce)
	var nonceInt int64
	fmt.Sscanf(nonce, "%d", &nonceInt)

	// Reconstruct the challenge for verification
	parts := strings.Split(challengeStr, "|")
	verifyChallenge := pow.Challenge{
		Data:      parts[0],
		Target:    parts[1],
		Timestamp: challenge.Timestamp,
	}
	log.Printf("Verifying challenge: %+v with nonce: %d", verifyChallenge, nonceInt)

	if !pow.VerifyPoW(verifyChallenge, nonceInt) {
		log.Printf("Invalid proof of work")
		conn.Write([]byte("Invalid proof of work\n"))
		return
	}

	quote := quotes.GetRandomQuote()
	log.Printf("Sending quote: %s", quote)
	conn.Write([]byte(quote + "\n"))
}
