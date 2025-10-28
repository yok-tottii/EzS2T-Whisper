package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/yok-tottii/EzS2T-Whisper/internal/config"
)

// SetupWizard manages the initial application setup flow
type SetupWizard struct {
	configDir     string
	configPath    string
	setupFlagFile string
	mu            sync.RWMutex
}

// NewSetupWizard creates a new setup wizard
func NewSetupWizard() (*SetupWizard, error) {
	configPath := config.GetConfigPath()
	configDir := filepath.Dir(configPath)

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	setupFlagFile := filepath.Join(configDir, ".setup_completed")

	return &SetupWizard{
		configDir:     configDir,
		configPath:    configPath,
		setupFlagFile: setupFlagFile,
	}, nil
}

// IsFirstRun checks if this is the first run of the application
func (w *SetupWizard) IsFirstRun() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// First run if config doesn't exist
	_, err := os.Stat(w.configPath)
	return os.IsNotExist(err)
}

// IsSetupCompleted checks if the initial setup wizard has been completed
func (w *SetupWizard) IsSetupCompleted() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	_, err := os.Stat(w.setupFlagFile)
	return !os.IsNotExist(err)
}

// MarkSetupCompleted marks the setup wizard as completed
func (w *SetupWizard) MarkSetupCompleted() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Create the setup completed flag file
	file, err := os.Create(w.setupFlagFile)
	if err != nil {
		return fmt.Errorf("failed to create setup flag file: %w", err)
	}
	file.Close()

	return nil
}

// ShouldShowWizard returns true if the setup wizard should be shown
// This is true if:
// 1. The application is running for the first time, OR
// 2. The setup has not been completed yet
func (w *SetupWizard) ShouldShowWizard() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Check if config exists
	_, configErr := os.Stat(w.configPath)
	if os.IsNotExist(configErr) {
		return true
	}

	// Check if setup is completed
	_, setupErr := os.Stat(w.setupFlagFile)
	return os.IsNotExist(setupErr)
}

// GetSetupProgress returns the current setup progress
// Returns a structure with completion status of each wizard step
type SetupProgress struct {
	PermissionsSetup bool `json:"permissions_setup"`
	ModelSelected    bool `json:"model_selected"`
	HotkeyConfigured bool `json:"hotkey_configured"`
	TestCompleted    bool `json:"test_completed"`
}

// GetProgress returns the current setup progress
func (w *SetupWizard) GetProgress() SetupProgress {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// For now, return a default progress structure
	// In a real implementation, this would track individual step completion
	return SetupProgress{
		PermissionsSetup: false,
		ModelSelected:    false,
		HotkeyConfigured: false,
		TestCompleted:    false,
	}
}

// ResetSetup resets the setup state (for testing or manual reset)
func (w *SetupWizard) ResetSetup() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Remove setup flag file
	if err := os.Remove(w.setupFlagFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove setup flag file: %w", err)
	}

	return nil
}

// GetConfigDir returns the configuration directory
func (w *SetupWizard) GetConfigDir() string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.configDir
}

// GetConfigPath returns the configuration file path
func (w *SetupWizard) GetConfigPath() string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.configPath
}
