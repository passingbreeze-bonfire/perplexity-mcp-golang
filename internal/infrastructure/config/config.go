package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration values
type Config struct {
	perplexityAPIKey string
	defaultModel     string
	requestTimeout   int
	logLevel         string
}

// NewConfig creates a new configuration instance from environment variables
func NewConfig() *Config {
	cfg := &Config{
		perplexityAPIKey: getEnvString("PERPLEXITY_API_KEY", ""),
		defaultModel:     getEnvString("PERPLEXITY_DEFAULT_MODEL", "sonar"),
		requestTimeout:   getEnvInt("REQUEST_TIMEOUT_SECONDS", 30),
		logLevel:         getEnvString("LOG_LEVEL", "info"),
	}

	// Normalize log level to lowercase
	cfg.logLevel = strings.ToLower(cfg.logLevel)

	return cfg
}

// GetPerplexityAPIKey returns the Perplexity API key
func (c *Config) GetPerplexityAPIKey() string {
	return c.perplexityAPIKey
}

// GetDefaultModel returns the default model to use for requests
func (c *Config) GetDefaultModel() string {
	return c.defaultModel
}

// GetRequestTimeout returns the request timeout in seconds
func (c *Config) GetRequestTimeout() int {
	return c.requestTimeout
}

// GetLogLevel returns the log level
func (c *Config) GetLogLevel() string {
	return c.logLevel
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.perplexityAPIKey == "" {
		return ErrMissingAPIKey
	}

	if c.defaultModel == "" {
		return ErrInvalidModel
	}

	if c.requestTimeout <= 0 {
		return ErrInvalidTimeout
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[c.logLevel] {
		return ErrInvalidLogLevel
	}

	return nil
}

// getEnvString gets a string environment variable with a default value
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil && intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}
