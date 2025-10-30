package hotkey

import (
	"fmt"
	"sync"

	"golang.design/x/hotkey"
)

// RecordingMode defines how the hotkey triggers recording
type RecordingMode int

const (
	// PressToHold mode: record while key is held down
	PressToHold RecordingMode = iota
	// Toggle mode: first press starts, second press stops
	Toggle
)

// EventType represents the type of hotkey event
type EventType int

const (
	// Pressed indicates the hotkey was pressed
	Pressed EventType = iota
	// Released indicates the hotkey was released
	Released
)

// Event represents a hotkey event
type Event struct {
	Type EventType
}

// Config holds hotkey configuration
type Config struct {
	Modifiers []hotkey.Modifier
	Key       hotkey.Key
	Mode      RecordingMode
}

// Manager manages global hotkey registration and events
type Manager struct {
	hk        *hotkey.Hotkey
	config    Config
	eventChan chan Event
	stopChan  chan struct{}
	wg        sync.WaitGroup
	mu        sync.Mutex
	running   bool
}

// New creates a new hotkey manager with default configuration
// Default: Ctrl+Option+Space
func New() *Manager {
	return &Manager{
		config: Config{
			Modifiers: []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModOption},
			Key:       hotkey.KeySpace,
			Mode:      PressToHold,
		},
		eventChan: make(chan Event, 10),
		stopChan:  make(chan struct{}),
	}
}

// Register registers the hotkey with the system
func (m *Manager) Register(config Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("hotkey is already running, call Close() first")
	}

	m.config = config

	// Recreate channels (they may have been closed by a previous Close())
	m.stopChan = make(chan struct{})
	m.eventChan = make(chan Event, 10)

	// Create hotkey instance
	hk := hotkey.New(m.config.Modifiers, m.config.Key)

	// Register the hotkey
	if err := hk.Register(); err != nil {
		return fmt.Errorf("failed to register hotkey: %w", err)
	}

	m.hk = hk
	m.running = true

	// Start listening in a goroutine
	m.wg.Add(1)
	go m.listen()

	return nil
}

// RegisterDefault registers the default hotkey (Ctrl+Option+Space)
func (m *Manager) RegisterDefault() error {
	return m.Register(m.config)
}

// listen monitors hotkey events and sends them to the event channel
func (m *Manager) listen() {
	defer m.wg.Done()

	toggleState := false

	for {
		select {
		case <-m.hk.Keydown():
			switch m.config.Mode {
			case PressToHold:
				m.eventChan <- Event{Type: Pressed}
			case Toggle:
				if !toggleState {
					m.eventChan <- Event{Type: Pressed}
					toggleState = true
				} else {
					m.eventChan <- Event{Type: Released}
					toggleState = false
				}
			}

		case <-m.hk.Keyup():
			if m.config.Mode == PressToHold {
				m.eventChan <- Event{Type: Released}
			}

		case <-m.stopChan:
			return
		}
	}
}

// Events returns the event channel for receiving hotkey events
func (m *Manager) Events() <-chan Event {
	return m.eventChan
}

// Close unregisters the hotkey and stops listening
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	var unregisterErr error

	// Signal the listener to stop
	close(m.stopChan)

	// Wait for the listener goroutine to finish
	m.wg.Wait()

	// Unregister the hotkey
	// 注意: エラーが発生しても続行し、必ずクリーンアップを実行する
	if m.hk != nil {
		if err := m.hk.Unregister(); err != nil {
			unregisterErr = fmt.Errorf("failed to unregister hotkey: %w", err)
		}
	}

	// Close event channel to notify consumers of shutdown
	if m.eventChan != nil {
		close(m.eventChan)
		m.eventChan = nil
	}

	// 必ず running フラグを false にセット
	// これにより、Unregister() が失敗しても次の Register() が可能になる
	m.running = false

	return unregisterErr
}

// IsRunning returns whether the hotkey is currently registered and running
func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// GetConfig returns a deep copy of the current hotkey configuration
// to prevent callers from modifying the Manager's internal state
func (m *Manager) GetConfig() Config {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a shallow copy of the config struct
	configCopy := m.config

	// Deep copy the Modifiers slice to prevent caller from mutating it
	if m.config.Modifiers != nil {
		configCopy.Modifiers = make([]hotkey.Modifier, len(m.config.Modifiers))
		copy(configCopy.Modifiers, m.config.Modifiers)
	}

	return configCopy
}
