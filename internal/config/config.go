package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config holds application configuration
type Config struct {
	Hotkey        HotkeyConfig `json:"hotkey"`
	RecordingMode string       `json:"recording_mode"` // "press-to-hold" or "toggle"
	ModelPath     string       `json:"model_path"`
	Language      string       `json:"language"` // "auto" for automatic detection, or specific language code
	AudioDeviceID int          `json:"audio_device_id"`
	UILanguage    string       `json:"ui_language"` // "ja" or "en"
	MaxRecordTime int          `json:"max_record_time"` // seconds
	PasteSplitSize int         `json:"paste_split_size"` // characters
	mu            sync.RWMutex
}

// HotkeyConfig holds hotkey configuration
type HotkeyConfig struct {
	Ctrl   bool   `json:"ctrl"`
	Shift  bool   `json:"shift"`
	Alt    bool   `json:"alt"`
	Cmd    bool   `json:"cmd"`
	Key    string `json:"key"` // e.g., "Space"
}

// IsValidModelExtension checks if the file has a valid Whisper model extension
// Supports both .bin (current official format) and .gguf (future format)
func IsValidModelExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".bin" || ext == ".gguf"
}

// GetRecommendedModelName returns the recommended model filename
func GetRecommendedModelName() string {
	return "ggml-large-v3-turbo-q5_0.bin"
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Hotkey: HotkeyConfig{
			Ctrl: true,
			Alt:  true,
			Key:  "Space",
		},
		RecordingMode:  "press-to-hold",
		ModelPath:      "", // Empty by default - user must specify
		Language:       "auto", // Automatic language detection
		AudioDeviceID:  -1, // -1 means use system default device
		UILanguage:     "ja",
		MaxRecordTime:  60, // 60 seconds
		PasteSplitSize: 500, // 500 characters
	}
}

// Load loads configuration from the specified path
func Load(path string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// ホットキー設定の検証と修正
	if config.Hotkey.Key == "" {
		config.Hotkey.Key = "Space" // デフォルト値で補完
	}

	return &config, nil
}

// Save saves configuration to the specified path
func (c *Config) Save(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library", "Application Support", "EzS2T-Whisper", "config.json")
}

// Update updates configuration fields
func (c *Config) Update(updates map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Apply updates
	for key, value := range updates {
		switch key {
		case "recording_mode":
			if v, ok := value.(string); ok {
				if v != "press-to-hold" && v != "toggle" {
					return fmt.Errorf("invalid recording_mode: %s", v)
				}
				c.RecordingMode = v
			}
		case "model_path":
			if v, ok := value.(string); ok {
				c.ModelPath = v
			}
		case "language":
			if v, ok := value.(string); ok {
				// Allow any language code - Whisper.cpp supports 100+ languages
				// "auto" enables automatic language detection
				c.Language = v
			}
		case "audio_device_id":
			if v, ok := value.(float64); ok {
				c.AudioDeviceID = int(v)
			}
		case "ui_language":
			if v, ok := value.(string); ok {
				if v != "ja" && v != "en" {
					return fmt.Errorf("invalid ui_language: %s", v)
				}
				c.UILanguage = v
			}
		case "max_record_time":
			if v, ok := value.(float64); ok {
				c.MaxRecordTime = int(v)
			}
		case "paste_split_size":
			if v, ok := value.(float64); ok {
				c.PasteSplitSize = int(v)
			}
		case "hotkey":
			if v, ok := value.(map[string]interface{}); ok {
				// HotkeyConfigの各フィールドを更新
				if ctrl, ok := v["ctrl"].(bool); ok {
					c.Hotkey.Ctrl = ctrl
				}
				if shift, ok := v["shift"].(bool); ok {
					c.Hotkey.Shift = shift
				}
				if alt, ok := v["alt"].(bool); ok {
					c.Hotkey.Alt = alt
				}
				if cmd, ok := v["cmd"].(bool); ok {
					c.Hotkey.Cmd = cmd
				}
				if key, ok := v["key"].(string); ok {
					c.Hotkey.Key = key
				}
			}
		}
	}

	return nil
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &Config{
		Hotkey:         c.Hotkey,
		RecordingMode:  c.RecordingMode,
		ModelPath:      c.ModelPath,
		Language:       c.Language,
		AudioDeviceID:  c.AudioDeviceID,
		UILanguage:     c.UILanguage,
		MaxRecordTime:  c.MaxRecordTime,
		PasteSplitSize: c.PasteSplitSize,
	}
}

// ExpandPath expands ~ to home directory in file paths
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(homeDir, path[2:]), nil
	}

	// Return absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// GetModelPath returns the expanded model path
func (c *Config) GetModelPath() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ExpandPath(c.ModelPath)
}

// ValidateModelPath validates the model file path
func (c *Config) ValidateModelPath() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.ModelPath == "" {
		return fmt.Errorf("model path is not set")
	}

	expandedPath, err := ExpandPath(c.ModelPath)
	if err != nil {
		return fmt.Errorf("failed to expand model path: %w", err)
	}

	// Check if file exists
	info, err := os.Stat(expandedPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", expandedPath)
	}
	if err != nil {
		return fmt.Errorf("failed to check model file: %w", err)
	}

	// Check if it's a regular file
	if info.IsDir() {
		return fmt.Errorf("model path is a directory, not a file: %s", expandedPath)
	}

	// Check file extension
	if !IsValidModelExtension(expandedPath) {
		return fmt.Errorf("model file must have .bin or .gguf extension: %s", expandedPath)
	}

	return nil
}

// Validate validates all configuration fields
func (c *Config) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Validate recording mode
	if c.RecordingMode != "press-to-hold" && c.RecordingMode != "toggle" {
		return fmt.Errorf("invalid recording_mode: %s (must be 'press-to-hold' or 'toggle')", c.RecordingMode)
	}

	// Validate language (allow any non-empty value - Whisper.cpp supports 100+ languages)
	// "auto" enables automatic language detection
	if c.Language == "" {
		return fmt.Errorf("language cannot be empty")
	}

	// Validate UI language
	if c.UILanguage != "ja" && c.UILanguage != "en" {
		return fmt.Errorf("invalid ui_language: %s (must be 'ja' or 'en')", c.UILanguage)
	}

	// Validate max record time
	if c.MaxRecordTime <= 0 || c.MaxRecordTime > 300 {
		return fmt.Errorf("invalid max_record_time: %d (must be between 1 and 300 seconds)", c.MaxRecordTime)
	}

	// Validate paste split size
	if c.PasteSplitSize <= 0 || c.PasteSplitSize > 10000 {
		return fmt.Errorf("invalid paste_split_size: %d (must be between 1 and 10000 characters)", c.PasteSplitSize)
	}

	// Model path validation is optional (can be empty for first run)
	// Use ValidateModelPath() separately when model path is required

	return nil
}
