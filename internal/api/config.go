package api

import (
	"os"
	"strconv"
)

// Config holds the configuration for the API server
type Config struct {
	Port        string
	Host        string
	Address     string
	APIKey      string
	RateLimit   int
	Environment string
	LogLevel    string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	port := getEnv("PORT", "8080")
	host := getEnv("HOST", "0.0.0.0")
	
	config := &Config{
		Port:        port,
		Host:        host,
		Address:     host + ":" + port,
		APIKey:      getEnv("API_KEY", ""),
		RateLimit:   getEnvInt("RATE_LIMIT", 100),
		Environment: getEnv("ENVIRONMENT", "production"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	return config
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}