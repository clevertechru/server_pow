package quotes

import (
	"math/rand"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Storage struct {
	mu     sync.RWMutex
	quotes []string
}

type Config struct {
	Quotes []string `yaml:"quotes"`
}

func NewStorage(configPath string) (*Storage, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &Storage{
		quotes: config.Quotes,
	}, nil
}

func (s *Storage) GetRandomQuote() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.quotes) == 0 {
		return ""
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return s.quotes[r.Intn(len(s.quotes))]
}
