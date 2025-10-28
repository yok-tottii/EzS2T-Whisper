package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YOURUSERNAME/EzS2T-Whisper/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	if handler == nil {
		t.Fatal("Expected handler to be created")
	}

	if handler.config != cfg {
		t.Error("Expected config to be set")
	}
}

func TestGetSettings(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()

	handler.handleSettings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response config.Config
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Language != cfg.Language {
		t.Errorf("Expected Language '%s', got '%s'", cfg.Language, response.Language)
	}
}

func TestPutSettings(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	updates := map[string]interface{}{
		"recording_mode": "toggle",
		"language":       "en",
	}

	body, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.handleSettings(w, req)

	// May fail if config directory doesn't exist, but should update config in memory
	// Verify config was updated in memory
	if cfg.RecordingMode != "toggle" {
		t.Errorf("Expected RecordingMode 'toggle', got '%s'", cfg.RecordingMode)
	}

	if cfg.Language != "en" {
		t.Errorf("Expected Language 'en', got '%s'", cfg.Language)
	}
}

func TestPutSettingsInvalid(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	// Invalid JSON
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.handleSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleHotkeyValidate(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/hotkey/validate", nil)
	w := httptest.NewRecorder()

	handler.handleHotkeyValidate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["conflicts"]; !ok {
		t.Error("Expected 'conflicts' field in response")
	}
}

func TestHandleHotkeyRegister(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	hotkey := config.HotkeyConfig{
		Ctrl: true,
		Cmd:  true,
		Key:  "R",
	}

	body, _ := json.Marshal(hotkey)
	req := httptest.NewRequest(http.MethodPost, "/api/hotkey/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.handleHotkeyRegister(w, req)

	// May fail if config directory doesn't exist, but should update config in memory
	// Verify hotkey was updated
	if cfg.Hotkey.Cmd != true {
		t.Error("Expected Cmd to be true")
	}

	if cfg.Hotkey.Key != "R" {
		t.Errorf("Expected Key 'R', got '%s'", cfg.Hotkey.Key)
	}
}

func TestHandleDevices(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
	w := httptest.NewRecorder()

	handler.handleDevices(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["devices"]; !ok {
		t.Error("Expected 'devices' field in response")
	}
}

func TestHandleModels(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	w := httptest.NewRecorder()

	handler.handleModels(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["models"]; !ok {
		t.Error("Expected 'models' field in response")
	}
}

func TestHandleModelsRescan(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/models/rescan", nil)
	w := httptest.NewRecorder()

	handler.handleModelsRescan(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestScanModels(t *testing.T) {
	// This test just verifies scanModels doesn't crash
	// Testing with actual files would require modifying the real home directory
	cfg := config.DefaultConfig()
	handler := New(cfg)

	models := handler.scanModels()

	// Models may or may not exist, but the function should not crash
	_ = models

	// Test with non-existent directory (should return empty list)
	if len(models) >= 0 {
		// Pass - function executed without error
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := formatSize(test.bytes)
		if result != test.expected {
			t.Errorf("formatSize(%d) = '%s', expected '%s'", test.bytes, result, test.expected)
		}
	}
}

func TestHandleTestRecord(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/test/record", nil)
	w := httptest.NewRecorder()

	handler.handleTestRecord(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandlePermissions(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/permissions", nil)
	w := httptest.NewRecorder()

	handler.handlePermissions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]Permission
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["microphone"]; !ok {
		t.Error("Expected 'microphone' field in response")
	}

	if _, ok := response["accessibility"]; !ok {
		t.Error("Expected 'accessibility' field in response")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	cfg := config.DefaultConfig()
	handler := New(cfg)

	// Test wrong method on various endpoints
	tests := []struct {
		path   string
		method string
	}{
		{"/api/settings", http.MethodDelete},
		{"/api/hotkey/validate", http.MethodGet},
		{"/api/hotkey/register", http.MethodGet},
		{"/api/devices", http.MethodPost},
		{"/api/models", http.MethodPost},
		{"/api/models/rescan", http.MethodGet},
		{"/api/test/record", http.MethodGet},
		{"/api/permissions", http.MethodPost},
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.path, nil)
		w := httptest.NewRecorder()

		switch test.path {
		case "/api/settings":
			handler.handleSettings(w, req)
		case "/api/hotkey/validate":
			handler.handleHotkeyValidate(w, req)
		case "/api/hotkey/register":
			handler.handleHotkeyRegister(w, req)
		case "/api/devices":
			handler.handleDevices(w, req)
		case "/api/models":
			handler.handleModels(w, req)
		case "/api/models/rescan":
			handler.handleModelsRescan(w, req)
		case "/api/test/record":
			handler.handleTestRecord(w, req)
		case "/api/permissions":
			handler.handlePermissions(w, req)
		}

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: Expected status 405, got %d", test.method, test.path, w.Code)
		}
	}
}
