package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config holds application configuration
type Config struct {
	Hotkey        HotkeyConfig `json:"hotkey"`
	RecordingMode string       `json:"recording_mode"` // "press-to-hold" or "toggle"
	ModelPath     string       `json:"model_path"`
	Language      string       `json:"language"` // "ja" or "en"
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

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	defaultModelPath := filepath.Join(homeDir, "Library", "Application Support", "EzS2T-Whisper", "models", "ggml-large-v3-turbo-q5_0.gguf")

	return &Config{
		Hotkey: HotkeyConfig{
			Ctrl: true,
			Alt:  true,
			Key:  "Space",
		},
		RecordingMode:  "press-to-hold",
		ModelPath:      defaultModelPath,
		Language:       "ja",
		AudioDeviceID:  0, // Default device
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
				if v != "ja" && v != "en" {
					return fmt.Errorf("invalid language: %s", v)
				}
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
