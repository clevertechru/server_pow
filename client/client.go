package main

import (
	"bufio"
	"client/config"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func solvePoW(challengeStr string) int64 {
	parts := strings.Split(challengeStr, "|")
	if len(parts) != 3 {
		log.Fatal("Invalid challenge format")
	}

	data := parts[0]
	target := parts[1]
	timestamp, _ := strconv.ParseInt(parts[2], 10, 64)

	nonce := int64(0)
	for {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s%d%d", data, timestamp, nonce)))
		hexHash := hex.EncodeToString(hash[:])

		if strings.HasPrefix(hexHash, target) {
			return nonce
		}
		nonce++
	}
}

func makeRequest(cfg *config.ClientConfig) error {
	addr := net.JoinHostPort(cfg.ServerHost, cfg.ServerPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	challenge, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read challenge: %v", err)
	}
	challenge = strings.TrimSpace(challenge)
	conn.Write([]byte(fmt.Sprintf("%d\n", solvePoW(challenge))))

	quote, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read quote: %v", err)
	}
	fmt.Printf("Received quote: %s", quote)
	return nil
}

func main() {
	cfg := config.New()
	for {
		if err := makeRequest(cfg); err != nil {
			log.Printf("Error: %v", err)
		}
		time.Sleep(cfg.RequestsDelayMs)
	}
}
