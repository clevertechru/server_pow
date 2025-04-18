package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Server struct {
		Host                string `yaml:"host"`
		Port                string `yaml:"port"`
		Mode                string `yaml:"mode"`                 // Server mode quotes or proxy
		ChallengeDifficulty int    `yaml:"challenge_difficulty"` // Difficulty target for PoW
		Protocol            string `yaml:"protocol"`             // Server protocol (http or https)
		TLS                 struct {
			CertFile string `yaml:"cert_file"` // Path to TLS certificate file
			KeyFile  string `yaml:"key_file"`  // Path to TLS private key file
		} `yaml:"tls"`
		Proxy struct {
			Target  string `yaml:"target"`  // Target URL for proxy
			Timeout string `yaml:"timeout"` // Timeout for proxy
		} `yaml:"proxy"`
		Quotes struct {
			File string `yaml:"file"` // File path for quotes.yml
		} `yaml:"quotes"`
		Connection struct {
			ReadTimeout    string `yaml:"read_timeout"`     // Read timeout for connections
			WriteTimeout   string `yaml:"write_timeout"`    // Write timeout for connections
			RateLimit      int    `yaml:"rate_limit"`       // Rate limit in requests per second
			BurstLimit     int    `yaml:"burst_limit"`      // Burst capacity for rate limiter
			MaxConnections int    `yaml:"max_connections"`  // Maximum number of concurrent connections
			WorkerPoolSize int    `yaml:"worker_pool_size"` // Size of the worker pool
			QueueSize      int    `yaml:"queue_size"`       // Size of the connection queue
			BaseBackoff    string `yaml:"base_backoff"`     // Base backoff duration
			MaxBackoff     string `yaml:"max_backoff"`      // Maximum backoff duration
		} `yaml:"connection"`
	} `yaml:"server"`
}

func LoadServerConfig(configPath string) (*ServerConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Replace environment variables
	re := regexp.MustCompile(`\${([^}]+)}`)
	configStr := re.ReplaceAllStringFunc(string(data), func(match string) string {
		envVar := strings.Trim(match, "${}")
		return os.Getenv(envVar)
	})

	var cfg ServerConfig
	if err := yaml.Unmarshal([]byte(configStr), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func (c *ServerConfig) GetProxyTimeout() (time.Duration, error) {
	return time.ParseDuration(c.Server.Proxy.Timeout)
}

// DefaultConfig returns a default server configuration
func DefaultServerConfig() *ServerConfig {
	cfg := &ServerConfig{}
	cfg.Server.Mode = "quotes"
	cfg.Server.Protocol = "http"
	cfg.Server.Proxy.Target = "http://example.com"
	cfg.Server.Proxy.Timeout = "5s"
	cfg.Server.Quotes.File = "quotes.yml"
	cfg.Server.Connection.ReadTimeout = "30s"
	cfg.Server.Connection.WriteTimeout = "30s"
	cfg.Server.Connection.RateLimit = 10
	cfg.Server.Connection.BurstLimit = 20
	cfg.Server.Connection.MaxConnections = 100
	cfg.Server.Connection.WorkerPoolSize = 10
	cfg.Server.Connection.QueueSize = 50
	cfg.Server.Connection.BaseBackoff = "100ms"
	cfg.Server.Connection.MaxBackoff = "5s"
	return cfg
}
