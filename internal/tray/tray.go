package tray

import (
	"fmt"
	"log"
	"os/exec"
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
	onRescanModels   func()
	onRecordTest     func()
	onAbout          func()
	onQuit           func()
	menuSettings     *systray.MenuItem
	menuRescan       *systray.MenuItem
	menuRecordTest   *systray.MenuItem
	menuAbout        *systray.MenuItem
	menuQuit         *systray.MenuItem
}

// Config holds tray manager configuration
type Config struct {
	OnReady        func() // Called when systray is ready for initialization
	OnSettings     func()
	OnRescanModels func()
	OnRecordTest   func()
	OnAbout        func()
	OnQuit         func()
}

// NewManager creates a new tray manager
func NewManager(config Config) *Manager {
	return &Manager{
		state:           StateIdle,
		onReadyCallback: config.OnReady,
		onSettings:      config.OnSettings,
		onRescanModels:  config.OnRescanModels,
		onRecordTest:    config.OnRecordTest,
		onAbout:         config.OnAbout,
		onQuit:          config.OnQuit,
	}
}

// Run starts the system tray (blocking call)
func (m *Manager) Run() {
	systray.Run(m.onReady, m.onExit)
}

// onReady is called when systray is ready
func (m *Manager) onReady() {
	// Set initial icon and tooltip
	m.updateIcon()
	systray.SetTitle("🎤") // テキスト表示で可視化
	systray.SetTooltip("EzS2T-Whisper")

	// Add menu items
	m.menuSettings = systray.AddMenuItem("設定を開く...", "Open settings page")
	m.menuRescan = systray.AddMenuItem("モデルを再スキャン", "Rescan model directory")
	m.menuRecordTest = systray.AddMenuItem("録音テスト", "Test recording pipeline")

	systray.AddSeparator()

	m.menuAbout = systray.AddMenuItem("バージョン情報", "Show version information")
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
		case <-m.menuRescan.ClickedCh:
			if m.onRescanModels != nil {
				m.onRescanModels()
			}
		case <-m.menuRecordTest.ClickedCh:
			if m.onRecordTest != nil {
				m.onRecordTest()
			}
		case <-m.menuAbout.ClickedCh:
			if m.onAbout != nil {
				m.onAbout()
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
		systray.SetIcon(getIdleIcon())
		systray.SetTitle("🎤")
		systray.SetTooltip("EzS2T-Whisper - 待機中")
	case StateRecording:
		systray.SetIcon(getRecordingIcon())
		systray.SetTitle("🔴")
		systray.SetTooltip("EzS2T-Whisper - 録音中")
	case StateProcessing:
		systray.SetIcon(getProcessingIcon())
		systray.SetTitle("⏳")
		systray.SetTooltip("EzS2T-Whisper - 処理中")
	}
}

// Quit quits the system tray
func (m *Manager) Quit() {
	systray.Quit()
}

// getIdleIcon returns the idle state icon (white microphone)
func getIdleIcon() []byte {
	// Simple base64-encoded PNG icon (16x16 white microphone)
	// This is a placeholder - in production, use proper icon files
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

// getRecordingIcon returns the recording state icon (red microphone)
func getRecordingIcon() []byte {
	// Simple base64-encoded PNG icon (16x16 red microphone)
	// This is a placeholder - in production, use proper icon files
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

// getProcessingIcon returns the processing state icon (blue microphone with spinner)
func getProcessingIcon() []byte {
	// Simple base64-encoded PNG icon (16x16 blue microphone)
	// This is a placeholder - in production, use proper icon files
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
