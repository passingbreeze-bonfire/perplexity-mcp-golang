package mocks

import (
	"fmt"
	"sync"
)

// MockConfigProvider provides a mock implementation of domain.ConfigProvider
// for testing with configurable values
type MockConfigProvider struct {
	mu             sync.RWMutex
	perplexityAPIKey string
	defaultModel     string
	requestTimeout   int
	logLevel         string
}

// NewMockConfig creates a new mock config with sensible defaults
func NewMockConfig() *MockConfigProvider {
	return &MockConfigProvider{
		perplexityAPIKey: "mock-api-key-for-testing",
		defaultModel:     "llama-3.1-sonar-small-128k-online",
		requestTimeout:   30,
		logLevel:         "info",
	}
}

// GetPerplexityAPIKey returns the configured API key
func (m *MockConfigProvider) GetPerplexityAPIKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.perplexityAPIKey
}

// GetDefaultModel returns the configured default model
func (m *MockConfigProvider) GetDefaultModel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultModel
}

// GetRequestTimeout returns the configured request timeout in seconds
func (m *MockConfigProvider) GetRequestTimeout() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requestTimeout
}

// GetLogLevel returns the configured log level
func (m *MockConfigProvider) GetLogLevel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logLevel
}

// Configuration methods for testing

// SetPerplexityAPIKey sets the API key for testing
func (m *MockConfigProvider) SetPerplexityAPIKey(apiKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.perplexityAPIKey = apiKey
}

// SetDefaultModel sets the default model for testing
func (m *MockConfigProvider) SetDefaultModel(model string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultModel = model
}

// SetRequestTimeout sets the request timeout for testing
func (m *MockConfigProvider) SetRequestTimeout(timeout int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestTimeout = timeout
}

// SetLogLevel sets the log level for testing
func (m *MockConfigProvider) SetLogLevel(level string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logLevel = level
}

// Validate provides mock validation that can be configured for testing
func (m *MockConfigProvider) Validate() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.perplexityAPIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if m.requestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[m.logLevel] {
		return fmt.Errorf("invalid log level: %s", m.logLevel)
	}
	return nil
}