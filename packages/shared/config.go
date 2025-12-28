// Package shared contains shared utilities and helpers.
package shared

import (
	"os"
	"strconv"
)

// Config holds application configuration.
type Config struct {
	// Server
	APIHost string
	APIPort string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// AI Providers
	OpenAIAPIKey    string
	AnthropicAPIKey string
	OllamaURL       string

	// Features
	EnablePlayground bool
	LogLevel         string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		// Server
		APIHost: getEnv("API_HOST", "0.0.0.0"),
		APIPort: getEnv("API_PORT", "8080"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/aiaggregator?sslmode=disable"),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// AI Providers
		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		OllamaURL:       getEnv("OLLAMA_URL", "http://localhost:11434"),

		// Features
		EnablePlayground: getEnvBool("ENABLE_PLAYGROUND", true),
		LogLevel:         getEnv("LOG_LEVEL", "debug"),
	}
}

// getEnv returns environment variable value or default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns environment variable as bool or default.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
