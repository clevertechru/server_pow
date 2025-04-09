package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net"
	"server/config"
	"strings"
	"time"
)

var quotes = []string{
	"The only way to do great work is to love what you do. - Steve Jobs",
	"Stay hungry, stay foolish. - Steve Jobs",
	"Code is like humor. When you have to explain it, it's bad. - Cory House",
	"First, solve the problem. Then, write the code. - John Johnson",
	"Experience is the name everyone gives to their mistakes. - Oscar Wilde",
	"Programming isn't about what you know; it's about what you can figure out. - Chris Pine",
	"The only way to learn a new programming language is by writing programs in it. - Dennis Ritchie",
	"Talk is cheap. Show me the code. - Linus Torvalds",
	"Programming is the art of telling another human what one wants the computer to do. - Donald Knuth",
	"Clean code always looks like it was written by someone who cares. - Robert C. Martin",
}

type Challenge struct {
	Data      string
	Target    string
	Timestamp int64
}

func generateChallenge(difficulty string) Challenge {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	data := make([]byte, 32)
	r.Read(data)

	return Challenge{
		Data:      hex.EncodeToString(data),
		Target:    difficulty,
		Timestamp: time.Now().Unix(),
	}
}

func verifyPoW(challenge Challenge, nonce int64) bool {
	data := fmt.Sprintf("%s%d%d", challenge.Data, challenge.Timestamp, nonce)
	hash := sha256.Sum256([]byte(data))
	hexHash := hex.EncodeToString(hash[:])
	return strings.HasPrefix(hexHash, challenge.Target)
}

func handleConnection(conn net.Conn, cfg *config.ServerConfig) {
	defer conn.Close()

	challenge := generateChallenge(cfg.ChallengeDifficulty)
	challengeStr := fmt.Sprintf("%s|%s|%d", challenge.Data, challenge.Target, challenge.Timestamp)
	conn.Write([]byte(challengeStr + "\n"))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading nonce: %v", err)
		return
	}

	nonce := string(buffer[:n])
	nonce = strings.TrimSpace(nonce)
	var nonceInt int64
	fmt.Sscanf(nonce, "%d", &nonceInt)
	if !verifyPoW(challenge, nonceInt) {
		conn.Write([]byte("Invalid proof of work\n"))
		return
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	quote := quotes[r.Intn(len(quotes))]
	conn.Write([]byte(quote + "\n"))
}

func main() {
	cfg := config.New()
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server started on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn, cfg)
	}
}
