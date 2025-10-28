package notification

import (
	"testing"
)

func TestNewNotificationManager(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	if nm == nil {
		t.Fatal("Expected notification manager to be created")
	}

	if nm.appName != "TestApp" {
		t.Errorf("Expected appName to be TestApp, got %s", nm.appName)
	}
}

func TestSendInfo(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	// In test environment, this may fail to send actual notification,
	// but we just verify the method doesn't panic
	err := nm.SendInfo("Test Title", "Test Message")

	// Error is acceptable in test environment (no display available)
	if err != nil {
		t.Logf("SendInfo returned error (expected in test env): %v", err)
	}
}

func TestSendWarning(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.SendWarning("Test Title", "Test Warning")

	if err != nil {
		t.Logf("SendWarning returned error (expected in test env): %v", err)
	}
}

func TestSendError(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.SendError("Test Title", "Test Error")

	if err != nil {
		t.Logf("SendError returned error (expected in test env): %v", err)
	}
}

func TestSendSuccess(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.SendSuccess("Test Title", "Test Success")

	if err != nil {
		t.Logf("SendSuccess returned error (expected in test env): %v", err)
	}
}

func TestRecordingStarted(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.RecordingStarted()

	if err != nil {
		t.Logf("RecordingStarted returned error (expected in test env): %v", err)
	}
}

func TestRecordingStopped(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.RecordingStopped()

	if err != nil {
		t.Logf("RecordingStopped returned error (expected in test env): %v", err)
	}
}

func TestTranscriptionComplete(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.TranscriptionComplete()

	if err != nil {
		t.Logf("TranscriptionComplete returned error (expected in test env): %v", err)
	}
}

func TestPasteComplete(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.PasteComplete()

	if err != nil {
		t.Logf("PasteComplete returned error (expected in test env): %v", err)
	}
}

func TestMicrophonePermissionDenied(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.MicrophonePermissionDenied()

	if err != nil {
		t.Logf("MicrophonePermissionDenied returned error (expected in test env): %v", err)
	}
}

func TestAccessibilityPermissionDenied(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.AccessibilityPermissionDenied()

	if err != nil {
		t.Logf("AccessibilityPermissionDenied returned error (expected in test env): %v", err)
	}
}

func TestRecordingFailed(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.RecordingFailed("Device not found")

	if err != nil {
		t.Logf("RecordingFailed returned error (expected in test env): %v", err)
	}
}

func TestTranscriptionFailed(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.TranscriptionFailed("Model not found")

	if err != nil {
		t.Logf("TranscriptionFailed returned error (expected in test env): %v", err)
	}
}

func TestRecordingTimeExceeded(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.RecordingTimeExceeded()

	if err != nil {
		t.Logf("RecordingTimeExceeded returned error (expected in test env): %v", err)
	}
}

func TestDeviceNotFound(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.DeviceNotFound()

	if err != nil {
		t.Logf("DeviceNotFound returned error (expected in test env): %v", err)
	}
}

func TestModelNotFound(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.ModelNotFound("/path/to/model.gguf")

	if err != nil {
		t.Logf("ModelNotFound returned error (expected in test env): %v", err)
	}
}

func TestSendNilNotification(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	err := nm.Send(nil)

	if err == nil {
		t.Error("Expected error when sending nil notification")
	}
}

func TestNotificationType(t *testing.T) {
	types := []NotificationType{TypeInfo, TypeWarning, TypeError, TypeSuccess}

	for _, nt := range types {
		if nt == "" {
			t.Errorf("Notification type should not be empty")
		}
	}
}

func TestCustomNotification(t *testing.T) {
	nm := NewNotificationManager("TestApp")

	notification := &Notification{
		Title:   "Custom Title",
		Message: "Custom Message",
		Type:    TypeInfo,
		AppName: "TestApp",
	}

	err := nm.Send(notification)

	if err != nil {
		t.Logf("Send custom notification returned error (expected in test env): %v", err)
	}
}
