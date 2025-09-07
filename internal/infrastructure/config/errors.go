package config

import "errors"

var (
	// ErrMissingAPIKey is returned when the Perplexity API key is not configured
	ErrMissingAPIKey = errors.New("PERPLEXITY_API_KEY environment variable is required")

	// ErrInvalidModel is returned when the model configuration is invalid
	ErrInvalidModel = errors.New("invalid or empty model configuration")

	// ErrInvalidTimeout is returned when the timeout configuration is invalid
	ErrInvalidTimeout = errors.New("request timeout must be greater than 0")

	// ErrInvalidLogLevel is returned when the log level is invalid
	ErrInvalidLogLevel = errors.New("log level must be one of: debug, info, warn, error")
)
