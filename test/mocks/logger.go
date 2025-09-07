package mocks

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// LogEntry represents a single log entry for verification
type LogEntry struct {
	Level     string
	Message   string
	Fields    map[string]any
	Timestamp time.Time
}

// MockLogger provides a mock implementation of domain.Logger
// for testing with log capture and verification capabilities
type MockLogger struct {
	mu      sync.RWMutex
	entries []LogEntry
	level   string
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		entries: make([]LogEntry, 0),
		level:   "info",
	}
}

// Info logs an info message
func (m *MockLogger) Info(msg string, fields ...any) {
	m.log("info", msg, fields...)
}

// Error logs an error message
func (m *MockLogger) Error(msg string, fields ...any) {
	m.log("error", msg, fields...)
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.log("debug", msg, fields...)
}

// Warn logs a warning message
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.log("warn", msg, fields...)
}

// log is the internal logging method
func (m *MockLogger) log(level, msg string, fields ...any) {
	if !m.shouldLog(level) {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Convert variadic fields to map
	fieldMap := make(map[string]any)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok && i+1 < len(fields) {
			fieldMap[key] = fields[i+1]
		}
	}

	entry := LogEntry{
		Level:     level,
		Message:   msg,
		Fields:    fieldMap,
		Timestamp: time.Now(),
	}

	m.entries = append(m.entries, entry)
}

// shouldLog determines if a message should be logged based on level
func (m *MockLogger) shouldLog(level string) bool {
	m.mu.RLock()
	currentLevelStr := m.level
	m.mu.RUnlock()
	
	levelOrder := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}

	currentLevel, exists := levelOrder[currentLevelStr]
	if !exists {
		return true // Log everything if unknown level
	}

	messageLevel, exists := levelOrder[level]
	if !exists {
		return true // Log if unknown level
	}

	return messageLevel >= currentLevel
}

// Testing methods

// SetLevel sets the minimum log level
func (m *MockLogger) SetLevel(level string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.level = level
}

// GetEntries returns all logged entries
func (m *MockLogger) GetEntries() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to avoid data races
	entries := make([]LogEntry, len(m.entries))
	copy(entries, m.entries)
	return entries
}

// GetEntriesByLevel returns entries matching the specified level
func (m *MockLogger) GetEntriesByLevel(level string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var matches []LogEntry
	for _, entry := range m.entries {
		if entry.Level == level {
			matches = append(matches, entry)
		}
	}
	return matches
}

// GetEntriesWithMessage returns entries containing the specified message pattern
func (m *MockLogger) GetEntriesWithMessage(pattern string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var matches []LogEntry
	for _, entry := range m.entries {
		if strings.Contains(entry.Message, pattern) {
			matches = append(matches, entry)
		}
	}
	return matches
}

// GetEntriesWithField returns entries containing the specified field
func (m *MockLogger) GetEntriesWithField(fieldName string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var matches []LogEntry
	for _, entry := range m.entries {
		if _, exists := entry.Fields[fieldName]; exists {
			matches = append(matches, entry)
		}
	}
	return matches
}

// HasEntry checks if an entry exists with the given level and message pattern
func (m *MockLogger) HasEntry(level, messagePattern string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, entry := range m.entries {
		if entry.Level == level && strings.Contains(entry.Message, messagePattern) {
			return true
		}
	}
	return false
}

// HasError checks if any error entries exist
func (m *MockLogger) HasError() bool {
	return len(m.GetEntriesByLevel("error")) > 0
}

// GetErrorCount returns the number of error entries
func (m *MockLogger) GetErrorCount() int {
	return len(m.GetEntriesByLevel("error"))
}

// Clear clears all logged entries
func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = make([]LogEntry, 0)
}

// String returns a string representation of all log entries for debugging
func (m *MockLogger) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.entries) == 0 {
		return "No log entries"
	}

	var builder strings.Builder
	for i, entry := range m.entries {
		builder.WriteString(fmt.Sprintf("[%s] %s: %s", 
			strings.ToUpper(entry.Level), 
			entry.Timestamp.Format("15:04:05"), 
			entry.Message))
		
		if len(entry.Fields) > 0 {
			builder.WriteString(" {")
			first := true
			for k, v := range entry.Fields {
				if !first {
					builder.WriteString(", ")
				}
				builder.WriteString(fmt.Sprintf("%s=%v", k, v))
				first = false
			}
			builder.WriteString("}")
		}
		
		if i < len(m.entries)-1 {
			builder.WriteString("\n")
		}
	}
	
	return builder.String()
}