package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	// Clean environment
	cleanEnv()

	t.Run("default values", func(t *testing.T) {
		config := NewConfig()

		if got, want := config.GetPerplexityAPIKey(), ""; got != want {
			t.Errorf("GetPerplexityAPIKey() = %q, want %q", got, want)
		}

		if got, want := config.GetDefaultModel(), "sonar"; got != want {
			t.Errorf("GetDefaultModel() = %q, want %q", got, want)
		}

		if got, want := config.GetRequestTimeout(), 30; got != want {
			t.Errorf("GetRequestTimeout() = %d, want %d", got, want)
		}

		if got, want := config.GetLogLevel(), "info"; got != want {
			t.Errorf("GetLogLevel() = %q, want %q", got, want)
		}
	})

	t.Run("environment variables", func(t *testing.T) {
		// Set environment variables
		os.Setenv("PERPLEXITY_API_KEY", "test-api-key")
		os.Setenv("PERPLEXITY_DEFAULT_MODEL", "custom-model")
		os.Setenv("REQUEST_TIMEOUT_SECONDS", "60")
		os.Setenv("LOG_LEVEL", "DEBUG")
		defer cleanEnv()

		config := NewConfig()

		if got, want := config.GetPerplexityAPIKey(), "test-api-key"; got != want {
			t.Errorf("GetPerplexityAPIKey() = %q, want %q", got, want)
		}

		if got, want := config.GetDefaultModel(), "custom-model"; got != want {
			t.Errorf("GetDefaultModel() = %q, want %q", got, want)
		}

		if got, want := config.GetRequestTimeout(), 60; got != want {
			t.Errorf("GetRequestTimeout() = %d, want %d", got, want)
		}

		if got, want := config.GetLogLevel(), "debug"; got != want {
			t.Errorf("GetLogLevel() = %q, want %q", got, want)
		}
	})

	t.Run("invalid timeout falls back to default", func(t *testing.T) {
		os.Setenv("REQUEST_TIMEOUT_SECONDS", "invalid")
		defer cleanEnv()

		config := NewConfig()
		if got, want := config.GetRequestTimeout(), 30; got != want {
			t.Errorf("GetRequestTimeout() = %d, want %d", got, want)
		}
	})

	t.Run("negative timeout falls back to default", func(t *testing.T) {
		os.Setenv("REQUEST_TIMEOUT_SECONDS", "-5")
		defer cleanEnv()

		config := NewConfig()
		if got, want := config.GetRequestTimeout(), 30; got != want {
			t.Errorf("GetRequestTimeout() = %d, want %d", got, want)
		}
	})

	t.Run("log level case normalization", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"DEBUG", "debug"},
			{"Info", "info"},
			{"WARN", "warn"},
			{"Error", "error"},
			{"MiXeD", "mixed"},
		}

		for _, tc := range testCases {
			os.Setenv("LOG_LEVEL", tc.input)
			config := NewConfig()
			if got := config.GetLogLevel(); got != tc.expected {
				t.Errorf("LOG_LEVEL=%q: got %q, want %q", tc.input, got, tc.expected)
			}
		}
		cleanEnv()
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := &Config{
			perplexityAPIKey: "test-key",
			defaultModel:     "test-model",
			requestTimeout:   30,
			logLevel:         "info",
		}

		if err := config.Validate(); err != nil {
			t.Errorf("Validate() returned error for valid config: %v", err)
		}
	})

	t.Run("missing API key", func(t *testing.T) {
		config := &Config{
			perplexityAPIKey: "",
			defaultModel:     "test-model",
			requestTimeout:   30,
			logLevel:         "info",
		}

		if err := config.Validate(); err != ErrMissingAPIKey {
			t.Errorf("Validate() = %v, want %v", err, ErrMissingAPIKey)
		}
	})

	t.Run("empty model", func(t *testing.T) {
		config := &Config{
			perplexityAPIKey: "test-key",
			defaultModel:     "",
			requestTimeout:   30,
			logLevel:         "info",
		}

		if err := config.Validate(); err != ErrInvalidModel {
			t.Errorf("Validate() = %v, want %v", err, ErrInvalidModel)
		}
	})

	t.Run("invalid timeout", func(t *testing.T) {
		config := &Config{
			perplexityAPIKey: "test-key",
			defaultModel:     "test-model",
			requestTimeout:   0,
			logLevel:         "info",
		}

		if err := config.Validate(); err != ErrInvalidTimeout {
			t.Errorf("Validate() = %v, want %v", err, ErrInvalidTimeout)
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		config := &Config{
			perplexityAPIKey: "test-key",
			defaultModel:     "test-model",
			requestTimeout:   30,
			logLevel:         "invalid",
		}

		if err := config.Validate(); err != ErrInvalidLogLevel {
			t.Errorf("Validate() = %v, want %v", err, ErrInvalidLogLevel)
		}
	})

	t.Run("valid log levels", func(t *testing.T) {
		validLevels := []string{"debug", "info", "warn", "error"}

		for _, level := range validLevels {
			config := &Config{
				perplexityAPIKey: "test-key",
				defaultModel:     "test-model",
				requestTimeout:   30,
				logLevel:         level,
			}

			if err := config.Validate(); err != nil {
				t.Errorf("Validate() returned error for valid log level %q: %v", level, err)
			}
		}
	})
}

