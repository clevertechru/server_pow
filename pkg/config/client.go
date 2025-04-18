package config

import (
	"fmt"
	"log"
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
		Protocol     string `yaml:"protocol"`      // Server protocol (http or https)
		TLS          struct {
			InsecureSkipVerify bool `yaml:"insecure_skip_verify"` // Skip TLS certificate verification
		} `yaml:"tls"`
		Connection struct {
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
	cfg.Client.Protocol = "http"
	cfg.Client.Connection.ReadTimeout = "30s"
	cfg.Client.Connection.WriteTimeout = "30s"
	return cfg
}

func (c *ClientConfig) GetReadTimeout() time.Time {
	timeout, err := time.ParseDuration(c.Client.Connection.ReadTimeout)
	if err != nil {
		log.Printf("Error parse read timeout, set default value: %v", err)
		timeout, _ = time.ParseDuration(DefaultClientConfig().Client.Connection.ReadTimeout)
	}
	return time.Now().Add(timeout)
}

func (c *ClientConfig) GetWriteTimeout() time.Time {
	timeout, err := time.ParseDuration(c.Client.Connection.WriteTimeout)
	if err != nil {
		log.Printf("Error parse write timeout, set default value: %v", err)
		timeout, _ = time.ParseDuration(DefaultClientConfig().Client.Connection.WriteTimeout)
	}
	return time.Now().Add(timeout)
}

func (c *ClientConfig) GetRequestDelay() time.Duration {
	delay, err := time.ParseDuration(c.Client.RequestDelay)
	if err != nil {
		log.Printf("Error parse duration, set default value: %v", err)
		delay, _ = time.ParseDuration(DefaultClientConfig().Client.RequestDelay)
	}
	return delay
}
