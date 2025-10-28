package recording

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxDuration != 60*time.Second {
		t.Errorf("Expected MaxDuration 60s, got %v", config.MaxDuration)
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{Idle, "Idle"},
		{Recording, "Recording"},
		{Processing, "Processing"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Note: Full integration tests with real hotkey.Manager and audio.AudioDriver
// should be run separately, as they require actual hardware and permissions.
// For now, we test the basic functionality without mocks.
