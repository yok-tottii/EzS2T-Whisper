package audio

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.SampleRate != 16000 {
		t.Errorf("Expected sample rate 16000, got %d", config.SampleRate)
	}

	if config.Channels != 1 {
		t.Errorf("Expected 1 channel, got %d", config.Channels)
	}

	if config.Latency != HighStability {
		t.Errorf("Expected HighStability latency, got %v", config.Latency)
	}

	if config.DeviceID != -1 {
		t.Errorf("Expected default device ID -1, got %d", config.DeviceID)
	}
}

func TestNewPortAudioDriver(t *testing.T) {
	driver, err := NewPortAudioDriver()
	if err != nil {
		t.Skipf("PortAudio not available: %v", err)
	}
	defer driver.Close()

	if driver == nil {
		t.Fatal("Expected non-nil driver")
	}
}

func TestListDevices(t *testing.T) {
	driver, err := NewPortAudioDriver()
	if err != nil {
		t.Skipf("PortAudio not available: %v", err)
	}
	defer driver.Close()

	devices, err := driver.ListDevices()
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	// Should have at least one device
	if len(devices) == 0 {
		t.Skip("No audio input devices available")
	}

	t.Logf("Found %d input devices", len(devices))
	for _, dev := range devices {
		t.Logf("Device %d: %s (default: %v)", dev.ID, dev.Name, dev.IsDefault)
	}

	// At least one device should be marked as default
	hasDefault := false
	for _, dev := range devices {
		if dev.IsDefault {
			hasDefault = true
			break
		}
	}

	if !hasDefault {
		t.Error("No default device found")
	}
}

func TestInitialize(t *testing.T) {
	driver, err := NewPortAudioDriver()
	if err != nil {
		t.Skipf("PortAudio not available: %v", err)
	}
	defer driver.Close()

	config := DefaultConfig()
	if err := driver.Initialize(config); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !driver.initialized {
		t.Error("Driver should be initialized")
	}
}

func TestIsRecording(t *testing.T) {
	driver, err := NewPortAudioDriver()
	if err != nil {
		t.Skipf("PortAudio not available: %v", err)
	}
	defer driver.Close()

	// Initialize first
	config := DefaultConfig()
	if err := driver.Initialize(config); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Should not be recording initially
	if driver.IsRecording() {
		t.Error("Should not be recording initially")
	}

	// Start recording
	if err := driver.StartRecording(); err != nil {
		t.Fatalf("StartRecording failed: %v", err)
	}

	// Should be recording now
	if !driver.IsRecording() {
		t.Error("Should be recording after StartRecording")
	}

	// Stop recording
	if _, err := driver.StopRecording(); err != nil {
		t.Fatalf("StopRecording failed: %v", err)
	}

	// Should not be recording anymore
	if driver.IsRecording() {
		t.Error("Should not be recording after StopRecording")
	}
}

func TestRecordingLifecycle(t *testing.T) {
	driver, err := NewPortAudioDriver()
	if err != nil {
		t.Skipf("PortAudio not available: %v", err)
	}
	defer driver.Close()

	// Initialize
	config := DefaultConfig()
	if err := driver.Initialize(config); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Start recording should succeed
	if err := driver.StartRecording(); err != nil {
		t.Fatalf("StartRecording failed: %v", err)
	}

	// Starting again should fail
	if err := driver.StartRecording(); err == nil {
		t.Error("StartRecording should fail when already recording")
	}

	// Stop recording
	data, err := driver.StopRecording()
	if err != nil {
		t.Fatalf("StopRecording failed: %v", err)
	}

	// Data should be non-nil (might be empty if recording was very short)
	if data == nil {
		t.Error("StopRecording returned nil data")
	}

	t.Logf("Recorded %d bytes", len(data))

	// Stopping again should fail
	if _, err := driver.StopRecording(); err == nil {
		t.Error("StopRecording should fail when not recording")
	}
}

func TestClose(t *testing.T) {
	driver, err := NewPortAudioDriver()
	if err != nil {
		t.Skipf("PortAudio not available: %v", err)
	}

	// Initialize
	config := DefaultConfig()
	if err := driver.Initialize(config); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Close should succeed
	if err := driver.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Should not be initialized after close
	if driver.initialized {
		t.Error("Driver should not be initialized after Close")
	}
}
