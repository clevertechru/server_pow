package quotes

import (
	"math/rand"
	"os"
	"time"

	genericStorage "github.com/clevertechru/server_pow/pkg/storage"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Quotes []string `yaml:"quotes"`
}

type QuotesStorage struct {
	storage genericStorage.Storage
}

func NewStorage(configPath string) (*QuotesStorage, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	s := genericStorage.NewInMemoryStorage()
	for i, quote := range config.Quotes {
		s.Set(string(rune(i)), quote)
	}

	return &QuotesStorage{
		storage: s,
	}, nil
}

func (s *QuotesStorage) GetRandomQuote() string {
	// Get all keys
	keys := make([]string, 0)
	for i := 0; ; i++ {
		key := string(rune(i))
		if _, ok := s.storage.Get(key); !ok {
			break
		}
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return ""
	}

	// Get random quote
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomKey := keys[r.Intn(len(keys))]
	quote, _ := s.storage.Get(randomKey)
	return quote.(string)
}
