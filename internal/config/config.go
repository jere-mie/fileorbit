package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port          string
	Host          string
	AdminPassword string
	SessionSecret string
	DatabasePath  string
	MaxFileSize   int64
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		Host:          getEnv("HOST", "localhost"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin"),
		SessionSecret: getEnv("SESSION_SECRET", "change-me-to-a-random-secret"),
		DatabasePath:  getEnv("DATABASE_PATH", "fileorbit.db"),
	}

	maxSize := getEnv("MAX_FILE_SIZE", "104857600")
	size, err := strconv.ParseInt(maxSize, 10, 64)
	if err != nil {
		size = 104857600 // 100MB default
	}
	cfg.MaxFileSize = size

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
