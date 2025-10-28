# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**EzS2T-Whisper** is a macOS-exclusive speech-to-text (STT) application that provides fast, offline local speech recognition using Whisper.cpp. The application is designed to run as a system tray resident, allowing users to press a hotkey (default: Ctrl+Option+Space) to record audio and have it automatically transcribed and pasted into any application.

**Key characteristics:**
- Pure offline operation (no internet required)
- Japanese-first language support with English UI option
- Apple Silicon (M1+) optimized, Intel Mac compatible
- MIT License, individual open-source project
- Comprehensive specification document in `specs/EzS2T-Whisper仕様書_v2.md`

## Build and Development Commands

### Initial Setup

```bash
# Install Xcode command-line tools (required)
xcode-select --install

# Install required system libraries via Homebrew
brew install libpng libjpeg portaudio

# Install Go dependencies
go mod download
```

### Building

```bash
# Development build
go build -o ezs2t-whisper

# Release build (binary size reduction)
go build -ldflags="-s -w" -o ezs2t-whisper
```

### Running

```bash
# Run the application (will launch in system tray)
./ezs2t-whisper
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestName ./package

# Run tests with coverage
go test -cover ./...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
go vet ./...

# Check for common mistakes (install golangci-lint first)
golangci-lint run ./...
```

## High-Level Architecture

The application follows a modular Go package structure with clear separation of concerns:

### Core Packages

| Package | Purpose | Key Dependencies |
|---------|---------|---|
| **main** | Application entry point and initialization | All packages |
| **hotkey** | Global hotkey detection and registration | golang-design/hotkey |
| **audio** | Audio input abstraction (AudioDriver interface) | gordonklaus/portaudio |
| **recording** | Audio recording logic and state management | audio package |
| **recognition** | Whisper.cpp integration and transcription | Whisper.cpp (CGO) |
| **clipboard** | Safe clipboard operations using changeCount method | robotgo |
| **tray** | System tray menu integration | getlantern/systray |
| **server** | Local HTTP server for settings UI | net/http, embed |
| **api** | REST API implementation for settings | server, config |
| **config** | Settings persistence and management | stdlib (JSON) |
| **i18n** | Multi-language support (ja.json, en.json) | Custom or stdlib |
| **permissions** | macOS system permission checks | CGO (Objective-C) |

### Key Design Patterns

1. **AudioDriver Interface**: Abstract interface for audio input to allow future replacement (PortAudio → miniaudio)
2. **Goroutine-based Concurrency**: Separate goroutines for hotkey monitoring, recording, and transcription
3. **changeCount-based Clipboard Safety**: Mechanism to restore clipboard state only if not modified externally
4. **Embedded Web Assets**: Go `embed` package compiles frontend resources into the binary
5. **REST API**: HTTP API with JSON request/response for settings management

### Data Flow

```
Hotkey Press
  ↓
Audio Recording (PortAudio)
  ↓
Whisper.cpp Transcription
  ↓
changeCount-safe Clipboard Insertion (robotgo)
  ↓
Text Pasted into Active Application
```

## Frontend Structure

The settings UI is a lightweight embedded web application:

```
frontend/ (planned)
├── index.html              # Main settings page
├── css/
│   └── style.css          # Styling
├── js/
│   ├── app.js             # Main logic
│   └── api.js             # API client
└── i18n/
    ├── ja.json            # Japanese strings
    └── en.json            # English strings
```

**Technology:**
- Embedding: Go `embed` package
- Server: Go `net/http` (no external web framework)
- Frontend: Vanilla JavaScript (no heavy frameworks)
- Security: localhost-only access

## Configuration and Persistence

**Config Location**: `~/Library/Application Support/EzS2T-Whisper/`

**Files:**
- `config.json` - Application settings (hotkey, audio device, model path, language, etc.)
- `models/` - Directory for Whisper model files (.gguf format)
- `logs/` - Daily log files with 7-day retention

**Models:**
- **Default**: `ggml-large-v3-turbo-q5_0.gguf` (~1.5 GB, precision priority)
- **Lightweight**: `ggml-small-q5_1.gguf` (~200 MB, battery-efficient)

