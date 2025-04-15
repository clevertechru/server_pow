package quotes

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetRandomQuote(t *testing.T) {
	// Create a temporary quotes file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "quotes.yml")
	quotes := []string{
		"Test quote 1",
		"Test quote 2",
		"Test quote 3",
	}

	data := struct {
		Quotes []string `yaml:"quotes"`
	}{
		Quotes: quotes,
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

	// Initialize storage
	if err := Init(configPath); err != nil {
		t.Fatalf("Failed to initialize quotes storage: %v", err)
	}

	// Test getting a random quote
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
