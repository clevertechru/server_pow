package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/clevertechru/server_pow/pkg/nonce"
	"github.com/clevertechru/server_pow/pkg/pow"
)

type PoWService struct {
	difficulty   int
	nonceTracker *nonce.Tracker
}

func NewPoWService(difficulty int, nonceWindow time.Duration) (*PoWService, error) {
	return &PoWService{
		difficulty:   difficulty,
		nonceTracker: nonce.NewTracker(nonceWindow),
	}, nil
}

func (s *PoWService) GenerateChallenge() pow.Challenge {
	return pow.GenerateChallenge(s.difficulty)
}

func (s *PoWService) FormatChallenge(challenge pow.Challenge) string {
	return fmt.Sprintf("%s|%d|%d", challenge.Data, challenge.Target, challenge.Timestamp)
}

func (s *PoWService) ParseChallenge(challengeStr string) (pow.Challenge, error) {
	parts := strings.Split(challengeStr, "|")
	if len(parts) != 3 {
		return pow.Challenge{}, fmt.Errorf("invalid challenge format")
	}

	target, err := strconv.Atoi(parts[1])
	if err != nil {
		return pow.Challenge{}, fmt.Errorf("invalid target: %w", err)
	}

	timestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return pow.Challenge{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	return pow.Challenge{
		Data:      parts[0],
		Target:    target,
		Timestamp: timestamp,
	}, nil
}

func (s *PoWService) VerifyPoW(challenge pow.Challenge, nonce int64) bool {
	return pow.VerifyPoW(challenge, nonce)
}

func (s *PoWService) ValidateNonce(nonce uint64) bool {
	return s.nonceTracker.IsValid(nonce)
}
