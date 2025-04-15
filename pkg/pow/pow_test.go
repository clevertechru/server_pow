package pow

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/clevertechru/server_pow/pkg/quotes"
	"gopkg.in/yaml.v3"
)

func TestGenerateChallenge(t *testing.T) {
	// Create a temporary quotes file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "quotes.yml")
	quotesList := []string{
		"Test quote 1",
		"Test quote 2",
		"Test quote 3",
	}

	data := struct {
		Quotes []string `yaml:"quotes"`
	}{
		Quotes: quotesList,
	}

	file, err := os.Create(configPath)
	if err != nil {
		t.Fatalf("Failed to create test quotes file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close test quotes file: %v", err)
		}
	}()

	if err := yaml.NewEncoder(file).Encode(data); err != nil {
		t.Fatalf("Failed to write test quotes: %v", err)
	}

	// Initialize quotes storage
	if err := quotes.Init(configPath); err != nil {
		t.Fatalf("Failed to initialize quotes storage: %v", err)
	}

	difficulty := 4
	challenge := GenerateChallenge(difficulty)

	if challenge.Data == "" {
		t.Error("Expected non-empty data")
	}
	if challenge.Target != difficulty {
		t.Errorf("Expected target %d, got %d", difficulty, challenge.Target)
	}
	if challenge.Timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}
}

func TestVerifyPoW(t *testing.T) {
	challenge := Challenge{
		Data:      "test data",
		Target:    2, // Lower difficulty for faster tests
		Timestamp: time.Now().Unix(),
	}

	// Find a valid nonce with timeout
	nonce := int64(0)
	start := time.Now()
	for !VerifyPoW(challenge, nonce) {
		nonce++
		if time.Since(start) > 100*time.Millisecond {
			t.Fatal("Timeout finding valid nonce")
		}
	}

	if !VerifyPoW(challenge, nonce) {
		t.Error("Expected valid PoW verification")
	}

	// Test with invalid nonce
	if VerifyPoW(challenge, nonce+1) {
		t.Error("Expected invalid PoW verification")
	}
}

func TestSolvePoW(t *testing.T) {
	challenge := Challenge{
		Data:      "test data",
		Target:    2, // Lower difficulty for faster tests
		Timestamp: time.Now().Unix(),
	}

	challengeStr := fmt.Sprintf("%s|%d|%d", challenge.Data, challenge.Target, challenge.Timestamp)

	// Run in parallel with timeout
	done := make(chan struct{})
	go func() {
		nonce, err := SolvePoW(challengeStr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !VerifyPoW(challenge, nonce) {
			t.Error("Expected valid PoW solution")
		}
		close(done)
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout solving PoW")
	}
}