## REST API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/settings` | Fetch current configuration |
| PUT | `/api/settings` | Update configuration |
| POST | `/api/hotkey/validate` | Check hotkey conflicts |
| POST | `/api/hotkey/register` | Register new hotkey |
| GET | `/api/devices` | List audio devices |
| GET | `/api/models` | List available models |
| POST | `/api/models/rescan` | Rescan models directory |
| POST | `/api/test/record` | Test record→transcribe→paste pipeline |
| GET | `/api/permissions` | Check system permissions status |

## Key Specifications and Requirements

### Hotkey Handling

**Default**: Ctrl (⌃) + Option (⌥) + Space (␣)

- **Storage**: Physical scan codes (keyboard layout independent)
- **Display**: Localized labels (ja/en)
- **Modes**: Press-to-hold (default) or toggle mode
- **Conflict Detection**: Check against Spotlight, Alfred, Raycast, IME, system shortcuts

**Implementation Note**: Use `golang-design/hotkey` with conflict detection logic.

### Audio Recording

- **Maximum Duration**: 60 seconds
- **Recording Modes**:
  - Press-to-hold: Records while key is held
  - Toggle: 1st press starts, 2nd press stops
- **Device Selection**: Default to system mic, allow user selection via PortAudio

### Text Insertion (Critical)

