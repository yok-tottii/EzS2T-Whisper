package i18n

import (
	"testing"
)

func TestNewTranslator(t *testing.T) {
	translator := NewTranslator(LanguageJapanese)

	if translator == nil {
		t.Fatal("Expected translator to be created")
	}

	if translator.GetLanguage() != LanguageJapanese {
		t.Errorf("Expected language to be ja, got %s", translator.GetLanguage())
	}
}

func TestLoadTranslations(t *testing.T) {
	translator := NewTranslator(LanguageJapanese)

	jaData := []byte(`{
		"menu.settings": "設定を開く...",
		"menu.quit": "終了"
	}`)

	err := translator.LoadTranslations(LanguageJapanese, jaData)
	if err != nil {
		t.Fatalf("Failed to load translations: %v", err)
	}

	text := translator.Translate("menu.settings")
	if text != "設定を開く..." {
		t.Errorf("Expected '設定を開く...', got '%s'", text)
	}
}

func TestSetLanguage(t *testing.T) {
	translator := NewTranslator(LanguageEnglish)

	translator.SetLanguage(LanguageJapanese)

	if translator.GetLanguage() != LanguageJapanese {
		t.Errorf("Expected language to be ja, got %s", translator.GetLanguage())
	}
}

func TestTranslate(t *testing.T) {
	translator := NewTranslator(LanguageEnglish)

	enData := []byte(`{
		"menu.quit": "Quit"
	}`)

	translator.LoadTranslations(LanguageEnglish, enData)

	text := translator.Translate("menu.quit")
	if text != "Quit" {
		t.Errorf("Expected 'Quit', got '%s'", text)
	}
}

func TestTranslateFallback(t *testing.T) {
	translator := NewTranslator(LanguageJapanese)

	enData := []byte(`{
		"menu.quit": "Quit"
	}`)

	// Only load English translations
	translator.LoadTranslations(LanguageEnglish, enData)

	// When asking for Japanese translation (which doesn't exist),
	// should fall back to English
	text := translator.Translate("menu.quit")
	if text != "Quit" {
		t.Errorf("Expected 'Quit' (fallback), got '%s'", text)
	}
}

func TestTranslateNotFound(t *testing.T) {
	translator := NewTranslator(LanguageEnglish)

	// When key doesn't exist, should return the key itself
	text := translator.Translate("nonexistent.key")
	if text != "nonexistent.key" {
		t.Errorf("Expected 'nonexistent.key', got '%s'", text)
	}
}

func TestTranslateWithFormat(t *testing.T) {
	translator := NewTranslator(LanguageEnglish)

	enData := []byte(`{
		"greeting": "Hello, {name}!"
	}`)

	translator.LoadTranslations(LanguageEnglish, enData)

	text := translator.TranslateWithFormat("greeting", map[string]string{
		"name": "World",
	})

	if text != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", text)
	}
}

func TestGetAllTranslations(t *testing.T) {
	translator := NewTranslator(LanguageEnglish)

	enData := []byte(`{
		"menu.quit": "Quit",
		"menu.settings": "Settings"
	}`)

	translator.LoadTranslations(LanguageEnglish, enData)

	translations := translator.GetAllTranslations()

	if len(translations) != 2 {
		t.Errorf("Expected 2 translations, got %d", len(translations))
	}

	if translations["menu.quit"] != "Quit" {
		t.Errorf("Expected 'Quit', got '%s'", translations["menu.quit"])
	}
}

func TestHasTranslation(t *testing.T) {
	translator := NewTranslator(LanguageEnglish)

	enData := []byte(`{
		"menu.quit": "Quit"
	}`)

	translator.LoadTranslations(LanguageEnglish, enData)

	if !translator.HasTranslation("menu.quit") {
		t.Error("Expected translation 'menu.quit' to exist")
	}

	if translator.HasTranslation("nonexistent.key") {
		t.Error("Expected translation 'nonexistent.key' to not exist")
	}
}

func TestValidateLanguage(t *testing.T) {
	tests := []struct {
		language string
		expected bool
	}{
		{"ja", true},
		{"en", true},
		{"fr", false},
		{"de", false},
		{"", false},
	}

	for _, test := range tests {
		result := ValidateLanguage(test.language)
		if result != test.expected {
			t.Errorf("ValidateLanguage(%s) = %v, expected %v", test.language, result, test.expected)
		}
	}
}

func TestDetectSystemLanguage(t *testing.T) {
	language := DetectSystemLanguage()

	if language != LanguageEnglish && language != LanguageJapanese {
		t.Errorf("Expected ja or en, got %s", language)
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	languages := GetSupportedLanguages()

	if len(languages) != 2 {
		t.Errorf("Expected 2 supported languages, got %d", len(languages))
	}

	hasJapanese := false
	hasEnglish := false

	for _, lang := range languages {
		if lang == LanguageJapanese {
			hasJapanese = true
		}
		if lang == LanguageEnglish {
			hasEnglish = true
		}
	}

	if !hasJapanese {
		t.Error("Expected Japanese to be supported")
	}

	if !hasEnglish {
		t.Error("Expected English to be supported")
	}
}

func TestDefaultEnglishTranslations(t *testing.T) {
	translations := DefaultEnglishTranslations()

	if len(translations) == 0 {
		t.Error("Expected default English translations to be returned")
	}

	if translations["menu.quit"] != "Quit" {
		t.Error("Expected 'menu.quit' translation to be 'Quit'")
	}
}

func TestDefaultJapaneseTranslations(t *testing.T) {
	translations := DefaultJapaneseTranslations()

	if len(translations) == 0 {
		t.Error("Expected default Japanese translations to be returned")
	}

	if translations["menu.quit"] != "終了" {
		t.Error("Expected 'menu.quit' translation to be '終了'")
	}
}

func TestLanguageConstant(t *testing.T) {
	if string(LanguageJapanese) != "ja" {
		t.Errorf("Expected 'ja', got '%s'", LanguageJapanese)
	}

	if string(LanguageEnglish) != "en" {
		t.Errorf("Expected 'en', got '%s'", LanguageEnglish)
	}
}

func TestConcurrentTranslation(t *testing.T) {
	translator := NewTranslator(LanguageEnglish)

	enData := []byte(`{
		"test.key": "value"
	}`)

	translator.LoadTranslations(LanguageEnglish, enData)

	done := make(chan bool, 10)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			text := translator.Translate("test.key")
			if text != "value" {
				t.Errorf("Expected 'value', got '%s'", text)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGlobalTranslator(t *testing.T) {
	// Initialize global translator
	GlobalTranslator = NewTranslator(LanguageEnglish)

	enData := []byte(`{
		"test.key": "value"
	}`)

	GlobalTranslator.LoadTranslations(LanguageEnglish, enData)

	// Test T function
	text := T("test.key")
	if text != "value" {
		t.Errorf("Expected 'value', got '%s'", text)
	}

	// Test TF function
	enData2 := []byte(`{
		"greeting": "Hello, {name}!"
	}`)

	GlobalTranslator.LoadTranslations(LanguageEnglish, enData2)

	text = TF("greeting", map[string]string{"name": "World"})
	if text != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", text)
	}

	// Clean up
	GlobalTranslator = nil
}
