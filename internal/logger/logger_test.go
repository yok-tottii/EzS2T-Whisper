package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != INFO {
		t.Errorf("Expected default level INFO, got %v", config.Level)
	}

	if config.RetentionDays != 7 {
		t.Errorf("Expected retention days 7, got %d", config.RetentionDays)
	}

	if config.LogDir == "" {
		t.Error("Expected non-empty log directory")
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tempDir := t.TempDir()

	config := Config{
		LogDir:        tempDir,
		Level:         INFO,
		RetentionDays: 7,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Check if log file was created
	today := time.Now().Format("20060102")
	filename := fmt.Sprintf("ezs2t-whisper-%s.log", today)
	logPath := filepath.Join(tempDir, filename)

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created: %s", logPath)
	}
}

func TestLogging(t *testing.T) {
	tempDir := t.TempDir()

	config := Config{
		LogDir:        tempDir,
		Level:         DEBUG,
		RetentionDays: 7,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test logging at different levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")

	// Read log file and check contents
	today := time.Now().Format("20060102")
	filename := fmt.Sprintf("ezs2t-whisper-%s.log", today)
	logPath := filepath.Join(tempDir, filename)

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Check if all messages are logged
	if !strings.Contains(logContent, "Debug message") {
		t.Error("Debug message not found in log")
	}
	if !strings.Contains(logContent, "Info message") {
		t.Error("Info message not found in log")
	}
	if !strings.Contains(logContent, "Warn message") {
		t.Error("Warn message not found in log")
	}
	if !strings.Contains(logContent, "Error message") {
		t.Error("Error message not found in log")
	}

	// Check if log levels are included
	if !strings.Contains(logContent, "[DEBUG]") {
		t.Error("[DEBUG] prefix not found in log")
	}
	if !strings.Contains(logContent, "[INFO]") {
		t.Error("[INFO] prefix not found in log")
	}
	if !strings.Contains(logContent, "[WARN]") {
		t.Error("[WARN] prefix not found in log")
	}
	if !strings.Contains(logContent, "[ERROR]") {
		t.Error("[ERROR] prefix not found in log")
	}
}

func TestLogLevel(t *testing.T) {
	tempDir := t.TempDir()

	config := Config{
		LogDir:        tempDir,
		Level:         WARN, // Only WARN and ERROR
		RetentionDays: 7,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Test logging at different levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")

	// Read log file and check contents
	today := time.Now().Format("20060102")
	filename := fmt.Sprintf("ezs2t-whisper-%s.log", today)
	logPath := filepath.Join(tempDir, filename)

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// DEBUG and INFO should not be logged
	if strings.Contains(logContent, "Debug message") {
		t.Error("Debug message should not be logged at WARN level")
	}
	if strings.Contains(logContent, "Info message") {
		t.Error("Info message should not be logged at WARN level")
	}

	// WARN and ERROR should be logged
	if !strings.Contains(logContent, "Warn message") {
		t.Error("Warn message not found in log")
	}
	if !strings.Contains(logContent, "Error message") {
		t.Error("Error message not found in log")
	}
}

func TestSetLevel(t *testing.T) {
	tempDir := t.TempDir()

	config := Config{
		LogDir:        tempDir,
		Level:         INFO,
		RetentionDays: 7,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Check initial level
	if logger.GetLevel() != INFO {
		t.Errorf("Expected initial level INFO, got %v", logger.GetLevel())
	}

	// Set new level
	logger.SetLevel(DEBUG)

	if logger.GetLevel() != DEBUG {
		t.Errorf("Expected level DEBUG, got %v", logger.GetLevel())
	}
}

func TestCleanOldLogs(t *testing.T) {
	tempDir := t.TempDir()

	// Create old log files
	oldDate := time.Now().AddDate(0, 0, -10)
	oldFilename := fmt.Sprintf("ezs2t-whisper-%s.log", oldDate.Format("20060102"))
	oldLogPath := filepath.Join(tempDir, oldFilename)

	if err := os.WriteFile(oldLogPath, []byte("old log"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}

	// Change modification time to 10 days ago
	tenDaysAgo := time.Now().AddDate(0, 0, -10)
	if err := os.Chtimes(oldLogPath, tenDaysAgo, tenDaysAgo); err != nil {
		t.Fatalf("Failed to change file times: %v", err)
	}

	config := Config{
		LogDir:        tempDir,
		Level:         INFO,
		RetentionDays: 7,
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Old log file should be deleted
	if _, err := os.Stat(oldLogPath); !os.IsNotExist(err) {
		t.Error("Old log file should have been deleted")
	}

	// Current log file should exist
	today := time.Now().Format("20060102")
	currentFilename := fmt.Sprintf("ezs2t-whisper-%s.log", today)
	currentLogPath := filepath.Join(tempDir, currentFilename)

	if _, err := os.Stat(currentLogPath); os.IsNotExist(err) {
		t.Error("Current log file should exist")
	}
}