func TestGetEnvString(t *testing.T) {
	t.Run("existing environment variable", func(t *testing.T) {
		key := "TEST_STRING_VAR"
		value := "test-value"
		os.Setenv(key, value)
		defer os.Unsetenv(key)

		result := getEnvString(key, "default")
		if result != value {
			t.Errorf("getEnvString() = %q, want %q", result, value)
		}
	})

	t.Run("missing environment variable", func(t *testing.T) {
		key := "NON_EXISTENT_VAR"
		defaultValue := "default-value"

		result := getEnvString(key, defaultValue)
		if result != defaultValue {
			t.Errorf("getEnvString() = %q, want %q", result, defaultValue)
		}
	})

	t.Run("empty environment variable", func(t *testing.T) {
		key := "EMPTY_VAR"
		defaultValue := "default-value"
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		result := getEnvString(key, defaultValue)
		if result != defaultValue {
			t.Errorf("getEnvString() = %q, want %q", result, defaultValue)
		}
	})
}

func TestGetEnvInt(t *testing.T) {
	t.Run("valid integer", func(t *testing.T) {
		key := "TEST_INT_VAR"
		value := "42"
		os.Setenv(key, value)
		defer os.Unsetenv(key)

		result := getEnvInt(key, 10)
		if result != 42 {
			t.Errorf("getEnvInt() = %d, want %d", result, 42)
		}
	})

	t.Run("invalid integer", func(t *testing.T) {
		key := "INVALID_INT_VAR"
		os.Setenv(key, "not-a-number")
		defer os.Unsetenv(key)

		defaultValue := 10
		result := getEnvInt(key, defaultValue)
		if result != defaultValue {
			t.Errorf("getEnvInt() = %d, want %d", result, defaultValue)
		}
	})

	t.Run("negative integer", func(t *testing.T) {
		key := "NEGATIVE_INT_VAR"
		os.Setenv(key, "-5")
		defer os.Unsetenv(key)

		defaultValue := 10
		result := getEnvInt(key, defaultValue)
		if result != defaultValue {
			t.Errorf("getEnvInt() = %d, want %d", result, defaultValue)
		}
	})

	t.Run("zero integer", func(t *testing.T) {
		key := "ZERO_INT_VAR"
		os.Setenv(key, "0")
		defer os.Unsetenv(key)

		defaultValue := 10
		result := getEnvInt(key, defaultValue)
		if result != defaultValue {
			t.Errorf("getEnvInt() = %d, want %d", result, defaultValue)
		}
	})

	t.Run("missing environment variable", func(t *testing.T) {
		key := "NON_EXISTENT_INT_VAR"
		defaultValue := 10

		result := getEnvInt(key, defaultValue)
		if result != defaultValue {
			t.Errorf("getEnvInt() = %d, want %d", result, defaultValue)
		}
	})
}

// cleanEnv cleans up test environment variables
func cleanEnv() {
	testVars := []string{
		"PERPLEXITY_API_KEY",
		"PERPLEXITY_DEFAULT_MODEL",
		"REQUEST_TIMEOUT_SECONDS",
		"LOG_LEVEL",
		"TEST_STRING_VAR",
		"TEST_INT_VAR",
		"INVALID_INT_VAR",
		"NEGATIVE_INT_VAR",
		"ZERO_INT_VAR",
		"EMPTY_VAR",
	}

	for _, v := range testVars {
		os.Unsetenv(v)
	}
}