**changeCount Method**:
1. Save clipboard state via `NSPasteboard.generalPasteboard().changeCount`
2. Copy transcribed text to clipboard
3. Send `Cmd+V` key event (robotgo)
4. Check if changeCount matches after paste
5. **Only restore original clipboard if changeCount unchanged** (user didn't copy during transcription)

This prevents data loss if the user copies something during the transcription process.

### System Permissions

**Required Permissions**:
1. **Microphone**: For audio recording
2. **Accessibility**: For hotkey monitoring and text insertion via robotgo

**User Guidance**:
- Show permission status in settings UI (✓ Granted / ✗ Denied)
- Provide "Open System Settings" button pointing to Privacy & Security
- Disable recording/pasting features if permissions not granted
- Show warning icon in system tray if permissions missing

### Multi-Language Support (i18n)

**Supported Languages**: Japanese (ja) and English (en)

**Implementation**:
- Use JSON dictionaries in `frontend/i18n/` or `internal/i18n/`
- Key-based translation (e.g., `error.mic_permission_denied`)
- Auto-detect OS language, allow manual override in settings
- Localize:
  - Menu bar items
  - Settings UI labels and descriptions
  - Error/notification messages
  - System permission descriptions in `Info.plist`

### Logging

**Log Location**: `~/Library/Application Support/EzS2T-Whisper/logs/`

**File Format**: `ezs2t-whisper-YYYYMMDD.log`

**Levels**:
- `INFO`: Startup, config changes, recording events
- `WARN`: Permission issues, device errors
- `ERROR`: Crashes, unexpected failures
- `DEBUG`: Detailed logs (only when debug mode enabled)

**Important**: Do NOT log audio content, transcribed text, or PII. Log only operational events.

**Retention**: Keep 7 days of logs, delete older ones.

## Performance Targets

**Real-Time Factor (RTF)** = processing_time / audio_duration

- **Target RTF < 1.0**: Process 10 seconds of audio in 10 seconds
- **Apple Silicon M1+ Goal**: RTF < 0.5 (process faster than real-time)
- **Intel Mac**: RTF < 1.5 acceptable

## Acceptance Criteria (MVP)

From the specification, the project is complete when:

1. **Accuracy**: 10-second Japanese test audio achieves <5% error rate with proper punctuation
2. **Performance**: RTF < 1.0 on Apple Silicon M1+
3. **Permissions**: Recording/pasting disabled when permissions missing, with clear user guidance
4. **Hotkey**: Ctrl+Opt+Space works without conflicts in typical macOS environment
5. **Clipboard**: changeCount method correctly restores clipboard when user intervenes
6. **Stability**: 100 consecutive record→transcribe→paste cycles without crashes or memory leaks

## Development Roadmap

**Planned 4-Week Schedule**:

**Week 1**: Hotkey detection, audio recording, recording modes
**Week 2**: Whisper.cpp integration, model loading, changeCount-safe text insertion
**Week 3**: System tray menu, settings web UI, REST API
**Week 4**: i18n, permission handling, setup wizard, testing, documentation

## Important Implementation Notes

### robotgo Considerations

- **Binary Size**: ~20-30 MB (due to dependencies, acceptable for personal use)
- **Japanese Input**: Use clipboard + paste, NOT direct character input (avoids mojibake)
- **Permissions**: Accessibility permission required; failures may be silent without explicit checks
- **Thread Safety**: macOS UI operations should run on main thread

### CGO and Build Environment

- **Xcode Required**: Cannot cross-compile; build on macOS for macOS
- **Dependencies**: libpng, libjpeg, portaudio (install via Homebrew)
- **Linking**: Ensure proper linking to system frameworks (Cocoa, Foundation)

### Whisper.cpp Integration

- **Model Format**: GGUF (.gguf files)
- **Task Mode**: Use `task=transcribe` (ASR only, no translation)
- **Language**: Set to Japanese (ja) by default
- **Inference**: Runs locally on CPU/GPU without external API calls

### macOS System Integration

- **System Tray**: Use `getlantern/systray` for menu bar integration
- **Permissions**: Check at startup and periodically; gracefully disable features if missing
- **Info.plist**: Must include localized permission descriptions for mic and accessibility
- **Accessibility**: Required for both hotkey monitoring and robotgo text insertion

## Code Organization Principles

1. **Pure Go Priority**: Avoid Objective-C/Swift; use CGO wrappers (robotgo, PortAudio, etc.)
2. **Interface Abstraction**: Use interfaces for major components (AudioDriver, Logger, Config)
3. **Error Handling**: Return errors explicitly; log with context; never silently fail on permissions
4. **Testing**: Unit tests for each package; integration tests for critical pipelines
5. **Concurrency**: Use goroutines; avoid shared state; use channels for communication
6. **Configuration**: Single source of truth in config file; validate on load and update

## Testing Guidelines

**Unit Tests** (`*_test.go`):
- Mock PortAudio for audio tests
- Mock Whisper for recognition tests
- Test clipboard changeCount logic thoroughly

**Integration Tests**:
- Full pipeline: record → transcribe → insert
- Permission handling scenarios
- Long-text splitting and pasting
- Clipboard state restoration edge cases

**Manual Testing**:
- Test hotkey registration in different macOS versions
- Verify permission dialogs and system settings redirects
- Test with different audio devices
- Test with both Japanese and English languages
- Check memory usage during long-running sessions

## References and Dependencies

**Key Libraries**:
- `ggerganov/whisper.cpp`: https://github.com/ggerganov/whisper.cpp (speech recognition)
- `getlantern/systray`: https://github.com/getlantern/systray (system tray)
- `gordonklaus/portaudio`: https://github.com/gordonklaus/portaudio (audio input)
- `golang-design/hotkey`: https://github.com/golang-design/hotkey (hotkey handling)
- `go-vgo/robotgo`: https://github.com/go-vgo/robotgo (clipboard, keyboard control)

**Documentation**:
- Full specification: `specs/EzS2T-Whisper仕様書_v2.md` (Japanese)
- Go documentation: https://golang.org/doc
- macOS SDK: https://developer.apple.com/documentation

## Common Tasks

### Adding a New API Endpoint

1. Define request/response structs in `api` package
2. Implement handler function in `server` package
3. Register route in HTTP router (typically in `server.go`)
4. Add tests in `*_test.go` file
5. Document in `CLAUDE.md` REST API table

### Updating Configuration Schema

1. Modify struct in `config/config.go`
2. Update default values in config initialization
3. Update REST API validation in `api` package
4. Update settings UI in `frontend/index.html`
5. Test config persistence and migration if needed

### Adding New Language

1. Create new JSON dictionary file (`frontend/i18n/xx.json`)
2. Add language option to settings UI
3. Update i18n loader to support new language code
4. Localize `Info.plist` strings if needed
5. Test UI in new language

### Debugging Permission Issues

1. Check system permissions: Settings → Privacy & Security → Microphone/Accessibility
2. Use `mdfind` to verify app is in accessibility approved list
3. Test permission check functions manually
4. Ensure `Info.plist` has correct permission description keys
5. Log permission check results at startup

## Specifications
@specs/prd.md

## Rules you must follow
You must think in English and respond in Japanese.
