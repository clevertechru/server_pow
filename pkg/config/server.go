package config

import "time"

type ServerConfig struct {
	Host                string
	Port                string
	ChallengeDifficulty string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	RateLimit           int
	BurstLimit          int
	MaxConnections      int
}

func ServerConfigNew() *ServerConfig {
	return &ServerConfig{
		Host:                getEnvOrDefault("HOST", "0.0.0.0"),
		Port:                getEnvOrDefault("PORT", "8080"),
		ChallengeDifficulty: getEnvOrDefault("CHALLENGE_DIFFICULTY", "0000"),
		ReadTimeout:         getDurationEnvOrDefault("READ_TIMEOUT_MS", 30_000*time.Millisecond),
		WriteTimeout:        getDurationEnvOrDefault("WRITE_TIMEOUT_MS", 30_000*time.Millisecond),
		RateLimit:           getIntEnvOrDefault("RATE_LIMIT_RPS", 10), // requests per second
		BurstLimit:          getIntEnvOrDefault("BURST_CAPACITY", 20), // burst capacity
		MaxConnections:      getIntEnvOrDefault("MAX_ACTIVE_CONNECTIONS", 100),
	}
}
