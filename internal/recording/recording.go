package recording

import (
	"fmt"
	"sync"
	"time"

	"github.com/yok-tottii/EzS2T-Whisper/internal/audio"
	"github.com/yok-tottii/EzS2T-Whisper/internal/hotkey"
)

// State represents the current recording state
type State int

const (
	// Idle means not recording
	Idle State = iota
	// Recording means currently recording audio
	Recording
	// Processing means processing recorded audio
	Processing
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case Idle:
		return "Idle"
	case Recording:
		return "Recording"
	case Processing:
		return "Processing"
	default:
		return "Unknown"
	}
}

// Manager manages the recording lifecycle and coordinates between hotkey and audio
type Manager struct {
	state       State
	hotkey      *hotkey.Manager
	audio       audio.AudioDriver
	maxDuration time.Duration
	dataChan    chan []byte
	stopTimer   *time.Timer
	mu          sync.Mutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// Config holds configuration for the recording manager
type Config struct {
	MaxDuration time.Duration
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		MaxDuration: 60 * time.Second,
	}
}

// New creates a new recording manager
func New(hk *hotkey.Manager, ad audio.AudioDriver, config Config) *Manager {
	return &Manager{
		state:       Idle,
		hotkey:      hk,
		audio:       ad,
		maxDuration: config.MaxDuration,
		dataChan:    make(chan []byte, 1),
		stopChan:    make(chan struct{}),
	}
}

// Start begins monitoring hotkey events and managing recording
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.handleHotkeyEvents()
}

// handleHotkeyEvents monitors hotkey events and triggers recording start/stop
func (m *Manager) handleHotkeyEvents() {
	defer m.wg.Done()

	for {
		select {
		case event, ok := <-m.hotkey.Events():
			if !ok {
				// Hotkey channel closed, exit
				return
			}

			switch event.Type {
			case hotkey.Pressed:
				if err := m.startRecording(); err != nil {
					// Log error (will be handled by logger in future)
					fmt.Printf("Failed to start recording: %v\n", err)
				}
			case hotkey.Released:
				if err := m.stopRecording(); err != nil {
					// Log error
					fmt.Printf("Failed to stop recording: %v\n", err)
				}
			}

		case <-m.stopChan:
			return
		}
	}
}

// startRecording starts recording audio
func (m *Manager) startRecording() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != Idle {
		return fmt.Errorf("already recording or processing (current state: %s)", m.state)
	}

	// Start audio recording
	if err := m.audio.StartRecording(); err != nil {
		return fmt.Errorf("failed to start audio recording: %w", err)
	}

	m.state = Recording

	// Set max duration timer
	m.stopTimer = time.AfterFunc(m.maxDuration, func() {
		if err := m.stopRecording(); err != nil {
			fmt.Printf("Auto-stop recording failed: %v\n", err)
		}
	})

	return nil
}

// stopRecording stops recording and sends the data to the data channel
func (m *Manager) stopRecording() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != Recording {
		return fmt.Errorf("not recording (current state: %s)", m.state)
	}

	// Cancel timer
	if m.stopTimer != nil {
		m.stopTimer.Stop()
		m.stopTimer = nil
	}

	// Change state to processing
	m.state = Processing

	// Stop audio recording (unlock mutex temporarily to avoid deadlock)
	m.mu.Unlock()
	data, err := m.audio.StopRecording()
	m.mu.Lock()

	if err != nil {
		m.state = Idle
		return fmt.Errorf("failed to stop audio recording: %w", err)
	}

	// Send data to channel (non-blocking)
	select {
	case m.dataChan <- data:
		// Data sent successfully
	default:
		// Channel full, skip this data
		fmt.Println("Warning: data channel full, skipping data")
	}

	// Reset to idle
	m.state = Idle

	return nil
}

// Data returns the channel for receiving recorded audio data
func (m *Manager) Data() <-chan []byte {
	return m.dataChan
}

// State returns the current recording state
func (m *Manager) GetState() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Stop stops the recording manager and releases resources
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If recording, stop it first
	if m.state == Recording {
		if m.stopTimer != nil {
			m.stopTimer.Stop()
			m.stopTimer = nil
		}

		// Stop audio recording
		if _, err := m.audio.StopRecording(); err != nil {
			return fmt.Errorf("failed to stop audio recording: %w", err)
		}

		m.state = Idle
	}

	// Signal handleHotkeyEvents to stop
	close(m.stopChan)

	// Wait for goroutines to finish
	m.wg.Wait()

	// Close data channel
	close(m.dataChan)

	return nil
}
