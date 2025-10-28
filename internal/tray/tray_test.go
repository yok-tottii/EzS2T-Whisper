package tray

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	settingsCalled := false
	rescanCalled := false
	recordTestCalled := false
	aboutCalled := false
	quitCalled := false

	config := Config{
		OnSettings: func() {
			settingsCalled = true
		},
		OnRescanModels: func() {
			rescanCalled = true
		},
		OnRecordTest: func() {
			recordTestCalled = true
		},
		OnAbout: func() {
			aboutCalled = true
		},
		OnQuit: func() {
			quitCalled = true
		},
	}

	manager := NewManager(config)

	if manager == nil {
		t.Fatal("Expected manager to be created")
	}

	if manager.state != StateIdle {
		t.Errorf("Expected initial state to be StateIdle, got %v", manager.state)
	}

	// Test callback invocation
	if manager.onSettings != nil {
		manager.onSettings()
		if !settingsCalled {
			t.Error("Expected onSettings callback to be called")
		}
	}

	if manager.onRescanModels != nil {
		manager.onRescanModels()
		if !rescanCalled {
			t.Error("Expected onRescanModels callback to be called")
		}
	}

	if manager.onRecordTest != nil {
		manager.onRecordTest()
		if !recordTestCalled {
			t.Error("Expected onRecordTest callback to be called")
		}
	}

	if manager.onAbout != nil {
		manager.onAbout()
		if !aboutCalled {
			t.Error("Expected onAbout callback to be called")
		}
	}

	if manager.onQuit != nil {
		manager.onQuit()
		if !quitCalled {
			t.Error("Expected onQuit callback to be called")
		}
	}
}

func TestSetState(t *testing.T) {
	manager := NewManager(Config{})

	// Test initial state
	if manager.state != StateIdle {
		t.Errorf("Expected initial state to be StateIdle, got %v", manager.state)
	}

	// Test state transitions
	manager.SetState(StateRecording)
	if manager.state != StateRecording {
		t.Errorf("Expected state to be StateRecording, got %v", manager.state)
	}

	manager.SetState(StateProcessing)
	if manager.state != StateProcessing {
		t.Errorf("Expected state to be StateProcessing, got %v", manager.state)
	}

	manager.SetState(StateIdle)
	if manager.state != StateIdle {
		t.Errorf("Expected state to be StateIdle, got %v", manager.state)
	}
}

func TestIconFunctions(t *testing.T) {
	// Test that icon functions return non-empty byte slices
	idleIcon := getIdleIcon()
	if len(idleIcon) == 0 {
		t.Error("Expected getIdleIcon to return non-empty byte slice")
	}

	recordingIcon := getRecordingIcon()
	if len(recordingIcon) == 0 {
		t.Error("Expected getRecordingIcon to return non-empty byte slice")
	}

	processingIcon := getProcessingIcon()
	if len(processingIcon) == 0 {
		t.Error("Expected getProcessingIcon to return non-empty byte slice")
	}

	// Verify they're different
	if string(idleIcon) == string(recordingIcon) {
		t.Error("Expected idle and recording icons to be different")
	}

	if string(idleIcon) == string(processingIcon) {
		t.Error("Expected idle and processing icons to be different")
	}

	if string(recordingIcon) == string(processingIcon) {
		t.Error("Expected recording and processing icons to be different")
	}
}

func TestShowNotification(t *testing.T) {
	manager := NewManager(Config{})

	// These are just basic tests to ensure they don't panic
	manager.ShowNotification("Test", "Test message")
	manager.ShowError("Test error")
	manager.ShowSuccess("Test success")
}

func TestCallbacksNil(t *testing.T) {
	// Test that manager works with nil callbacks
	manager := NewManager(Config{})

	if manager == nil {
		t.Fatal("Expected manager to be created with nil callbacks")
	}

	// These should not panic even with nil callbacks
	if manager.onSettings != nil {
		manager.onSettings()
	}
	if manager.onRescanModels != nil {
		manager.onRescanModels()
	}
	if manager.onRecordTest != nil {
		manager.onRecordTest()
	}
	if manager.onAbout != nil {
		manager.onAbout()
	}
	if manager.onQuit != nil {
		manager.onQuit()
	}
}

func TestStateConstants(t *testing.T) {
	// Verify state constants have expected values
	if StateIdle != 0 {
		t.Errorf("Expected StateIdle to be 0, got %d", StateIdle)
	}
	if StateRecording != 1 {
		t.Errorf("Expected StateRecording to be 1, got %d", StateRecording)
	}
	if StateProcessing != 2 {
		t.Errorf("Expected StateProcessing to be 2, got %d", StateProcessing)
	}
}

func TestUpdateIcon(t *testing.T) {
	manager := NewManager(Config{})

	// Test that updateIcon doesn't panic for each state
	manager.state = StateIdle
	manager.updateIcon()

	manager.state = StateRecording
	manager.updateIcon()

	manager.state = StateProcessing
	manager.updateIcon()
}

func TestConcurrentStateUpdates(t *testing.T) {
	manager := NewManager(Config{})

	// Test concurrent state updates don't cause races
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			manager.SetState(StateRecording)
			time.Sleep(1 * time.Millisecond)
			manager.SetState(StateProcessing)
			time.Sleep(1 * time.Millisecond)
			manager.SetState(StateIdle)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Final state should be one of the valid states
	if manager.state != StateIdle && manager.state != StateRecording && manager.state != StateProcessing {
		t.Errorf("Invalid final state: %v", manager.state)
	}
}
