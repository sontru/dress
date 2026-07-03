package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
	Environment string
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8082"))

	return &Config{
		Port:        port,
		DatabaseURL: getEnv("DATABASE_URL", "photo_library.db"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
