package config

import (
	"os"
)

type ServerConfig struct {
	Host                string
	Port                string
	ChallengeDifficulty string
}

func New() *ServerConfig {
	return &ServerConfig{
		Host:                getEnvOrDefault("HOST", "0.0.0.0"),
		Port:                getEnvOrDefault("PORT", "8080"),
		ChallengeDifficulty: getEnvOrDefault("CHALLENGE_DIFFICULTY", "0000"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
