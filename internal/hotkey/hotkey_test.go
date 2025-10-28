package hotkey

import (
	"testing"
	"time"

	"golang.design/x/hotkey"
)

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}

	config := m.GetConfig()
	if len(config.Modifiers) != 2 {
		t.Errorf("Expected 2 modifiers, got %d", len(config.Modifiers))
	}

	if config.Key != hotkey.KeySpace {
		t.Errorf("Expected KeySpace, got %v", config.Key)
	}

	if config.Mode != PressToHold {
		t.Errorf("Expected PressToHold mode, got %v", config.Mode)
	}
}

func TestCheckConflicts(t *testing.T) {
	tests := []struct {
		name           string
		modifiers      []hotkey.Modifier
		key            hotkey.Key
		expectConflict bool
	}{
		{
			name:           "Spotlight conflict (Cmd+Space)",
			modifiers:      []hotkey.Modifier{hotkey.ModCmd},
			key:            hotkey.KeySpace,
			expectConflict: true,
		},
		{
			name:           "No conflict (Ctrl+Option+Space)",
			modifiers:      []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModOption},
			key:            hotkey.KeySpace,
			expectConflict: false,
		},
		{
			name:           "Force Quit conflict (Cmd+Option+Esc)",
			modifiers:      []hotkey.Modifier{hotkey.ModCmd, hotkey.ModOption},
			key:            hotkey.KeyEscape,
			expectConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflicts := CheckConflicts(tt.modifiers, tt.key)
			hasConflict := len(conflicts) > 0

			if hasConflict != tt.expectConflict {
				t.Errorf("Expected conflict=%v, got conflict=%v (found %d conflicts)",
					tt.expectConflict, hasConflict, len(conflicts))
			}
		})
	}
}

func TestFormatHotkey(t *testing.T) {
	tests := []struct {
		name      string
		modifiers []hotkey.Modifier
		key       hotkey.Key
		expected  string
	}{
		{
			name:      "Ctrl+Option+Space",
			modifiers: []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModOption},
			key:       hotkey.KeySpace,
			expected:  "⌃⌥Space",
		},
		{
			name:      "Cmd+Space",
			modifiers: []hotkey.Modifier{hotkey.ModCmd},
			key:       hotkey.KeySpace,
			expected:  "⌘Space",
		},
		{
			name:      "Cmd+Shift+A",
			modifiers: []hotkey.Modifier{hotkey.ModCmd, hotkey.ModShift},
			key:       hotkey.KeyA,
			expected:  "⌘⇧A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatHotkey(tt.modifiers, tt.key)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHotkeyMatches(t *testing.T) {
	tests := []struct {
		name     string
		mods1    []hotkey.Modifier
		key1     hotkey.Key
		mods2    []hotkey.Modifier
		key2     hotkey.Key
		expected bool
	}{
		{
			name:     "Same hotkey",
			mods1:    []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModOption},
			key1:     hotkey.KeySpace,
			mods2:    []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModOption},
			key2:     hotkey.KeySpace,
			expected: true,
		},
		{
			name:     "Different key",
			mods1:    []hotkey.Modifier{hotkey.ModCtrl},
			key1:     hotkey.KeySpace,
			mods2:    []hotkey.Modifier{hotkey.ModCtrl},
			key2:     hotkey.KeyReturn,
			expected: false,
		},
		{
			name:     "Different modifiers",
			mods1:    []hotkey.Modifier{hotkey.ModCtrl},
			key1:     hotkey.KeySpace,
			mods2:    []hotkey.Modifier{hotkey.ModCmd},
			key2:     hotkey.KeySpace,
			expected: false,
		},
		{
			name:     "Same modifiers, different order",
			mods1:    []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModOption},
			key1:     hotkey.KeySpace,
			mods2:    []hotkey.Modifier{hotkey.ModOption, hotkey.ModCtrl},
			key2:     hotkey.KeySpace,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hotkeyMatches(tt.mods1, tt.key1, tt.mods2, tt.key2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestManagerLifecycle(t *testing.T) {
	m := New()

	// Initially should not be running
	if m.IsRunning() {
		t.Error("Manager should not be running initially")
	}

	// Close should be safe on non-running manager
	if err := m.Close(); err != nil {
		t.Errorf("Close() on non-running manager returned error: %v", err)
	}

	// Note: We cannot test actual registration here because it requires
	// proper permissions and may conflict with the test environment.
	// Integration tests should be run separately.
}

func TestEventChannel(t *testing.T) {
	m := New()

	eventChan := m.Events()
	if eventChan == nil {
		t.Fatal("Events() returned nil channel")
	}

	// Channel should be non-blocking initially
	select {
	case <-eventChan:
		t.Error("Events channel should be empty initially")
	case <-time.After(10 * time.Millisecond):
		// Expected: timeout
	}
}

func TestGetConfig(t *testing.T) {
	m := New()

	config := m.GetConfig()

	// Check default configuration
	if len(config.Modifiers) != 2 {
		t.Errorf("Expected 2 default modifiers, got %d", len(config.Modifiers))
	}

	if config.Key != hotkey.KeySpace {
		t.Errorf("Expected default key to be Space, got %v", config.Key)
	}

	if config.Mode != PressToHold {
		t.Errorf("Expected default mode to be PressToHold, got %v", config.Mode)
	}
}
