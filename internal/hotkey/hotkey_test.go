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

func TestGetConfig_DeepCopy(t *testing.T) {
	m := New()

	// Get initial config
	config1 := m.GetConfig()
	originalLen := len(config1.Modifiers)

	// Try to mutate the returned config (these mutations should not affect internal state)
	if len(config1.Modifiers) > 0 {
		config1.Modifiers[0] = hotkey.ModCmd // Try to change first modifier
	}
	config1.Key = hotkey.KeyA  // Try to change key
	config1.Mode = Toggle      // Try to change mode
	_ = config1.Key            // Use the mutated values to avoid unused write warnings
	_ = config1.Mode

	// Get config again from manager
	config2 := m.GetConfig()

	// Verify internal state wasn't changed
	if len(config2.Modifiers) != originalLen {
		t.Errorf("Internal Modifiers length changed: expected %d, got %d",
			originalLen, len(config2.Modifiers))
	}

	if config2.Modifiers[0] != hotkey.ModCtrl {
		t.Errorf("Internal Modifiers[0] was mutated: expected ModCtrl, got %v",
			config2.Modifiers[0])
	}

	if config2.Key != hotkey.KeySpace {
		t.Errorf("Internal Key was mutated: expected KeySpace, got %v", config2.Key)
	}

	if config2.Mode != PressToHold {
		t.Errorf("Internal Mode was mutated: expected PressToHold, got %v", config2.Mode)
	}
}

func TestGetConfig_SliceMutation(t *testing.T) {
	m := New()

	// Get config and verify we get a copy of the slice
	config := m.GetConfig()

	// Store original modifier values
	originalFirstModifier := config.Modifiers[0]

	// Try to mutate the slice by appending
	config.Modifiers = append(config.Modifiers, hotkey.ModShift)

	// Get config again - should still have original length and values
	config2 := m.GetConfig()

	if len(config2.Modifiers) != 2 {
		t.Errorf("Appending to returned slice affected internal state: "+
			"expected 2 modifiers, got %d", len(config2.Modifiers))
	}

	if config2.Modifiers[0] != originalFirstModifier {
		t.Errorf("First modifier was changed: expected %v, got %v",
			originalFirstModifier, config2.Modifiers[0])
	}
}
