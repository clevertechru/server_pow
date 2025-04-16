package service

import (
	"testing"
	"time"

	"github.com/clevertechru/server_pow/pkg/quotes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	err := quotes.Init("testdata/quotes.yaml")
	if err != nil {
		panic(err)
	}
}

func TestPoWService_GenerateChallenge(t *testing.T) {
	service, err := NewPoWService(5, time.Minute)
	require.NoError(t, err)

	challenge := service.GenerateChallenge()
	assert.NotEmpty(t, challenge.Data, "Expected non-empty challenge data")
	assert.Equal(t, 5, challenge.Target, "Expected target 5")
	assert.NotZero(t, challenge.Timestamp, "Expected non-zero timestamp")
}

func TestPoWService_FormatAndParseChallenge(t *testing.T) {
	service, err := NewPoWService(5, time.Minute)
	require.NoError(t, err)

	challenge := service.GenerateChallenge()
	formatted := service.FormatChallenge(challenge)

	parsed, err := service.ParseChallenge(formatted)
	require.NoError(t, err)

	assert.Equal(t, challenge.Data, parsed.Data, "Data mismatch")
	assert.Equal(t, challenge.Target, parsed.Target, "Target mismatch")
	assert.Equal(t, challenge.Timestamp, parsed.Timestamp, "Timestamp mismatch")
}

func TestPoWService_ParseChallenge_Invalid(t *testing.T) {
	service, err := NewPoWService(5, time.Minute)
	require.NoError(t, err)

	// Test invalid format
	_, err = service.ParseChallenge("invalid")
	assert.Error(t, err, "Expected error for invalid format")

	// Test invalid target
	_, err = service.ParseChallenge("data|invalid|1234567890")
	assert.Error(t, err, "Expected error for invalid target")

	// Test invalid timestamp
	_, err = service.ParseChallenge("data|5|invalid")
	assert.Error(t, err, "Expected error for invalid timestamp")
}

func TestPoWService_VerifyPoW(t *testing.T) {
	service, err := NewPoWService(5, time.Minute)
	require.NoError(t, err)

	challenge := service.GenerateChallenge()

	// Test with invalid nonce
	assert.False(t, service.VerifyPoW(challenge, 0), "Expected false for invalid nonce")

	// Note: We can't easily test a valid nonce since finding one requires actual PoW computation
}

func TestPoWService_ValidateNonce(t *testing.T) {
	service, err := NewPoWService(5, time.Millisecond*100)
	require.NoError(t, err)

	// Test valid nonce
	assert.True(t, service.ValidateNonce(12345), "Expected true for valid nonce")

	// Test same nonce again (should be invalid)
	assert.False(t, service.ValidateNonce(12345), "Expected false for duplicate nonce")

	// Wait for nonce to expire
	time.Sleep(time.Millisecond * 150)

	// Test nonce after expiration
	assert.True(t, service.ValidateNonce(12345), "Expected true for expired and reused nonce")
}
