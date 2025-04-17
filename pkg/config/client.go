package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	Client struct {
		ServerHost   string `yaml:"server_host"`   // Server host to connect to
		ServerPort   string `yaml:"server_port"`   // Server port to connect to
		RequestDelay string `yaml:"request_delay"` // Requests delay
		Connection   struct {
			ReadTimeout  string `yaml:"read_timeout"`  // Read timeout for connections
			WriteTimeout string `yaml:"write_timeout"` // Write timeout for connections
		} `yaml:"connection"`
	} `yaml:"client"`
}

func LoadClientConfig(configPath string) (*ClientConfig, error) {
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

	var cfg ClientConfig
	if err := yaml.Unmarshal([]byte(configStr), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// DefaultClientConfig returns a default client configuration
func DefaultClientConfig() *ClientConfig {
	cfg := &ClientConfig{}
	cfg.Client.ServerHost = "localhost"
	cfg.Client.ServerPort = "8080"
	cfg.Client.RequestDelay = "100ms"
	cfg.Client.Connection.ReadTimeout = "30s"
	cfg.Client.Connection.WriteTimeout = "30s"
	return cfg
}

func (c *ClientConfig) GetReadTimeout() (time.Duration, error) {
	return time.ParseDuration(c.Client.Connection.ReadTimeout)
}

func (c *ClientConfig) GetWriteTimeout() (time.Duration, error) {
	return time.ParseDuration(c.Client.Connection.WriteTimeout)
}

func (c *ClientConfig) GetRequestDelay() (time.Duration, error) {
	return time.ParseDuration(c.Client.RequestDelay)
}
