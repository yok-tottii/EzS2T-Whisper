package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents the logging level
type Level int

const (
	// DEBUG level for detailed debugging information
	DEBUG Level = iota
	// INFO level for informational messages
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger handles logging to file with rotation
type Logger struct {
	mu            sync.RWMutex
	level         Level
	file          *os.File
	infoLog       *log.Logger
	warnLog       *log.Logger
	errorLog      *log.Logger
	debugLog      *log.Logger
	logDir        string
	currentDay    string
	retentionDays int
}

// Config holds logger configuration
type Config struct {
	LogDir        string
	Level         Level
	RetentionDays int
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	logDir := filepath.Join(homeDir, "Library", "Application Support", "EzS2T-Whisper", "logs")

	return Config{
		LogDir:        logDir,
		Level:         INFO,
		RetentionDays: 7,
	}
}

// New creates a new logger
func New(config Config) (*Logger, error) {
	l := &Logger{
		level:         config.Level,
		logDir:        config.LogDir,
		retentionDays: config.RetentionDays,
	}

	if err := l.rotateLog(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return l, nil
}

// rotateLog rotates the log file if necessary
func (l *Logger) rotateLog() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	today := time.Now().Format("20060102")

	// Check if we need to rotate (new day)
	if l.currentDay == today && l.file != nil {
		return nil
	}

	// Close existing file
	if l.file != nil {
		l.file.Close()
	}

	// Create log directory if not exists
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create new log file
	filename := fmt.Sprintf("ezs2t-whisper-%s.log", today)
	filePath := filepath.Join(l.logDir, filename)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	l.currentDay = today

	// Create loggers
	l.infoLog = log.New(file, "[INFO] ", log.LstdFlags)
	l.warnLog = log.New(file, "[WARN] ", log.LstdFlags)
	l.errorLog = log.New(file, "[ERROR] ", log.LstdFlags)
	l.debugLog = log.New(file, "[DEBUG] ", log.LstdFlags)

	// Clean old logs
	if err := l.cleanOldLogs(); err != nil {
		// Log error but don't fail
		l.Warn("Failed to clean old logs: %v", err)
	}

	return nil
}

// cleanOldLogs deletes log files older than retentionDays
func (l *Logger) cleanOldLogs() error {
	cutoffDate := time.Now().AddDate(0, 0, -l.retentionDays)

	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a log file with the expected pattern
		if filepath.Ext(entry.Name()) != ".log" {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Delete if older than cutoff date
		if info.ModTime().Before(cutoffDate) {
			filePath := filepath.Join(l.logDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				// Continue even if we can't delete a file
				continue
			}
		}
	}

	return nil
}

// checkRotation checks if log rotation is needed and performs it
func (l *Logger) checkRotation() {
	l.mu.RLock()
	currentDay := l.currentDay
	l.mu.RUnlock()

	today := time.Now().Format("20060102")
	if currentDay != today {
		if err := l.rotateLog(); err != nil {
			// Can't log this error since logging is failing
			fmt.Fprintf(os.Stderr, "Failed to rotate log: %v\n", err)
		}
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= DEBUG {
		l.checkRotation()
		l.mu.RLock()
		debugLog := l.debugLog
		l.mu.RUnlock()
		if debugLog != nil {
			debugLog.Printf(format, v...)
		}
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= INFO {
		l.checkRotation()
		l.mu.RLock()
		infoLog := l.infoLog
		l.mu.RUnlock()
		if infoLog != nil {
			infoLog.Printf(format, v...)
		}
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= WARN {
		l.checkRotation()
		l.mu.RLock()
		warnLog := l.warnLog
		l.mu.RUnlock()
		if warnLog != nil {
			warnLog.Printf(format, v...)
		}
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.mu.RLock()
	level := l.level
	l.mu.RUnlock()

	if level <= ERROR {
		l.checkRotation()
		l.mu.RLock()
		errorLog := l.errorLog
		l.mu.RUnlock()
		if errorLog != nil {
			errorLog.Printf(format, v...)
		}
	}
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.level
}
