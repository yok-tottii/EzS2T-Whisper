package audio

// Device represents an audio input device
type Device struct {
	ID        int
	Name      string
	IsDefault bool
}

// LatencyMode defines the latency priority
type LatencyMode int

const (
	// LowLatency prioritizes low latency (real-time)
	LowLatency LatencyMode = iota
	// HighStability prioritizes stability (larger buffer)
	HighStability
)

// Config holds audio configuration
type Config struct {
	DeviceID   int
	SampleRate int
	Channels   int
	Latency    LatencyMode
}

// DefaultConfig returns the default audio configuration
// Sample rate: 16kHz (Whisper recommended)
// Channels: 1 (mono)
// Latency: HighStability
func DefaultConfig() Config {
	return Config{
		DeviceID:   -1, // -1 means use default device
		SampleRate: 16000,
		Channels:   1,
		Latency:    HighStability,
	}
}

// AudioDriver is the interface for audio input
// This abstraction allows for future replacement of PortAudio with other libraries (e.g., miniaudio)
type AudioDriver interface {
	// ListDevices returns a list of available audio input devices
	ListDevices() ([]Device, error)

	// Initialize initializes the audio driver with the given configuration
	Initialize(config Config) error

	// StartRecording starts recording audio
	StartRecording() error

	// StopRecording stops recording and returns the recorded audio data (PCM format)
	StopRecording() ([]byte, error)

	// IsRecording returns whether recording is currently active
	IsRecording() bool

	// Close releases all resources
	Close() error
}
