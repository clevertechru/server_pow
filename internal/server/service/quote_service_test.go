package service

import (
	"testing"

	"github.com/clevertechru/server_pow/pkg/quotes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuoteService_GetRandomQuote(t *testing.T) {
	err := quotes.Init("testdata/quotes.yaml")
	require.NoError(t, err)

	service := NewQuoteService()

	// Test that we get a non-empty quote
	quote := service.GetRandomQuote()
	assert.NotEmpty(t, quote, "Expected non-empty quote")

	// Test that we get different quotes (not guaranteed but likely)
	quote2 := service.GetRandomQuote()
	if quote == quote2 {
		t.Log("Got same quote twice, this is possible but unlikely")
	}
}
