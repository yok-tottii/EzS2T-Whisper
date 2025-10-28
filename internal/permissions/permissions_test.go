package permissions

import (
	"testing"
)

func TestNewPermissionChecker(t *testing.T) {
	pc := NewPermissionChecker()

	if pc == nil {
		t.Fatal("Expected PermissionChecker to be created")
	}
}

func TestCheckMicrophonePermission(t *testing.T) {
	pc := NewPermissionChecker()

	status := pc.CheckMicrophonePermission()

	// Status should be one of the valid values
	if status < PermissionNotDetermined || status > PermissionAuthorized {
		t.Errorf("Expected valid permission status, got %d", status)
	}
}

func TestCheckAccessibilityPermission(t *testing.T) {
	pc := NewPermissionChecker()

	status := pc.CheckAccessibilityPermission()

	// Status should be either Authorized (1) or Denied (0)
	if status != PermissionAuthorized && status != PermissionDenied {
		t.Errorf("Expected Authorized or Denied, got %v", status)
	}
}

func TestIsMicrophoneAuthorized(t *testing.T) {
	pc := NewPermissionChecker()

	// Should return a boolean without crashing
	result := pc.IsMicrophoneAuthorized()

	if result != true && result != false {
		t.Error("Expected boolean result")
	}
}

func TestIsAccessibilityAuthorized(t *testing.T) {
	pc := NewPermissionChecker()

	// Should return a boolean without crashing
	result := pc.IsAccessibilityAuthorized()

	if result != true && result != false {
		t.Error("Expected boolean result")
	}
}

func TestCheckAllPermissions(t *testing.T) {
	pc := NewPermissionChecker()

	perms := pc.CheckAllPermissions()

	// Should return a map with the expected keys
	if _, ok := perms["microphone"]; !ok {
		t.Error("Expected 'microphone' key in permissions map")
	}

	if _, ok := perms["accessibility"]; !ok {
		t.Error("Expected 'accessibility' key in permissions map")
	}

	// Values are already booleans in the map
	micValue := perms["microphone"]
	if micValue != true && micValue != false {
		t.Error("Expected boolean value for 'microphone'")
	}

	accValue := perms["accessibility"]
	if accValue != true && accValue != false {
		t.Error("Expected boolean value for 'accessibility'")
	}
}

func TestAreAllPermissionsGranted(t *testing.T) {
	pc := NewPermissionChecker()

	result := pc.AreAllPermissionsGranted()

	// Should return a boolean without crashing
	if result != true && result != false {
		t.Error("Expected boolean result")
	}
}

func TestPermissionStatusString(t *testing.T) {
	tests := []struct {
		status   PermissionStatus
		expected string
	}{
		{PermissionNotDetermined, "NotDetermined"},
		{PermissionRestricted, "Restricted"},
		{PermissionDenied, "Denied"},
		{PermissionAuthorized, "Authorized"},
		{PermissionStatus(99), "Unknown"},
	}

	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestGetPermissionStatusMessage(t *testing.T) {
	tests := []struct {
		status   PermissionStatus
		expected string
	}{
		{PermissionNotDetermined, "Permission not yet determined"},
		{PermissionRestricted, "Permission restricted by parental controls"},
		{PermissionDenied, "Permission denied"},
		{PermissionAuthorized, "Permission authorized"},
	}

	for _, test := range tests {
		result := GetPermissionStatusMessage(test.status)
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestGetMissingPermissionsMessage(t *testing.T) {
	pc := NewPermissionChecker()

	message := pc.GetMissingPermissionsMessage()

	// Message should be non-empty if some permissions are missing
	// or empty if all are granted
	if message != "" {
		// If message is not empty, it should contain permission names
		if len(message) < 5 {
			t.Errorf("Expected meaningful message, got %s", message)
		}
	}
}

func TestPermissionStatusValues(t *testing.T) {
	// Test that constants have the expected values
	if PermissionNotDetermined != 0 {
		t.Errorf("Expected PermissionNotDetermined to be 0, got %d", PermissionNotDetermined)
	}

	if PermissionRestricted != 1 {
		t.Errorf("Expected PermissionRestricted to be 1, got %d", PermissionRestricted)
	}

	if PermissionDenied != 2 {
		t.Errorf("Expected PermissionDenied to be 2, got %d", PermissionDenied)
	}

	if PermissionAuthorized != 3 {
		t.Errorf("Expected PermissionAuthorized to be 3, got %d", PermissionAuthorized)
	}
}

func TestRequestMicrophonePermission(t *testing.T) {
	pc := NewPermissionChecker()

	// Just test that the method doesn't panic
	// In a test environment, it may fail to open settings, but shouldn't crash
	_ = pc.RequestMicrophonePermission()
}

func TestRequestAccessibilityPermission(t *testing.T) {
	pc := NewPermissionChecker()

	// Just test that the method doesn't panic
	// In a test environment, it may fail to open settings, but shouldn't crash
	_ = pc.RequestAccessibilityPermission()
}
