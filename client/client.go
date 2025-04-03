package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
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

func makeRequest() error {
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "server"
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	addr := net.JoinHostPort(host, port)
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

func getRequestsDelay() time.Duration {
	envRequestsDelay := os.Getenv("REQUESTS_DELAY_MS")
	if envRequestsDelay != "" {
		delay, err := strconv.Atoi(envRequestsDelay)
		if err != nil {
			panic("Cannot parse integer from REQUESTS_DELAY_MS")
		}
		return time.Duration(delay) * time.Millisecond
	}
	return 100 * time.Millisecond // default delay
}

func main() {
	delay := getRequestsDelay()
	for {
		if err := makeRequest(); err != nil {
			log.Printf("Error: %v", err)
		}
		time.Sleep(delay)
	}
}
