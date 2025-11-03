package tray

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/getlantern/systray"
)

// State represents the current application state
type State int

const (
	StateIdle State = iota
	StateRecording
	StateProcessing
)

// Manager manages the system tray icon and menu
type Manager struct {
	stateMutex       sync.RWMutex
	state            State
	onReadyCallback  func()
	onSettings       func()
	onRecordTest     func()
	onDeviceChange   func(deviceID int) // Called when user selects a device
	onQuit           func()
	menuSettings      *systray.MenuItem
	menuDevices       *systray.MenuItem      // Parent menu for device selection
	menuRecordTest    *systray.MenuItem
	menuQuit          *systray.MenuItem
	deviceMenuItems   []*systray.MenuItem    // Device submenu items
	deviceCancelFuncs []context.CancelFunc   // Cancel functions for device menu goroutines

	// Icon cache
	iconIdle       []byte
	iconRecording  []byte
	iconProcessing []byte
}

// Config holds tray manager configuration
type Config struct {
	OnReady        func() // Called when systray is ready for initialization
	OnSettings     func()
	OnRecordTest   func()
	OnDeviceChange func(deviceID int) // Called when user selects a device
	OnQuit         func()
}

// NewManager creates a new tray manager
func NewManager(config Config) *Manager {
	m := &Manager{
		state:           StateIdle,
		onReadyCallback: config.OnReady,
		onSettings:      config.OnSettings,
		onRecordTest:    config.OnRecordTest,
		onDeviceChange:  config.OnDeviceChange,
		onQuit:          config.OnQuit,
	}

	// Load icons once at initialization
	m.iconIdle = loadIconData("speech_to_text_32dp_E3E3E3_FILL0_wght400_GRAD0_opsz40.png", getIdleFallback())
	m.iconRecording = loadIconData("graphic_eq_32dp_F19E39_FILL0_wght400_GRAD0_opsz40.png", getRecordingFallback())
	m.iconProcessing = loadIconData("hourglass_empty_32dp_75FB4C_FILL0_wght400_GRAD0_opsz40.png", getProcessingFallback())

	return m
}

// Run starts the system tray (blocking call)
func (m *Manager) Run() {
	systray.Run(m.onReady, m.onExit)
}

// onReady is called when systray is ready
func (m *Manager) onReady() {
	// Set initial icon and tooltip
	m.updateIcon()
	systray.SetTooltip("EzS2T-Whisper")

	// Add menu items
	m.menuSettings = systray.AddMenuItem("設定を開く...", "Open settings page")
	m.menuDevices = systray.AddMenuItem("入力デバイス", "Select input device")
	m.menuRecordTest = systray.AddMenuItem("録音テスト", "Test recording pipeline")

	systray.AddSeparator()

	m.menuQuit = systray.AddMenuItem("終了", "Quit the application")

	// Start event loop
	go m.handleMenuEvents()

	// Call the OnReady callback if provided
	if m.onReadyCallback != nil {
		m.onReadyCallback()
	}
}

// onExit is called when systray is exiting
func (m *Manager) onExit() {
	// Cleanup if needed
}

// handleMenuEvents handles menu item clicks
func (m *Manager) handleMenuEvents() {
	for {
		select {
		case <-m.menuSettings.ClickedCh:
			if m.onSettings != nil {
				m.onSettings()
			}
		case <-m.menuRecordTest.ClickedCh:
			if m.onRecordTest != nil {
				m.onRecordTest()
			}
		case <-m.menuQuit.ClickedCh:
			if m.onQuit != nil {
				m.onQuit()
			}
			systray.Quit()
			return
		}
	}
}

// SetState updates the tray icon based on the current state
func (m *Manager) SetState(state State) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()
	m.state = state
	m.updateIcon()
}

// updateIcon updates the tray icon based on the current state
func (m *Manager) updateIcon() {
	switch m.state {
	case StateIdle:
		systray.SetIcon(m.iconIdle)
		systray.SetTooltip("EzS2T-Whisper - 待機中")
	case StateRecording:
		systray.SetIcon(m.iconRecording)
		systray.SetTooltip("EzS2T-Whisper - 録音中")
	case StateProcessing:
		systray.SetIcon(m.iconProcessing)
		systray.SetTooltip("EzS2T-Whisper - 処理中")
	}
}

// Device represents an audio device for the menu
type Device struct {
	ID        int
	Name      string
	IsDefault bool
	IsCurrent bool
}

// UpdateDeviceMenu updates the device submenu with available devices
func (m *Manager) UpdateDeviceMenu(devices []Device) {
	// Cancel existing device menu goroutines
	for _, cancel := range m.deviceCancelFuncs {
		if cancel != nil {
			cancel()
		}
	}
	m.deviceCancelFuncs = nil

	// Remove existing device menu items
	for _, item := range m.deviceMenuItems {
		item.Hide()
	}
	m.deviceMenuItems = nil

	// Add new device menu items
	for _, device := range devices {
		// Create closure to capture device ID
		deviceID := device.ID
		deviceName := device.Name

		// Add checkmark if current device
		prefix := ""
		if device.IsCurrent {
			prefix = "✓ "
		}

		// Add tooltip for default device
		tooltip := ""
		if device.IsDefault {
			tooltip = "System default device"
		}

		menuItem := m.menuDevices.AddSubMenuItem(prefix+deviceName, tooltip)
		m.deviceMenuItems = append(m.deviceMenuItems, menuItem)

		// Create context for this goroutine
		ctx, cancel := context.WithCancel(context.Background())
		m.deviceCancelFuncs = append(m.deviceCancelFuncs, cancel)

		// Handle device selection in a goroutine with cancellation
		go func(id int, item *systray.MenuItem, ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					// Context cancelled, exit goroutine
					return
				case <-item.ClickedCh:
					if m.onDeviceChange != nil {
						m.onDeviceChange(id)
					}
				}
			}
		}(deviceID, menuItem, ctx)
	}
}

