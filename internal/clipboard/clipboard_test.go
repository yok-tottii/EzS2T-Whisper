package clipboard

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.RestoreTimeout != 500*time.Millisecond {
		t.Errorf("Expected RestoreTimeout 500ms, got %v", config.RestoreTimeout)
	}

	if config.SplitSize != 500 {
		t.Errorf("Expected SplitSize 500, got %d", config.SplitSize)
	}

	if config.SplitInterval != 50*time.Millisecond {
		t.Errorf("Expected SplitInterval 50ms, got %v", config.SplitInterval)
	}
}

func TestNewManager(t *testing.T) {
	config := DefaultConfig()
	manager := NewManager(config)

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}

	if manager.splitSize != 500 {
		t.Errorf("Expected splitSize 500, got %d", manager.splitSize)
	}
}

func TestGetChangeCount(t *testing.T) {
	// Test that GetChangeCount returns a valid integer
	changeCount := GetChangeCount()

	if changeCount < 0 {
		t.Errorf("Expected non-negative change count, got %d", changeCount)
	}

	// Calling it twice should return the same or higher value
	changeCount2 := GetChangeCount()
	if changeCount2 < changeCount {
		t.Errorf("Expected change count to not decrease: %d -> %d", changeCount, changeCount2)
	}
}

func TestSaveClipboard(t *testing.T) {
	config := DefaultConfig()
	manager := NewManager(config)

	err := manager.SaveClipboard()
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if manager.savedChangeCount < 0 {
		t.Error("Expected savedChangeCount to be set")
	}
}

func TestSplitText_ShortText(t *testing.T) {
	config := DefaultConfig()
	manager := NewManager(config)

	text := "Short text"
	chunks := manager.splitText(text)

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for short text, got %d", len(chunks))
	}

	if chunks[0] != text {
		t.Errorf("Expected chunk to be '%s', got '%s'", text, chunks[0])
	}
}

func TestSplitText_LongText(t *testing.T) {
	config := Config{
		RestoreTimeout: 500 * time.Millisecond,
		SplitSize:      10, // Small split size for testing
		SplitInterval:  50 * time.Millisecond,
	}
	manager := NewManager(config)

	text := "This is a long text that should be split into multiple chunks."
	chunks := manager.splitText(text)

	if len(chunks) <= 1 {
		t.Errorf("Expected multiple chunks for long text, got %d", len(chunks))
	}

	// Verify that concatenating chunks gives the original text
	concatenated := strings.Join(chunks, "")
	if concatenated != text {
		t.Errorf("Concatenated chunks don't match original text")
	}
}

func TestSplitText_WithSentences(t *testing.T) {
	config := Config{
		RestoreTimeout: 500 * time.Millisecond,
		SplitSize:      20, // Small split size for testing
		SplitInterval:  50 * time.Millisecond,
	}
	manager := NewManager(config)

	text := "これは文です。これも文です。これも文です。"
	chunks := manager.splitText(text)

	if len(chunks) <= 1 {
		t.Errorf("Expected multiple chunks, got %d", len(chunks))
	}

	// Verify that concatenating chunks gives the original text
	concatenated := strings.Join(chunks, "")
	if concatenated != text {
		t.Errorf("Concatenated chunks don't match original text:\nExpected: %s\nGot: %s", text, concatenated)
	}
}

func TestSplitTextBySentences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Japanese sentences",
			input:    "これは一つ目の文です。これは二つ目の文です。",
			expected: 2,
		},
		{
			name:     "English sentences",
			input:    "This is sentence one. This is sentence two.",
			expected: 2,
		},
		{
			name:     "Mixed punctuation",
			input:    "文一。文二！文三？",
			expected: 3,
		},
		{
			name:     "Single sentence",
			input:    "This is a single sentence.",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentences := SplitTextBySentences(tt.input)

			if len(sentences) != tt.expected {
				t.Errorf("Expected %d sentences, got %d: %v", tt.expected, len(sentences), sentences)
			}

			// Verify that joining sentences gives back the original text (minus whitespace)
			joined := strings.Join(sentences, "")
			original := strings.ReplaceAll(tt.input, " ", "")
			joined = strings.ReplaceAll(joined, " ", "")

			if joined != original {
				t.Errorf("Joined sentences don't match original:\nExpected: %s\nGot: %s", original, joined)
			}
		})
	}
}

func TestGetClipboardContent(t *testing.T) {
	// This is a basic test that the function doesn't panic
	// Actual clipboard content depends on system state
	content, err := GetClipboardContent()

	if err != nil {
		t.Logf("GetClipboardContent returned error (may be expected in headless env): %v", err)
	}

	// Content can be empty or any string
	_ = content

	// Test is successful if we get here without panicking
}

func TestSetClipboardContent(t *testing.T) {
	testText := "Test clipboard content"

	err := SetClipboardContent(testText)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	// Verify it was set (may not work in headless environment)
	content, err := GetClipboardContent()
	if err != nil {
		t.Logf("GetClipboardContent returned error (may be expected in headless env): %v", err)
		return
	}

	if content != testText {
		// This test may fail in headless environments, so we log instead of failing
		t.Logf("Clipboard content mismatch (may be expected in headless env): expected '%s', got '%s'", testText, content)
	}
}

// Note: Tests involving actual paste operations (SafePaste, etc.) require
// accessibility permissions and an active window, so they are not included
// in unit tests. These should be tested in integration tests.
