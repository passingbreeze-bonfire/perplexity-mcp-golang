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
	HTTPHost         string
	HTTPPort         string
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
		HTTPHost:         getEnvWithDefault("HTTP_HOST", "0.0.0.0"),
		HTTPPort:         getEnvWithDefault("HTTP_PORT", "8080"),
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
	if c.HTTPHost == "" {
		return fmt.Errorf("HTTP host is required")
	}
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP port is required")
	}
	return nil
}
