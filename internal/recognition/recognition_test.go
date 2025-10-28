package recognition

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Language != "ja" {
		t.Errorf("Expected default language 'ja', got '%s'", config.Language)
	}

	if config.Threads != 0 {
		t.Errorf("Expected default threads 0 (auto), got %d", config.Threads)
	}
}

func TestNewWhisperRecognizer(t *testing.T) {
	config := DefaultConfig()
	recognizer := NewWhisperRecognizer(config)

	if recognizer == nil {
		t.Fatal("Expected recognizer to be created")
	}

	if recognizer.language != "ja" {
		t.Errorf("Expected language 'ja', got '%s'", recognizer.language)
	}
}

func TestGetDefaultModelPath(t *testing.T) {
	modelPath := GetDefaultModelPath()

	if modelPath == "" {
		t.Error("Expected non-empty model path")
	}

	expectedSuffix := filepath.Join("Library", "Application Support", "EzS2T-Whisper", "models")
	if !filepath.IsAbs(modelPath) {
		t.Error("Expected absolute path")
	}

	if len(modelPath) < len(expectedSuffix) {
		t.Errorf("Model path too short: %s", modelPath)
	}
}

func TestFindModel_NonExistentDirectory(t *testing.T) {
	// Create a temporary directory and remove it
	tmpDir, err := os.MkdirTemp("", "ezs2t-test")
	if err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(tmpDir)

	// FindModel should return error for non-existent directory
	// Note: This test assumes the default model directory doesn't exist
	_, err = FindModel("nonexistent-model.gguf")
	if err == nil {
		t.Error("Expected error for non-existent model directory, got nil")
	}
}

func TestLoadModel_NonExistentFile(t *testing.T) {
	config := DefaultConfig()
	recognizer := NewWhisperRecognizer(config)
	defer recognizer.Close()

	err := recognizer.LoadModel("/nonexistent/path/model.gguf")
	if err == nil {
		t.Error("Expected error for non-existent model file, got nil")
	}
}

func TestTranscribe_ModelNotLoaded(t *testing.T) {
	config := DefaultConfig()
	recognizer := NewWhisperRecognizer(config)
	defer recognizer.Close()

	// Transcribe without loading model should fail
	audioData := make([]byte, 1000)
	_, err := recognizer.Transcribe(audioData, 16000)
	if err == nil {
		t.Error("Expected error when model not loaded, got nil")
	}
}

func TestTranscribe_EmptyAudio(t *testing.T) {
	config := DefaultConfig()
	recognizer := NewWhisperRecognizer(config)
	defer recognizer.Close()

	// Even without model, empty audio should fail
	audioData := []byte{}
	_, err := recognizer.Transcribe(audioData, 16000)
	if err == nil {
		t.Error("Expected error for empty audio data, got nil")
	}
}

func TestClose_WithoutModel(t *testing.T) {
	config := DefaultConfig()
	recognizer := NewWhisperRecognizer(config)

	err := recognizer.Close()
	if err != nil {
		t.Errorf("Expected nil error when closing without model, got: %v", err)
	}
}

// Note: Integration tests with actual model files should be in a separate test suite
// as they require downloading large model files
