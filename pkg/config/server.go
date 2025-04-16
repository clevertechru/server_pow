package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerSettings struct {
	Host                string        // Host to bind to
	Port                string        // Port to listen on
	ChallengeDifficulty int           // Difficulty target for PoW
	ReadTimeout         time.Duration // Read timeout for connections
	WriteTimeout        time.Duration // Write timeout for connections
	RateLimit           int           // Rate limit in requests per second
	BurstLimit          int           // Burst capacity for rate limiter
	MaxConnections      int           // Maximum number of concurrent connections
	WorkerPoolSize      int           // Size of the worker pool
	QueueSize           int           // Size of the connection queue
	BaseBackoff         time.Duration // Base backoff duration
	MaxBackoff          time.Duration // Maximum backoff duration
}

type ServerConfig struct {
	Server struct {
		Mode  string `yaml:"mode"`
		Proxy struct {
			Target  string `yaml:"target"`
			Timeout string `yaml:"timeout"`
		} `yaml:"proxy"`
		Quotes struct {
			File string `yaml:"file"`
		} `yaml:"quotes"`
	} `yaml:"server"`
}

func NewServerSettings() *ServerSettings {
	return &ServerSettings{
		Host:                getEnvOrDefault("HOST", "0.0.0.0"),
		Port:                getEnvOrDefault("PORT", "8080"),
		ChallengeDifficulty: getIntEnvOrDefault("CHALLENGE_DIFFICULTY", 2),
		ReadTimeout:         getDurationEnvOrDefault("READ_TIMEOUT_MS", 30_000*time.Millisecond),
		WriteTimeout:        getDurationEnvOrDefault("WRITE_TIMEOUT_MS", 30_000*time.Millisecond),
		RateLimit:           getIntEnvOrDefault("RATE_LIMIT_RPS", 10),
		BurstLimit:          getIntEnvOrDefault("BURST_CAPACITY", 20),
		MaxConnections:      getIntEnvOrDefault("MAX_ACTIVE_CONNECTIONS", 100),
		WorkerPoolSize:      getIntEnvOrDefault("WORKER_POOL_SIZE", 10),
		QueueSize:           getIntEnvOrDefault("QUEUE_SIZE", 50),
		BaseBackoff:         getDurationEnvOrDefault("BASE_BACKOFF_MS", 100*time.Millisecond),
		MaxBackoff:          getDurationEnvOrDefault("MAX_BACKOFF_MS", 5000*time.Millisecond),
	}
}

func LoadConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *ServerConfig) GetProxyTimeout() (time.Duration, error) {
	return time.ParseDuration(c.Server.Proxy.Timeout)
}

// DefaultConfig returns a default server configuration
func DefaultConfig() *ServerConfig {
	cfg := &ServerConfig{}
	cfg.Server.Mode = "quotes"
	cfg.Server.Proxy.Target = "http://example.com"
	cfg.Server.Proxy.Timeout = "5s"
	cfg.Server.Quotes.File = "quotes.yml"
	return cfg
}
