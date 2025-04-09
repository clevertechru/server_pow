package config

type ServerConfig struct {
	Host                string
	Port                string
	ChallengeDifficulty string
}

func ServerConfigNew() *ServerConfig {
	return &ServerConfig{
		Host:                getEnvOrDefault("HOST", "0.0.0.0"),
		Port:                getEnvOrDefault("PORT", "8080"),
		ChallengeDifficulty: getEnvOrDefault("CHALLENGE_DIFFICULTY", "0000"),
	}
}
