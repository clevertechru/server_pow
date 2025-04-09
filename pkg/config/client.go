package config

import (
	"time"
)

type ClientConfig struct {
	ServerHost      string
	ServerPort      string
	RequestsDelayMs time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

func ClientConfigNew() *ClientConfig {
	return &ClientConfig{
		ServerHost:      getEnvOrDefault("SERVER_HOST", "server"),
		ServerPort:      getEnvOrDefault("SERVER_PORT", "8080"),
		RequestsDelayMs: getDurationEnvOrDefault("REQUESTS_DELAY_MS", 100*time.Millisecond),
		ReadTimeout:     getDurationEnvOrDefault("READ_TIMEOUT_MS", 30_000*time.Millisecond),
		WriteTimeout:    getDurationEnvOrDefault("WRITE_TIMEOUT_MS", 30_000*time.Millisecond),
	}
}
