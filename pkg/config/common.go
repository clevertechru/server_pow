package config

import (
	"os"
	"strconv"
	"time"
)

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
