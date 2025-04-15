package pow

import (
	"crypto/sha256"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
	"time"

	"github.com/clevertechru/server_pow/pkg/quotes"
)

type Challenge struct {
	Data      string
	Target    int
	Timestamp int64
}

func GenerateChallenge(difficulty int) Challenge {
	return Challenge{
		Data:      quotes.GetRandomQuote(),
		Target:    difficulty,
		Timestamp: time.Now().Unix(),
	}
}

func VerifyPoW(challenge Challenge, nonce int64) bool {
	data := fmt.Sprintf("%s%d%d", challenge.Data, challenge.Timestamp, nonce)
	hash := sha256.Sum256([]byte(data))

	// Count leading zeros in the hash
	leadingZeros := 0
	for _, b := range hash {
		if b == 0 {
			leadingZeros += 8
		} else {
			leadingZeros += bits.LeadingZeros8(b)
			break
		}
	}

	// Target is the number of leading zeros required
	// Each byte has 8 bits, so we need to multiply by 8
	return leadingZeros >= challenge.Target*8
}

func SolvePoW(challengeStr string) (int64, error) {
	parts := strings.Split(challengeStr, "|")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid challenge format")
	}

	data := parts[0]
	target, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid target: %v", err)
	}
	timestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp: %v", err)
	}

	nonce := int64(0)
	for {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s%d%d", data, timestamp, nonce)))

		// Count leading zeros in the hash
		leadingZeros := 0
		for _, b := range hash {
			if b == 0 {
				leadingZeros += 8
			} else {
				leadingZeros += bits.LeadingZeros8(b)
				break
			}
		}

		if leadingZeros >= target*8 {
			return nonce, nil
		}
		nonce++
	}
}