// Quit quits the system tray
func (m *Manager) Quit() {
	systray.Quit()
}

// loadIconData loads an icon from the assets directory
// If the file cannot be loaded, it returns a fallback placeholder icon
func loadIconData(filename string, fallback []byte) []byte {
	// Get executable directory
	exe, err := os.Executable()
	if err != nil {
		log.Printf("警告: 実行ファイルのパスを取得できませんでした: %v", err)
		return fallback
	}
	exeDir := filepath.Dir(exe)

	// Try to load icon from assets/icon/ relative to executable
	iconPath := filepath.Join(exeDir, "assets", "icon", filename)
	data, err := os.ReadFile(iconPath)
	if err != nil {
		log.Printf("警告: アイコンファイルを読み込めませんでした (%s): %v", iconPath, err)
		return fallback
	}

	return data
}

// getIdleFallback returns the fallback icon data for idle state
func getIdleFallback() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff,
		0x61, 0x00, 0x00, 0x00, 0x19, 0x74, 0x45, 0x58,
		0x74, 0x53, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72,
		0x65, 0x00, 0x41, 0x64, 0x6f, 0x62, 0x65, 0x20,
		0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x61,
		0x64, 0x79, 0x71, 0xc9, 0x65, 0x3c, 0x00, 0x00,
		0x00, 0x18, 0x49, 0x44, 0x41, 0x54, 0x78, 0xda,
		0x62, 0xfc, 0xff, 0xff, 0x3f, 0x03, 0x00, 0x00,
		0x00, 0xff, 0xff, 0x03, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60,
		0x82,
	}
}

// getRecordingFallback returns the fallback icon data for recording state
func getRecordingFallback() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff,
		0x61, 0x00, 0x00, 0x00, 0x19, 0x74, 0x45, 0x58,
		0x74, 0x53, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72,
		0x65, 0x00, 0x41, 0x64, 0x6f, 0x62, 0x65, 0x20,
		0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x61,
		0x64, 0x79, 0x71, 0xc9, 0x65, 0x3c, 0x00, 0x00,
		0x00, 0x20, 0x49, 0x44, 0x41, 0x54, 0x78, 0xda,
		0x62, 0xfc, 0xcf, 0xc0, 0xc0, 0xc0, 0xf0, 0x9f,
		0x81, 0x81, 0x81, 0x81, 0xff, 0x19, 0x18, 0x18,
		0x18, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0x03,
		0x00, 0x0c, 0x10, 0x02, 0x01, 0x8b, 0xd5, 0xf8,
		0x23, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}

// getProcessingFallback returns the fallback icon data for processing state
func getProcessingFallback() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff,
		0x61, 0x00, 0x00, 0x00, 0x19, 0x74, 0x45, 0x58,
		0x74, 0x53, 0x6f, 0x66, 0x74, 0x77, 0x61, 0x72,
		0x65, 0x00, 0x41, 0x64, 0x6f, 0x62, 0x65, 0x20,
		0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x61,
		0x64, 0x79, 0x71, 0xc9, 0x65, 0x3c, 0x00, 0x00,
		0x00, 0x20, 0x49, 0x44, 0x41, 0x54, 0x78, 0xda,
		0x62, 0xfc, 0xcf, 0xf0, 0x9f, 0xc1, 0xc8, 0xc0,
		0xc0, 0xc0, 0xff, 0x0c, 0x0c, 0x0c, 0xfc, 0xcf,
		0xc0, 0xc0, 0xc0, 0x00, 0x00, 0x00, 0x00, 0xff,
		0xff, 0x03, 0x00, 0x0c, 0x50, 0x02, 0x01, 0x3e,
		0x0a, 0xe4, 0x5b, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}

// ShowNotification shows a notification using macOS Notification Center
func (m *Manager) ShowNotification(title, message string) {
	log.Printf("Notification: %s - %s", title, message)

	// macOS通知センターを使用
	script := fmt.Sprintf(`display notification "%s" with title "%s"`,
		escapeAppleScript(message),
		escapeAppleScript(title))
	exec.Command("osascript", "-e", script).Run()
}

// escapeAppleScript escapes special characters for AppleScript
func escapeAppleScript(s string) string {
	// Escape backslashes first to avoid double-escaping
	s = strings.ReplaceAll(s, `\`, `\\`)
	// Escape double quotes
	s = strings.ReplaceAll(s, `"`, `\"`)
	// Escape control characters
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// ShowError shows an error notification
func (m *Manager) ShowError(message string) {
	m.ShowNotification("EzS2T-Whisper Error", message)
}

// ShowSuccess shows a success notification
func (m *Manager) ShowSuccess(message string) {
	m.ShowNotification("EzS2T-Whisper", message)
}
