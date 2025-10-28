package clipboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

int get_pasteboard_change_count() {
    return (int)[[NSPasteboard generalPasteboard] changeCount];
}
*/
import "C"
import (
	"fmt"
	"strings"
	"time"

	"github.com/go-vgo/robotgo"
)

// Manager manages clipboard operations with safe restoration
type Manager struct {
	savedChangeCount int
	savedContent     string
	restoreTimeout   time.Duration
	splitSize        int
	splitInterval    time.Duration
}

// Config holds clipboard manager configuration
type Config struct {
	RestoreTimeout time.Duration // Timeout for clipboard restoration (default: 500ms)
	SplitSize      int           // Maximum characters per paste operation (default: 500)
	SplitInterval  time.Duration // Interval between split pastes (default: 50ms)
}

// DefaultConfig returns the default clipboard configuration
func DefaultConfig() Config {
	return Config{
		RestoreTimeout: 500 * time.Millisecond,
		SplitSize:      500,
		SplitInterval:  50 * time.Millisecond,
	}
}

// NewManager creates a new clipboard manager
func NewManager(config Config) *Manager {
	return &Manager{
		restoreTimeout: config.RestoreTimeout,
		splitSize:      config.SplitSize,
		splitInterval:  config.SplitInterval,
	}
}

// GetChangeCount returns the current pasteboard change count
func GetChangeCount() int {
	return int(C.get_pasteboard_change_count())
}

// SaveClipboard saves the current clipboard state
func (m *Manager) SaveClipboard() error {
	m.savedChangeCount = GetChangeCount()
	content, err := robotgo.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read clipboard: %w", err)
	}
	m.savedContent = content
	return nil
}

// RestoreClipboard restores the clipboard if it hasn't been modified externally
func (m *Manager) RestoreClipboard() error {
	// Wait a bit for the paste operation to complete
	time.Sleep(m.restoreTimeout)

	// Check if the change count matches (only one change = our paste operation)
	currentChangeCount := GetChangeCount()

	// If the change count increased by exactly 1, we're the only one who modified it
	// In this case, restore the original content
	if currentChangeCount == m.savedChangeCount+1 {
		robotgo.WriteAll(m.savedContent)
		return nil
	}

	// If the change count is different, the user modified the clipboard during our operation
	// Don't restore in this case
	return nil
}

// SafePaste pastes text to the active application with safe clipboard restoration
func (m *Manager) SafePaste(text string) error {
	// Save current clipboard state
	if err := m.SaveClipboard(); err != nil {
		return fmt.Errorf("failed to save clipboard: %w", err)
	}

	// Copy the text to clipboard
	robotgo.WriteAll(text)

	// Wait a bit for clipboard to update
	time.Sleep(10 * time.Millisecond)

	// Send Cmd+V to paste
	robotgo.KeyTap("v", "cmd")

	// Restore clipboard after a timeout
	return m.RestoreClipboard()
}

// SafePasteWithSplit pastes text with automatic splitting for long texts
func (m *Manager) SafePasteWithSplit(text string) error {
	// If text is short enough, paste directly
	if len(text) <= m.splitSize {
		return m.SafePaste(text)
	}

	// Split text into chunks
	chunks := m.splitText(text)

	// Paste each chunk
	for i, chunk := range chunks {
		if err := m.SafePaste(chunk); err != nil {
			return fmt.Errorf("failed to paste chunk %d: %w", i, err)
		}

		// Wait between chunks (except for the last one)
		if i < len(chunks)-1 {
			time.Sleep(m.splitInterval)
		}
	}

	return nil
}

// splitText splits text into chunks of maximum splitSize characters
// Tries to split at sentence boundaries (。、. ,) when possible
func (m *Manager) splitText(text string) []string {
	if len(text) <= m.splitSize {
		return []string{text}
	}

	var chunks []string
	runes := []rune(text)
	start := 0

	for start < len(runes) {
		// Calculate end position
		end := start + m.splitSize
		if end > len(runes) {
			end = len(runes)
		}

		// Try to find a good split point (sentence boundary)
		if end < len(runes) {
			// Look for sentence boundaries in the last 50 characters
			searchStart := end - 50
			if searchStart < start {
				searchStart = start
			}

			bestSplit := -1
			for i := end - 1; i >= searchStart; i-- {
				ch := runes[i]
				// Check for sentence endings
				if ch == '。' || ch == '、' || ch == '.' || ch == ',' || ch == '\n' {
					bestSplit = i + 1
					break
				}
			}

			// Use the best split point if found
			if bestSplit != -1 {
				end = bestSplit
			}
		}

		// Add chunk
		chunks = append(chunks, string(runes[start:end]))
		start = end
	}

	return chunks
}

// PasteDirectly pastes text without clipboard restoration (for testing)
func PasteDirectly(text string) error {
	robotgo.WriteAll(text)
	time.Sleep(10 * time.Millisecond)
	robotgo.KeyTap("v", "cmd")
	return nil
}

// GetClipboardContent returns the current clipboard content
func GetClipboardContent() (string, error) {
	return robotgo.ReadAll()
}

// SetClipboardContent sets the clipboard content
func SetClipboardContent(text string) error {
	robotgo.WriteAll(text)
	return nil
}

// SplitTextBySentences is a helper function to split text by sentences
// This is useful for preprocessing before pasting
func SplitTextBySentences(text string) []string {
	// Split by common sentence delimiters
	delimiters := []string{"。", ".", "！", "!", "？", "?"}

	sentences := []string{text}

	for _, delimiter := range delimiters {
		var newSentences []string
		for _, sentence := range sentences {
			parts := strings.Split(sentence, delimiter)
			for i, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					if i < len(parts)-1 {
						part += delimiter
					}
					newSentences = append(newSentences, part)
				}
			}
		}
		sentences = newSentences
	}

	return sentences
}
