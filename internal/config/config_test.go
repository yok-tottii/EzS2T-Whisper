package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("Expected default config to be created")
	}

	if config.Hotkey.Ctrl != true {
		t.Error("Expected Ctrl to be true")
	}

	if config.Hotkey.Alt != true {
		t.Error("Expected Alt to be true")
	}

	if config.Hotkey.Key != "Space" {
		t.Errorf("Expected Key to be 'Space', got '%s'", config.Hotkey.Key)
	}

	if config.RecordingMode != "press-to-hold" {
		t.Errorf("Expected RecordingMode 'press-to-hold', got '%s'", config.RecordingMode)
	}

	if config.Language != "ja" {
		t.Errorf("Expected Language 'ja', got '%s'", config.Language)
	}

	if config.UILanguage != "ja" {
		t.Errorf("Expected UILanguage 'ja', got '%s'", config.UILanguage)
	}

	if config.MaxRecordTime != 60 {
		t.Errorf("Expected MaxRecordTime 60, got %d", config.MaxRecordTime)
	}

	if config.PasteSplitSize != 500 {
		t.Errorf("Expected PasteSplitSize 500, got %d", config.PasteSplitSize)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create config
	config := DefaultConfig()
	config.RecordingMode = "toggle"
	config.Language = "en"

	// Save config
	if err := config.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config
	if loaded.RecordingMode != "toggle" {
		t.Errorf("Expected RecordingMode 'toggle', got '%s'", loaded.RecordingMode)
	}

	if loaded.Language != "en" {
		t.Errorf("Expected Language 'en', got '%s'", loaded.Language)
	}
}

func TestLoadNonexistent(t *testing.T) {
	// Load from nonexistent path should return default config
	config, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("Expected no error when loading nonexistent file, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected default config to be returned")
	}

	// Should match default config
	defaultConfig := DefaultConfig()
	if config.Language != defaultConfig.Language {
		t.Errorf("Expected Language '%s', got '%s'", defaultConfig.Language, config.Language)
	}
}

func TestUpdate(t *testing.T) {
	config := DefaultConfig()

	updates := map[string]interface{}{
		"recording_mode":  "toggle",
		"language":        "en",
		"audio_device_id": float64(1),
		"max_record_time": float64(90),
	}

	if err := config.Update(updates); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	if config.RecordingMode != "toggle" {
		t.Errorf("Expected RecordingMode 'toggle', got '%s'", config.RecordingMode)
	}

	if config.Language != "en" {
		t.Errorf("Expected Language 'en', got '%s'", config.Language)
	}

	if config.AudioDeviceID != 1 {
		t.Errorf("Expected AudioDeviceID 1, got %d", config.AudioDeviceID)
	}

	if config.MaxRecordTime != 90 {
		t.Errorf("Expected MaxRecordTime 90, got %d", config.MaxRecordTime)
	}
}

func TestUpdateInvalidValues(t *testing.T) {
	config := DefaultConfig()

	// Test invalid recording_mode
	updates := map[string]interface{}{
		"recording_mode": "invalid",
	}

	if err := config.Update(updates); err == nil {
		t.Error("Expected error for invalid recording_mode")
	}

	// Test invalid language
	updates = map[string]interface{}{
		"language": "invalid",
	}

	if err := config.Update(updates); err == nil {
		t.Error("Expected error for invalid language")
	}

	// Test invalid ui_language
	updates = map[string]interface{}{
		"ui_language": "invalid",
	}

	if err := config.Update(updates); err == nil {
		t.Error("Expected error for invalid ui_language")
	}
}

func TestClone(t *testing.T) {
	original := DefaultConfig()
	original.RecordingMode = "toggle"
	original.Language = "en"

	cloned := original.Clone()

	// Verify values match
	if cloned.RecordingMode != original.RecordingMode {
		t.Errorf("Expected RecordingMode '%s', got '%s'", original.RecordingMode, cloned.RecordingMode)
	}

	if cloned.Language != original.Language {
		t.Errorf("Expected Language '%s', got '%s'", original.Language, cloned.Language)
	}

	// Modify clone and verify original is unaffected
	cloned.Language = "ja"

	if original.Language != "en" {
		t.Error("Modifying clone affected original")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()

	if path == "" {
		t.Error("Expected non-empty config path")
	}

	// Should contain expected components
	expectedDir := filepath.Join("Library", "Application Support", "EzS2T-Whisper")
	if !contains(path, expectedDir) {
		t.Errorf("Expected path to contain '%s', got '%s'", expectedDir, path)
	}

	if !contains(path, "config.json") {
		t.Errorf("Expected path to contain 'config.json', got '%s'", path)
	}
}

func TestHotkeyConfig(t *testing.T) {
	config := DefaultConfig()

	// Test default hotkey
	if config.Hotkey.Ctrl != true {
		t.Error("Expected Ctrl to be true")
	}

	if config.Hotkey.Alt != true {
		t.Error("Expected Alt to be true")
	}

	if config.Hotkey.Shift != false {
		t.Error("Expected Shift to be false")
	}

	if config.Hotkey.Cmd != false {
		t.Error("Expected Cmd to be false")
	}

	if config.Hotkey.Key != "Space" {
		t.Errorf("Expected Key 'Space', got '%s'", config.Hotkey.Key)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
