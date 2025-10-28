package recognition

/*
#cgo CFLAGS: -I${SRCDIR}/../../whisper.cpp/include -I${SRCDIR}/../../whisper.cpp/ggml/include
#cgo LDFLAGS: -L${SRCDIR}/../../whisper.cpp/build/src -L${SRCDIR}/../../whisper.cpp/build/ggml/src -lwhisper -lggml -lm -Wl,-rpath,${SRCDIR}/../../whisper.cpp/build/src -Wl,-rpath,${SRCDIR}/../../whisper.cpp/build/ggml/src
#include "whisper.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unsafe"
)

// Recognizer is the interface for speech recognition
type Recognizer interface {
	LoadModel(modelPath string) error
	Transcribe(audioData []byte, sampleRate int) (string, error)
	Close() error
}

// WhisperRecognizer implements Recognizer using Whisper.cpp
type WhisperRecognizer struct {
	ctx      *C.struct_whisper_context
	mu       sync.Mutex
	language string
}

// Config holds recognition configuration
type Config struct {
	Language string // Default: "ja"
	Threads  int    // Number of threads, 0 = auto
}

// DefaultConfig returns the default recognition configuration
func DefaultConfig() Config {
	return Config{
		Language: "ja",
		Threads:  0, // Auto-detect
	}
}

// NewWhisperRecognizer creates a new Whisper recognizer
func NewWhisperRecognizer(config Config) *WhisperRecognizer {
	return &WhisperRecognizer{
		language: config.Language,
	}
}

// LoadModel loads a Whisper model from the specified path
func (r *WhisperRecognizer) LoadModel(modelPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if model file exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", modelPath)
	}

	// Convert Go string to C string
	cModelPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cModelPath))

	// Load the model
	ctx := C.whisper_init_from_file(cModelPath)
	if ctx == nil {
		return fmt.Errorf("failed to load model from: %s", modelPath)
	}

	// Close old context if exists
	if r.ctx != nil {
		C.whisper_free(r.ctx)
	}

	r.ctx = ctx
	return nil
}

// Transcribe performs speech recognition on the given audio data
func (r *WhisperRecognizer) Transcribe(audioData []byte, sampleRate int) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ctx == nil {
		return "", fmt.Errorf("model not loaded")
	}

	if len(audioData) == 0 {
		return "", fmt.Errorf("audio data is empty")
	}

	// Convert byte array to float32 samples
	// Assuming audioData is 16-bit PCM (2 bytes per sample)
	numSamples := len(audioData) / 2
	samples := make([]float32, numSamples)

	for i := 0; i < numSamples; i++ {
		// Convert 16-bit PCM to float32 in range [-1.0, 1.0]
		sample := int16(audioData[i*2]) | (int16(audioData[i*2+1]) << 8)
		samples[i] = float32(sample) / 32768.0
	}

	// Create whisper parameters
	params := C.whisper_full_default_params(C.WHISPER_SAMPLING_GREEDY)

	// Set language
	cLanguage := C.CString(r.language)
	defer C.free(unsafe.Pointer(cLanguage))
	params.language = cLanguage

	// Set task to transcribe (not translate)
	params.translate = C.bool(false)

	// Run inference
	result := C.whisper_full(
		r.ctx,
		params,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.int(numSamples),
	)

	if result != 0 {
		return "", fmt.Errorf("whisper_full failed with code: %d", result)
	}

	// Get the number of segments
	nSegments := C.whisper_full_n_segments(r.ctx)

	// Concatenate all segments
	var transcription string
	for i := 0; i < int(nSegments); i++ {
		text := C.whisper_full_get_segment_text(r.ctx, C.int(i))
		transcription += C.GoString(text)
	}

	return transcription, nil
}

// Close releases resources
func (r *WhisperRecognizer) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ctx != nil {
		C.whisper_free(r.ctx)
		r.ctx = nil
	}

	return nil
}

// GetDefaultModelPath returns the default path for Whisper models
func GetDefaultModelPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, "Library", "Application Support", "EzS2T-Whisper", "models")
}

// FindModel searches for a model file in the default model directory
func FindModel(modelName string) (string, error) {
	modelDir := GetDefaultModelPath()

	// Check if the model directory exists
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		return "", fmt.Errorf("model directory not found: %s", modelDir)
	}

	// Look for the model file
	modelPath := filepath.Join(modelDir, modelName)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("model file not found: %s", modelPath)
	}

	return modelPath, nil
}
