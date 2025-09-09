package internal

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	PerplexityAPIKey string
	DefaultModel     string
	RequestTimeout   time.Duration
	LogLevel         string
}

func NewConfig() (*Config, error) {
	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PERPLEXITY_API_KEY environment variable is required")
	}

	config := &Config{
		PerplexityAPIKey: apiKey,
		DefaultModel:     getEnvWithDefault("PERPLEXITY_DEFAULT_MODEL", "sonar"),
		RequestTimeout:   30 * time.Second,
		LogLevel:         getEnvWithDefault("LOG_LEVEL", "INFO"),
	}

	if timeoutStr := os.Getenv("REQUEST_TIMEOUT"); timeoutStr != "" {
		if timeoutSec, err := strconv.Atoi(timeoutStr); err == nil && timeoutSec > 0 {
			config.RequestTimeout = time.Duration(timeoutSec) * time.Second
		}
	}

	return config, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.PerplexityAPIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}
	return nil
}
