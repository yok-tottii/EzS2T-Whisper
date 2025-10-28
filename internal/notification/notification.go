package notification

import (
	"fmt"
	"os/exec"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// TypeInfo is an informational notification
	TypeInfo NotificationType = "info"
	// TypeWarning is a warning notification
	TypeWarning NotificationType = "warning"
	// TypeError is an error notification
	TypeError NotificationType = "error"
	// TypeSuccess is a success notification
	TypeSuccess NotificationType = "success"
)

// Notification represents a macOS notification
type Notification struct {
	Title      string
	Message    string
	Type       NotificationType
	AppName    string
}

// NotificationManager handles sending notifications to the user
type NotificationManager struct {
	appName string
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(appName string) *NotificationManager {
	return &NotificationManager{
		appName: appName,
	}
}

// Send sends a notification to the user via macOS notification center
func (nm *NotificationManager) Send(notification *Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	// Use osascript to send notification via macOS notification center
	script := fmt.Sprintf(
		`display notification "%s" with title "%s"`,
		notification.Message,
		notification.Title,
	)

	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// SendInfo sends an informational notification
func (nm *NotificationManager) SendInfo(title, message string) error {
	return nm.Send(&Notification{
		Title:   title,
		Message: message,
		Type:    TypeInfo,
	})
}

// SendWarning sends a warning notification
func (nm *NotificationManager) SendWarning(title, message string) error {
	return nm.Send(&Notification{
		Title:   title,
		Message: message,
		Type:    TypeWarning,
	})
}

// SendError sends an error notification
func (nm *NotificationManager) SendError(title, message string) error {
	return nm.Send(&Notification{
		Title:   title,
		Message: message,
		Type:    TypeError,
	})
}

// SendSuccess sends a success notification
func (nm *NotificationManager) SendSuccess(title, message string) error {
	return nm.Send(&Notification{
		Title:   title,
		Message: message,
		Type:    TypeSuccess,
	})
}

// RecordingStarted sends a notification that recording has started
func (nm *NotificationManager) RecordingStarted() error {
	return nm.SendInfo(nm.appName, "録音が開始されました")
}

// RecordingStopped sends a notification that recording has stopped
func (nm *NotificationManager) RecordingStopped() error {
	return nm.SendInfo(nm.appName, "録音が停止されました")
}

// TranscriptionComplete sends a notification that transcription is complete
func (nm *NotificationManager) TranscriptionComplete() error {
	return nm.SendSuccess(nm.appName, "文字起こしが完了しました")
}

// PasteComplete sends a notification that text has been pasted
func (nm *NotificationManager) PasteComplete() error {
	return nm.SendSuccess(nm.appName, "テキストが貼り付けられました")
}

// MicrophonePermissionDenied sends a notification that microphone permission is denied
func (nm *NotificationManager) MicrophonePermissionDenied() error {
	return nm.SendError(
		nm.appName,
		"マイクへのアクセスが拒否されました。システム設定で許可してください。",
	)
}

// AccessibilityPermissionDenied sends a notification that accessibility permission is denied
func (nm *NotificationManager) AccessibilityPermissionDenied() error {
	return nm.SendError(
		nm.appName,
		"アクセシビリティ権限が拒否されました。システム設定で許可してください。",
	)
}

// RecordingFailed sends a notification that recording failed
func (nm *NotificationManager) RecordingFailed(reason string) error {
	message := "録音に失敗しました"
	if reason != "" {
		message += "：" + reason
	}
	return nm.SendError(nm.appName, message)
}

// TranscriptionFailed sends a notification that transcription failed
func (nm *NotificationManager) TranscriptionFailed(reason string) error {
	message := "文字起こしに失敗しました"
	if reason != "" {
		message += "：" + reason
	}
	return nm.SendError(nm.appName, message)
}

// RecordingTimeExceeded sends a notification that recording time has exceeded the limit
func (nm *NotificationManager) RecordingTimeExceeded() error {
	return nm.SendWarning(
		nm.appName,
		"録音が60秒に達したため、自動停止しました。",
	)
}

// DeviceNotFound sends a notification that audio device is not found
func (nm *NotificationManager) DeviceNotFound() error {
	return nm.SendError(
		nm.appName,
		"オーディオデバイスが見つかりません。デバイスを再接続してください。",
	)
}

// ModelNotFound sends a notification that the model file is not found
func (nm *NotificationManager) ModelNotFound(modelPath string) error {
	message := fmt.Sprintf("モデルファイルが見つかりません: %s", modelPath)
	return nm.SendError(nm.appName, message)
}
