package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger to implement the domain.Logger interface
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a new structured logger with the specified log level and output
func NewLogger(level string, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	// Parse log level
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Create handler with JSON format for structured logging
	opts := &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Sanitize sensitive information
			if isSensitiveKey(a.Key) {
				return slog.String(a.Key, "[REDACTED]")
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(output, opts)
	logger := slog.New(handler)

	return &Logger{
		logger: logger,
	}
}

// Info logs an info level message with optional fields
func (l *Logger) Info(msg string, fields ...any) {
	l.logger.Info(msg, fields...)
}

// Error logs an error level message with optional fields
func (l *Logger) Error(msg string, fields ...any) {
	l.logger.Error(msg, fields...)
}

// Debug logs a debug level message with optional fields
func (l *Logger) Debug(msg string, fields ...any) {
	l.logger.Debug(msg, fields...)
}

// Warn logs a warning level message with optional fields
func (l *Logger) Warn(msg string, fields ...any) {
	l.logger.Warn(msg, fields...)
}

// WithContext returns a logger with context information
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: l.logger,
		ctx:    ctx,
	}
}

// ContextLogger wraps a logger with context
type ContextLogger struct {
	logger *slog.Logger
	ctx    context.Context
}

// Info logs an info level message with context
func (cl *ContextLogger) Info(msg string, fields ...any) {
	cl.logger.InfoContext(cl.ctx, msg, fields...)
}

// Error logs an error level message with context
func (cl *ContextLogger) Error(msg string, fields ...any) {
	cl.logger.ErrorContext(cl.ctx, msg, fields...)
}

// Debug logs a debug level message with context
func (cl *ContextLogger) Debug(msg string, fields ...any) {
	cl.logger.DebugContext(cl.ctx, msg, fields...)
}

// Warn logs a warning level message with context
func (cl *ContextLogger) Warn(msg string, fields ...any) {
	cl.logger.WarnContext(cl.ctx, msg, fields...)
}

// isSensitiveKey checks if a log key contains sensitive information
func isSensitiveKey(key string) bool {
	sensitivePatterns := []string{
		"password", "passwd", "pwd",
		"token", "api_key", "apikey", "api-key",
		"secret", "auth", "authorization",
		"private_key", "privatekey", "private-key",
		"certificate", "cert",
		"session", "cookie",
	}

	lowerKey := strings.ToLower(key)

	// Check for exact matches or patterns ending with _key
	if strings.HasSuffix(lowerKey, "_key") || lowerKey == "key" {
		return true
	}

	for _, sensitive := range sensitivePatterns {
		if strings.Contains(lowerKey, sensitive) {
			return true
		}
	}
	return false
}
