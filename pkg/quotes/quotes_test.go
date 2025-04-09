package quotes

import (
	"testing"
)

func TestGetRandomQuote(t *testing.T) {
	quote := GetRandomQuote()
	if quote == "" {
		t.Error("Expected non-empty quote")
	}

	// Check if the quote is in our list
	found := false
	for _, q := range quotes {
		if q == quote {
			found = true
			break
		}
	}
	if !found {
		t.Error("Quote not found in the list")
	}
}
