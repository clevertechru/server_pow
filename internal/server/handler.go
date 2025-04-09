package server

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/clevertechru/server_pow/pkg/config"

	"github.com/clevertechru/server_pow/pkg/pow"
	"github.com/clevertechru/server_pow/pkg/quotes"
)

type Handler struct {
	config *config.ServerConfig
}

func NewHandler(config *config.ServerConfig) *Handler {
	return &Handler{config: config}
}

func (h *Handler) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// Set read/write timeouts
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	challenge := pow.GenerateChallenge(h.config.ChallengeDifficulty)
	challengeStr := fmt.Sprintf("%s|%s|%d", challenge.Data, challenge.Target, challenge.Timestamp)
	log.Printf("Sending challenge: %s", challengeStr)
	conn.Write([]byte(challengeStr + "\n"))

	buffer := make([]byte, 1024)
	var nonce string
	for {
		// Reset read deadline for each read attempt
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
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
