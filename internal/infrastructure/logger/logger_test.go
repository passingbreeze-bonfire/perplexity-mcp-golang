package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestNewLogger(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("info", &buf)

		if logger == nil {
			t.Error("NewLogger() returned nil")
		}

		if logger.logger == nil {
			t.Error("NewLogger() returned logger with nil slog.Logger")
		}
	})

	t.Run("different log levels", func(t *testing.T) {
		levels := []string{"debug", "info", "warn", "error", "invalid"}

		for _, level := range levels {
			var buf bytes.Buffer
			logger := NewLogger(level, &buf)

			if logger == nil {
				t.Errorf("NewLogger(%q) returned nil", level)
			}
		}
	})

	t.Run("nil output defaults to stdout", func(t *testing.T) {
		logger := NewLogger("info", nil)

		if logger == nil {
			t.Error("NewLogger() with nil output returned nil")
		}
	})
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)

	logger.Info("test message", "field", "value")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}

	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", logEntry["level"])
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg 'test message', got %v", logEntry["msg"])
	}

	if logEntry["field"] != "value" {
		t.Errorf("Expected field 'value', got %v", logEntry["field"])
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)

	logger.Error("error message", "error_code", 500)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}

	if logEntry["level"] != "ERROR" {
		t.Errorf("Expected level ERROR, got %v", logEntry["level"])
	}

	if logEntry["msg"] != "error message" {
		t.Errorf("Expected msg 'error message', got %v", logEntry["msg"])
	}

	if logEntry["error_code"] != float64(500) {
		t.Errorf("Expected error_code 500, got %v", logEntry["error_code"])
	}
}

func TestLogger_Debug(t *testing.T) {
	t.Run("debug enabled", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("debug", &buf)

		logger.Debug("debug message", "trace_id", "12345")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log output as JSON: %v", err)
		}

		if logEntry["level"] != "DEBUG" {
			t.Errorf("Expected level DEBUG, got %v", logEntry["level"])
		}

		if logEntry["msg"] != "debug message" {
			t.Errorf("Expected msg 'debug message', got %v", logEntry["msg"])
		}
	})

	t.Run("debug disabled", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger("info", &buf)

		logger.Debug("debug message that should not appear")

		if buf.Len() > 0 {
			t.Error("Debug message was logged when log level is info")
		}
	})
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)

	logger.Warn("warning message", "component", "auth")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}

	if logEntry["level"] != "WARN" {
		t.Errorf("Expected level WARN, got %v", logEntry["level"])
	}

	if logEntry["msg"] != "warning message" {
		t.Errorf("Expected msg 'warning message', got %v", logEntry["msg"])
	}

	if logEntry["component"] != "auth" {
		t.Errorf("Expected component 'auth', got %v", logEntry["component"])
	}
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	contextLogger := logger.WithContext(ctx)

	if contextLogger == nil {
		t.Error("WithContext() returned nil")
	}

	contextLogger.Info("context message", "user_id", "user-456")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}

	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", logEntry["level"])
	}

	if logEntry["msg"] != "context message" {
		t.Errorf("Expected msg 'context message', got %v", logEntry["msg"])
	}
}

func TestContextLogger_Methods(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)
	ctx := context.Background()
	contextLogger := logger.WithContext(ctx)

	tests := []struct {
		name    string
		logFunc func(string, ...any)
		level   string
		message string
	}{
		{"Info", contextLogger.Info, "INFO", "info message"},
		{"Error", contextLogger.Error, "ERROR", "error message"},
		{"Debug", contextLogger.Debug, "DEBUG", "debug message"},
		{"Warn", contextLogger.Warn, "WARN", "warn message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.message, "test_field", "test_value")

			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("Failed to parse log output as JSON: %v", err)
			}

			if logEntry["level"] != tt.level {
				t.Errorf("Expected level %s, got %v", tt.level, logEntry["level"])
			}

			if logEntry["msg"] != tt.message {
				t.Errorf("Expected msg '%s', got %v", tt.message, logEntry["msg"])
			}

			if logEntry["test_field"] != "test_value" {
				t.Errorf("Expected test_field 'test_value', got %v", logEntry["test_field"])
			}
		})
	}
}

func TestSensitiveKeyRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)

	sensitiveFields := map[string]string{
		"password":      "secret123",
		"api_key":       "sk-123456",
		"token":         "jwt-token",
		"secret":        "top-secret",
		"private_key":   "-----BEGIN PRIVATE KEY-----",
		"authorization": "Bearer token",
		"session":       "session-id",
		"API_KEY":       "uppercase-key", // test case sensitivity
		"Password":      "mixed-case",    // test mixed case
	}

	for key, value := range sensitiveFields {
		buf.Reset()
		logger.Info("test message", key, value)

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log output as JSON: %v", err)
		}

		if logEntry[key] != "[REDACTED]" {
			t.Errorf("Sensitive field %q was not redacted. Got: %v", key, logEntry[key])
		}
	}
}

func TestNonSensitiveKeysNotRedacted(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("debug", &buf)

	nonSensitiveFields := map[string]string{
		"username":   "john_doe",
		"user_id":    "12345",
		"request_id": "req-789",
		"component":  "auth",
		"method":     "GET",
		"path":       "/api/users",
		"status":     "200",
	}

	for key, value := range nonSensitiveFields {
		buf.Reset()
		logger.Info("test message", key, value)

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log output as JSON: %v", err)
		}

		if logEntry[key] != value {
			t.Errorf("Non-sensitive field %q was incorrectly modified. Expected: %v, Got: %v", key, value, logEntry[key])
		}
	}
}

func TestIsSensitiveKey(t *testing.T) {
	testCases := []struct {
		key       string
		sensitive bool
	}{
		// Sensitive keys
		{"password", true},
		{"PASSWORD", true},
		{"Password", true},
		{"user_password", true},
		{"api_key", true},
		{"API_KEY", true},
		{"token", true},
		{"access_token", true},
		{"refresh_token", true},
		{"secret", true},
		{"client_secret", true},
		{"private_key", true},
		{"authorization", true},
		{"session", true},
		{"session_id", true},
		{"cookie", true},
		{"certificate", true},
		{"cert", true},
		{"key", true},

		// Non-sensitive keys
		{"username", false},
		{"user_id", false},
		{"email", false},
		{"name", false},
		{"age", false},
		{"status", false},
		{"message", false},
		{"request_id", false},
		{"component", false},
		{"method", false},
		{"path", false},
		{"keychain", false}, // contains "key" but not _key suffix
		{"monkey", false},   // contains "key" but not _key suffix
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			result := isSensitiveKey(tc.key)
			if result != tc.sensitive {
				t.Errorf("isSensitiveKey(%q) = %v, want %v", tc.key, result, tc.sensitive)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	testCases := []struct {
		logLevel     string
		debugVisible bool
		infoVisible  bool
		warnVisible  bool
		errorVisible bool
	}{
		{"debug", true, true, true, true},
		{"info", false, true, true, true},
		{"warn", false, false, true, true},
		{"error", false, false, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.logLevel, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tc.logLevel, &buf)

			// Test debug
			buf.Reset()
			logger.Debug("debug message")
			debugLogged := buf.Len() > 0

			// Test info
			buf.Reset()
			logger.Info("info message")
			infoLogged := buf.Len() > 0

			// Test warn
			buf.Reset()
			logger.Warn("warn message")
			warnLogged := buf.Len() > 0

			// Test error
			buf.Reset()
			logger.Error("error message")
			errorLogged := buf.Len() > 0

			if debugLogged != tc.debugVisible {
				t.Errorf("Debug visibility mismatch for level %s: got %v, want %v", tc.logLevel, debugLogged, tc.debugVisible)
			}
			if infoLogged != tc.infoVisible {
				t.Errorf("Info visibility mismatch for level %s: got %v, want %v", tc.logLevel, infoLogged, tc.infoVisible)
			}
			if warnLogged != tc.warnVisible {
				t.Errorf("Warn visibility mismatch for level %s: got %v, want %v", tc.logLevel, warnLogged, tc.warnVisible)
			}
			if errorLogged != tc.errorVisible {
				t.Errorf("Error visibility mismatch for level %s: got %v, want %v", tc.logLevel, errorLogged, tc.errorVisible)
			}
		})
	}
}
