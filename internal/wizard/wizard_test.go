package wizard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSetupWizard(t *testing.T) {
	wizard, err := NewSetupWizard()

	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	if wizard == nil {
		t.Error("Expected wizard to be created")
	}

	if wizard.configDir == "" {
		t.Error("Expected configDir to be set")
	}

	if wizard.configPath == "" {
		t.Error("Expected configPath to be set")
	}

	if wizard.setupFlagFile == "" {
		t.Error("Expected setupFlagFile to be set")
	}
}

func TestIsFirstRun(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	// If config doesn't exist, it should be first run
	// Remove config if it exists
	os.Remove(wizard.configPath)

	if !wizard.IsFirstRun() {
		t.Error("Expected IsFirstRun to return true when config doesn't exist")
	}

	// Create a dummy config file
	file, err := os.Create(wizard.configPath)
	if err != nil {
		t.Fatalf("Failed to create dummy config: %v", err)
	}
	file.Close()

	// Now it should not be first run
	if wizard.IsFirstRun() {
		t.Error("Expected IsFirstRun to return false when config exists")
	}

	// Cleanup
	os.Remove(wizard.configPath)
}

func TestIsSetupCompleted(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	// Remove setup flag if it exists
	os.Remove(wizard.setupFlagFile)

	if wizard.IsSetupCompleted() {
		t.Error("Expected IsSetupCompleted to return false when flag doesn't exist")
	}

	// Create the setup flag
	err = wizard.MarkSetupCompleted()
	if err != nil {
		t.Fatalf("Failed to mark setup completed: %v", err)
	}

	if !wizard.IsSetupCompleted() {
		t.Error("Expected IsSetupCompleted to return true after marking completed")
	}

	// Cleanup
	os.Remove(wizard.setupFlagFile)
}

func TestMarkSetupCompleted(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	// Remove setup flag if it exists
	os.Remove(wizard.setupFlagFile)

	err = wizard.MarkSetupCompleted()
	if err != nil {
		t.Fatalf("Failed to mark setup completed: %v", err)
	}

	// Check if file was created
	_, err = os.Stat(wizard.setupFlagFile)
	if err != nil {
		t.Errorf("Setup flag file was not created: %v", err)
	}

	// Cleanup
	os.Remove(wizard.setupFlagFile)
}

func TestShouldShowWizard(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	// Clean up any existing files
	os.Remove(wizard.configPath)
	os.Remove(wizard.setupFlagFile)

	// Should show wizard if config doesn't exist
	if !wizard.ShouldShowWizard() {
		t.Error("Expected ShouldShowWizard to return true when config doesn't exist")
	}

	// Create config file
	file, err := os.Create(wizard.configPath)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	file.Close()

	// Should still show wizard if setup not completed
	if !wizard.ShouldShowWizard() {
		t.Error("Expected ShouldShowWizard to return true when setup not completed")
	}

	// Mark setup as completed
	err = wizard.MarkSetupCompleted()
	if err != nil {
		t.Fatalf("Failed to mark setup completed: %v", err)
	}

	// Should not show wizard if setup is completed
	if wizard.ShouldShowWizard() {
		t.Error("Expected ShouldShowWizard to return false when setup is completed")
	}

	// Cleanup
	os.Remove(wizard.configPath)
	os.Remove(wizard.setupFlagFile)
}

func TestGetProgress(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	progress := wizard.GetProgress()

	if progress.PermissionsSetup {
		t.Error("Expected PermissionsSetup to be false")
	}

	if progress.ModelSelected {
		t.Error("Expected ModelSelected to be false")
	}

	if progress.HotkeyConfigured {
		t.Error("Expected HotkeyConfigured to be false")
	}

	if progress.TestCompleted {
		t.Error("Expected TestCompleted to be false")
	}
}

func TestResetSetup(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	// Mark setup as completed
	err = wizard.MarkSetupCompleted()
	if err != nil {
		t.Fatalf("Failed to mark setup completed: %v", err)
	}

	// Verify it was marked
	if !wizard.IsSetupCompleted() {
		t.Error("Setup flag should have been created")
	}

	// Reset setup
	err = wizard.ResetSetup()
	if err != nil {
		t.Fatalf("Failed to reset setup: %v", err)
	}

	// Verify reset worked
	if wizard.IsSetupCompleted() {
		t.Error("Expected IsSetupCompleted to return false after reset")
	}
}

func TestGetConfigDir(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	configDir := wizard.GetConfigDir()

	if configDir == "" {
		t.Error("Expected configDir to be non-empty")
	}

	// Check if directory exists
	info, err := os.Stat(configDir)
	if err != nil {
		t.Errorf("Config directory should exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path should be a directory")
	}
}

func TestGetConfigPath(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	configPath := wizard.GetConfigPath()

	if configPath == "" {
		t.Error("Expected configPath to be non-empty")
	}

	// Verify it has the correct filename
	if filepath.Base(configPath) != "config.json" {
		t.Errorf("Expected config.json, got %s", filepath.Base(configPath))
	}
}

func TestConcurrentWizardOperations(t *testing.T) {
	wizard, err := NewSetupWizard()
	if err != nil {
		t.Fatalf("Failed to create wizard: %v", err)
	}

	// Clean up
	os.Remove(wizard.setupFlagFile)

	done := make(chan bool, 10)

	// Run concurrent operations
	for i := 0; i < 10; i++ {
		go func() {
			wizard.IsSetupCompleted()
			wizard.ShouldShowWizard()
			wizard.GetProgress()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without race conditions
	t.Log("Concurrent operations completed successfully")

	// Cleanup
	os.Remove(wizard.setupFlagFile)
}
