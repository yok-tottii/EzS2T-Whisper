package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Language represents a supported language
type Language string

const (
	// Japanese language
	LanguageJapanese Language = "ja"
	// English language
	LanguageEnglish Language = "en"
)

// Translator manages translations for the application
type Translator struct {
	currentLanguage Language
	translations    map[Language]map[string]string
	mu              sync.RWMutex
}

// NewTranslator creates a new translator with default language
func NewTranslator(language Language) *Translator {
	return &Translator{
		currentLanguage: language,
		translations:    make(map[Language]map[string]string),
	}
}

// LoadTranslations loads translations from JSON data
func (t *Translator) LoadTranslations(language Language, data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to unmarshal translations: %w", err)
	}

	t.translations[language] = translations
	return nil
}

// LoadTranslationsFromFile loads translations from a JSON file
func (t *Translator) LoadTranslationsFromFile(language Language, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read translation file: %w", err)
	}

	return t.LoadTranslations(language, data)
}

// SetLanguage sets the current language
func (t *Translator) SetLanguage(language Language) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.currentLanguage = language
}

// GetLanguage returns the current language
func (t *Translator) GetLanguage() Language {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentLanguage
}

// Translate translates a key in the current language
func (t *Translator) Translate(key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if translations, ok := t.translations[t.currentLanguage]; ok {
		if text, ok := translations[key]; ok {
			return text
		}
	}

	// Fallback to English if translation not found
	if t.currentLanguage != LanguageEnglish {
		if translations, ok := t.translations[LanguageEnglish]; ok {
			if text, ok := translations[key]; ok {
				return text
			}
		}
	}

	// Return key itself if no translation found
	return key
}

// TranslateWithFormat translates a key and formats with parameters
func (t *Translator) TranslateWithFormat(key string, params map[string]string) string {
	text := t.Translate(key)

	// Simple string replacement for parameters
	for param, value := range params {
		placeholder := fmt.Sprintf("{%s}", param)
		text = strings.ReplaceAll(text, placeholder, value)
	}

	return text
}

// GetAllTranslations returns all translations for the current language
func (t *Translator) GetAllTranslations() map[string]string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if translations, ok := t.translations[t.currentLanguage]; ok {
		// Return a copy to prevent external modifications
		result := make(map[string]string)
		for k, v := range translations {
			result[k] = v
		}
		return result
	}

	return make(map[string]string)
}

// HasTranslation checks if a translation key exists
func (t *Translator) HasTranslation(key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if translations, ok := t.translations[t.currentLanguage]; ok {
		_, ok := translations[key]
		return ok
	}

	return false
}

// ValidateLanguage validates that a language is supported
func ValidateLanguage(language string) bool {
	return language == string(LanguageJapanese) || language == string(LanguageEnglish)
}

// DetectSystemLanguage attempts to detect the system language
// For now, returns Japanese as default for macOS Japanese users
func DetectSystemLanguage() Language {
	// In a real implementation, we would check system locale
	// For now, default to English
	return LanguageEnglish
}

// GetSupportedLanguages returns a list of supported languages
func GetSupportedLanguages() []Language {
	return []Language{LanguageJapanese, LanguageEnglish}
}

// T is a convenience function for quick translation (assumes global translator)
// This would be set up in main.go
var GlobalTranslator *Translator

// T translates using the global translator
func T(key string) string {
	if GlobalTranslator == nil {
		return key
	}
	return GlobalTranslator.Translate(key)
}

// TF translates with formatting using the global translator
func TF(key string, params map[string]string) string {
	if GlobalTranslator == nil {
		return key
	}
	return GlobalTranslator.TranslateWithFormat(key, params)
}

// DefaultEnglishTranslations returns default English translations
func DefaultEnglishTranslations() map[string]string {
	return map[string]string{
		// Menu items
		"menu.settings":        "Open Settings...",
		"menu.rescan_models":   "Rescan Models",
		"menu.test_recording":  "Test Recording",
		"menu.about":           "About",
		"menu.quit":            "Quit",

		// Settings
		"settings.title":              "EzS2T-Whisper Settings",
		"settings.hotkey":             "Hotkey",
		"settings.recording_mode":     "Recording Mode",
		"settings.model":              "Model",
		"settings.language":           "Language",
		"settings.audio_device":       "Audio Device",
		"settings.ui_language":        "UI Language",
		"settings.save":               "Save",

		// Permissions
		"permission.microphone":     "Microphone",
		"permission.accessibility": "Accessibility",
		"permission.granted":       "✓ Granted",
		"permission.denied":        "✗ Denied",
		"permission.request":       "Open Settings",

		// Errors
		"error.mic_permission_denied":         "Microphone access denied",
		"error.accessibility_permission_denied": "Accessibility permission denied",
		"error.recording_failed":              "Recording failed",
		"error.transcription_failed":          "Transcription failed",

		// Notifications
		"notification.recording_started": "Recording started",
		"notification.recording_stopped": "Recording stopped",
		"notification.transcription_complete": "Transcription complete",
		"notification.paste_complete":   "Text pasted",

		// Status
		"status.idle":       "Idle",
		"status.recording":  "Recording",
		"status.processing": "Processing",
	}
}

// DefaultJapaneseTranslations returns default Japanese translations
func DefaultJapaneseTranslations() map[string]string {
	return map[string]string{
		// Menu items
		"menu.settings":        "設定を開く...",
		"menu.rescan_models":   "モデルを再スキャン",
		"menu.test_recording":  "録音テスト",
		"menu.about":           "バージョン情報",
		"menu.quit":            "終了",

		// Settings
		"settings.title":              "EzS2T-Whisper 設定",
		"settings.hotkey":             "ホットキー",
		"settings.recording_mode":     "録音モード",
		"settings.model":              "モデル",
		"settings.language":           "言語",
		"settings.audio_device":       "オーディオデバイス",
		"settings.ui_language":        "UI言語",
		"settings.save":               "保存",

		// Permissions
		"permission.microphone":     "マイク",
		"permission.accessibility": "アクセシビリティ",
		"permission.granted":       "✓ 許可済み",
		"permission.denied":        "✗ 拒否",
		"permission.request":       "設定を開く",

		// Errors
		"error.mic_permission_denied":         "マイクへのアクセスが拒否されました",
		"error.accessibility_permission_denied": "アクセシビリティ権限が拒否されました",
		"error.recording_failed":              "録音に失敗しました",
		"error.transcription_failed":          "文字起こしに失敗しました",

		// Notifications
		"notification.recording_started": "録音が開始されました",
		"notification.recording_stopped": "録音が停止されました",
		"notification.transcription_complete": "文字起こしが完了しました",
		"notification.paste_complete":   "テキストが貼り付けられました",

		// Status
		"status.idle":       "待機中",
		"status.recording":  "録音中",
		"status.processing": "処理中",
	}
}
