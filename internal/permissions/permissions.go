package permissions

/*
#cgo CFLAGS: -x objective-c -fmodules
#cgo LDFLAGS: -framework AVFoundation -framework ApplicationServices

#import <AVFoundation/AVFoundation.h>
#import <ApplicationServices/ApplicationServices.h>

int check_microphone_permission() {
    AVAuthorizationStatus status = [AVCaptureDevice authorizationStatusForMediaType:AVMediaTypeAudio];
    return (int)status;
}

int check_accessibility_permission() {
    Boolean isAccessibilityEnabled = AXIsProcessTrusted();
    return isAccessibilityEnabled ? 1 : 0;
}
*/
import "C"

import (
	"os/exec"
)

// PermissionStatus represents the status of a system permission
type PermissionStatus int

const (
	// PermissionNotDetermined means the user hasn't been asked yet
	PermissionNotDetermined PermissionStatus = 0
	// PermissionRestricted means the permission is restricted by parental controls
	PermissionRestricted PermissionStatus = 1
	// PermissionDenied means the user has explicitly denied the permission
	PermissionDenied PermissionStatus = 2
	// PermissionAuthorized means the user has authorized the permission
	PermissionAuthorized PermissionStatus = 3
)

// PermissionChecker provides methods for checking macOS system permissions
type PermissionChecker struct{}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{}
}

// CheckMicrophonePermission checks if the application has microphone access permission
func (pc *PermissionChecker) CheckMicrophonePermission() PermissionStatus {
	status := C.check_microphone_permission()
	return PermissionStatus(status)
}

// CheckAccessibilityPermission checks if the application has accessibility permission
func (pc *PermissionChecker) CheckAccessibilityPermission() PermissionStatus {
	status := C.check_accessibility_permission()
	if status == 1 {
		return PermissionAuthorized
	}
	return PermissionDenied
}

// IsMicrophoneAuthorized returns whether microphone permission is granted
func (pc *PermissionChecker) IsMicrophoneAuthorized() bool {
	return pc.CheckMicrophonePermission() == PermissionAuthorized
}

// IsAccessibilityAuthorized returns whether accessibility permission is granted
func (pc *PermissionChecker) IsAccessibilityAuthorized() bool {
	return pc.CheckAccessibilityPermission() == PermissionAuthorized
}

// RequestMicrophonePermission opens system settings for microphone permission
func (pc *PermissionChecker) RequestMicrophonePermission() error {
	url := "x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone"
	cmd := exec.Command("open", url)
	return cmd.Run()
}

// RequestAccessibilityPermission opens system settings for accessibility permission
func (pc *PermissionChecker) RequestAccessibilityPermission() error {
	url := "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
	cmd := exec.Command("open", url)
	return cmd.Run()
}

// PermissionStatus string representation
func (ps PermissionStatus) String() string {
	switch ps {
	case PermissionNotDetermined:
		return "NotDetermined"
	case PermissionRestricted:
		return "Restricted"
	case PermissionDenied:
		return "Denied"
	case PermissionAuthorized:
		return "Authorized"
	default:
		return "Unknown"
	}
}

// CheckAllPermissions checks both microphone and accessibility permissions
func (pc *PermissionChecker) CheckAllPermissions() map[string]bool {
	return map[string]bool{
		"microphone":    pc.IsMicrophoneAuthorized(),
		"accessibility": pc.IsAccessibilityAuthorized(),
	}
}

// AreAllPermissionsGranted returns whether all required permissions are granted
func (pc *PermissionChecker) AreAllPermissionsGranted() bool {
	perms := pc.CheckAllPermissions()
	for _, granted := range perms {
		if !granted {
			return false
		}
	}
	return true
}

// GetPermissionStatusMessage returns a human-readable message for a permission status
func GetPermissionStatusMessage(status PermissionStatus) string {
	switch status {
	case PermissionNotDetermined:
		return "Permission not yet determined"
	case PermissionRestricted:
		return "Permission restricted by parental controls"
	case PermissionDenied:
		return "Permission denied"
	case PermissionAuthorized:
		return "Permission authorized"
	default:
		return "Unknown permission status"
	}
}

// GetMissingPermissionsMessage returns a message listing missing permissions
func (pc *PermissionChecker) GetMissingPermissionsMessage() string {
	var missing []string

	if !pc.IsMicrophoneAuthorized() {
		missing = append(missing, "マイク (Microphone)")
	}
	if !pc.IsAccessibilityAuthorized() {
		missing = append(missing, "アクセシビリティ (Accessibility)")
	}

	if len(missing) == 0 {
		return ""
	}

	message := "以下の権限が必要です:\n"
	for _, perm := range missing {
		message += "  • " + perm + "\n"
	}
	return message
}
