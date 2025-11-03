package audio

import (
	"fmt"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

// PortAudioDriver implements AudioDriver using PortAudio
type PortAudioDriver struct {
	config    Config
	stream    *portaudio.Stream
	buffer    []int16
	mu        sync.Mutex
	recording bool
	initialized bool
}

// NewPortAudioDriver creates a new PortAudio driver
func NewPortAudioDriver() (*PortAudioDriver, error) {
	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %w", err)
	}

	return &PortAudioDriver{
		buffer: make([]int16, 0, 1024*1024), // Pre-allocate 1MB buffer
	}, nil
}

// ListDevices returns a list of available audio input devices
func (d *PortAudioDriver) ListDevices() ([]Device, error) {
	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	defaultInput, err := portaudio.DefaultInputDevice()
	if err != nil {
		// If we can't get the default device, continue without marking any as default
		defaultInput = nil
	}

	var result []Device
	for i, dev := range devices {
		// Only include devices with input channels
		if dev.MaxInputChannels > 0 {
			isDefault := false
			if defaultInput != nil && dev.Name == defaultInput.Name {
				isDefault = true
			}

			result = append(result, Device{
				ID:        i,
				Name:      dev.Name,
				IsDefault: isDefault,
			})
		}
	}

	return result, nil
}

// Initialize initializes the audio driver with the given configuration
func (d *PortAudioDriver) Initialize(config Config) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.recording {
		return fmt.Errorf("cannot initialize while recording")
	}

	// Close existing stream if any
	if d.stream != nil {
		if err := d.stream.Close(); err != nil {
			return fmt.Errorf("failed to close existing stream: %w", err)
		}
		d.stream = nil
	}

	// Get the device
	var device *portaudio.DeviceInfo
	var err error

	if config.DeviceID == -1 {
		// Use default input device
		device, err = portaudio.DefaultInputDevice()
		if err != nil {
			return fmt.Errorf("failed to get default input device: %w", err)
		}
	} else {
		// Use specified device
		devices, err := portaudio.Devices()
		if err != nil {
			return fmt.Errorf("failed to list devices: %w", err)
		}

		if config.DeviceID < 0 || config.DeviceID >= len(devices) {
			return fmt.Errorf("invalid device ID: %d", config.DeviceID)
		}

		device = devices[config.DeviceID]
	}

	// Validate device has input channels
	if device.MaxInputChannels <= 0 {
		return fmt.Errorf("selected device '%s' (ID: %d) has no input channels (output-only device)",
			device.Name, config.DeviceID)
	}

	// Set latency
	var latency time.Duration
	switch config.Latency {
	case LowLatency:
		latency = device.DefaultLowInputLatency
	case HighStability:
		latency = device.DefaultHighInputLatency
	default:
		latency = device.DefaultHighInputLatency
	}

	// Create stream parameters
	streamParams := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: config.Channels,
			Latency:  latency,
		},
		SampleRate:      float64(config.SampleRate),
		FramesPerBuffer: 1024,
	}

	// Open stream
	stream, err := portaudio.OpenStream(streamParams, d.callback)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	d.stream = stream
	d.config = config
	d.initialized = true

	return nil
}

// callback is called by PortAudio when audio data is available
func (d *PortAudioDriver) callback(in []int16) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.recording {
		d.buffer = append(d.buffer, in...)
	}
}

// StartRecording starts recording audio
func (d *PortAudioDriver) StartRecording() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.initialized {
		return fmt.Errorf("driver not initialized")
	}

	if d.recording {
		return fmt.Errorf("already recording")
	}

	// Clear buffer
	d.buffer = d.buffer[:0]

	// Start stream
	if err := d.stream.Start(); err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}

	d.recording = true
	return nil
}

// StopRecording stops recording and returns the recorded audio data
func (d *PortAudioDriver) StopRecording() ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.recording {
		return nil, fmt.Errorf("not recording")
	}

	// Stop stream
	if err := d.stream.Stop(); err != nil {
		return nil, fmt.Errorf("failed to stop stream: %w", err)
	}

	d.recording = false

	// Convert int16 buffer to bytes
	data := make([]byte, len(d.buffer)*2)
	for i, sample := range d.buffer {
		data[i*2] = byte(sample)
		data[i*2+1] = byte(sample >> 8)
	}

	return data, nil
}

// IsRecording returns whether recording is currently active
func (d *PortAudioDriver) IsRecording() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.recording
}

// Close releases all resources
func (d *PortAudioDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Stop recording if active
	if d.recording {
		if err := d.stream.Stop(); err != nil {
			return fmt.Errorf("failed to stop stream: %w", err)
		}
		d.recording = false
	}

	// Close stream
	if d.stream != nil {
		if err := d.stream.Close(); err != nil {
			return fmt.Errorf("failed to close stream: %w", err)
		}
		d.stream = nil
	}

	// Terminate PortAudio
	if err := portaudio.Terminate(); err != nil {
		return fmt.Errorf("failed to terminate PortAudio: %w", err)
	}

	d.initialized = false
	return nil
}
