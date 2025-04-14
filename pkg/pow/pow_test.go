package pow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

	var nonce int64
	for nonce = 0; nonce < 1000000 && !VerifyPoW(challenge, nonce); nonce++ {
	}
	require.True(t, VerifyPoW(challenge, nonce), "Failed to find valid nonce")
}

func TestSolvePoW(t *testing.T) {
	challenge := GenerateChallenge("0000")
	var nonce int64
	for nonce = 0; nonce < 1000000 && !VerifyPoW(challenge, nonce); nonce++ {
	}
	require.True(t, VerifyPoW(challenge, nonce), "Failed to find valid nonce")
}
