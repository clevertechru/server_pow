package pow

import (
	"fmt"
	"testing"
	"time"
)

func TestGenerateChallenge(t *testing.T) {
	difficulty := "0000"
	challenge := GenerateChallenge(difficulty)

	if challenge.Target != difficulty {
		t.Errorf("Expected target %s, got %s", difficulty, challenge.Target)
	}

	if challenge.Timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}

	if len(challenge.Data) == 0 {
		t.Error("Expected non-empty data")
	}
}

func TestVerifyPoW(t *testing.T) {
	challenge := Challenge{
		Data:      "testdata",
		Target:    "0000",
		Timestamp: time.Now().Unix(),
	}

	if VerifyPoW(challenge, 0) {
		t.Error("Expected invalid PoW")
	}

	nonce := int64(0)
	for {
		if VerifyPoW(challenge, nonce) {
			break
		}
		nonce++
		if nonce > 1000000 {
			t.Fatal("Could not find valid nonce")
		}
	}

	if !VerifyPoW(challenge, nonce) {
		t.Error("Expected valid PoW")
	}
}

func TestSolvePoW(t *testing.T) {
	challenge := Challenge{
		Data:      "testdata",
		Target:    "0000",
		Timestamp: time.Now().Unix(),
	}
	challengeStr := fmt.Sprintf("%s|%s|%d", challenge.Data, challenge.Target, challenge.Timestamp)

	nonce, err := SolvePoW(challengeStr)
	if err != nil {
		t.Fatalf("Failed to solve PoW: %v", err)
	}

	if !VerifyPoW(challenge, nonce) {
		t.Error("Solved nonce does not verify")
	}
}
