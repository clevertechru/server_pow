package config

import (
	"os"
	"strconv"
	"time"
)

type ClientConfig struct {
	ServerHost      string
	ServerPort      string
	RequestsDelayMs time.Duration
}

func New() *ClientConfig {
	return &ClientConfig{
		ServerHost:      getEnvOrDefault("SERVER_HOST", "server"),
		ServerPort:      getEnvOrDefault("SERVER_PORT", "8080"),
		RequestsDelayMs: getDurationEnvOrDefault("REQUESTS_DELAY_MS", 100*time.Millisecond),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if delay, err := strconv.Atoi(value); err == nil {
			return time.Duration(delay) * time.Millisecond
		}
	}
	return defaultValue
}
