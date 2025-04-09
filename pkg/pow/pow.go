package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/clevertechru/server_pow/pkg/quotes"
	"strconv"
	"strings"
	"time"
)

type Challenge struct {
	Data      string
	Target    string
	Timestamp int64
}

func GenerateChallenge(difficulty string) Challenge {
	return Challenge{
		Data:      quotes.GetRandomQuote(),
		Target:    difficulty,
		Timestamp: time.Now().Unix(),
	}
}

func VerifyPoW(challenge Challenge, nonce int64) bool {
	data := fmt.Sprintf("%s%d%d", challenge.Data, challenge.Timestamp, nonce)
	hash := sha256.Sum256([]byte(data))
	hexHash := hex.EncodeToString(hash[:])
	return strings.HasPrefix(hexHash, challenge.Target)
}

func SolvePoW(challengeStr string) (int64, error) {
	parts := strings.Split(challengeStr, "|")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid challenge format")
	}

	data := parts[0]
	target := parts[1]
	timestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp: %v", err)
	}

	nonce := int64(0)
	for {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s%d%d", data, timestamp, nonce)))
		hexHash := hex.EncodeToString(hash[:])

		if strings.HasPrefix(hexHash, target) {
			return nonce, nil
		}
		nonce++
	}
}
