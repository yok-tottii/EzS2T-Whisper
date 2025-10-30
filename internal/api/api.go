package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yok-tottii/EzS2T-Whisper/internal/audio"
	"github.com/yok-tottii/EzS2T-Whisper/internal/config"
	"github.com/yok-tottii/EzS2T-Whisper/internal/hotkey"
	"github.com/yok-tottii/EzS2T-Whisper/internal/wizard"
	hk "golang.design/x/hotkey"
)

// Handler manages API endpoints
type Handler struct {
	config          *config.Config
	wizard          *wizard.SetupWizard
	audioDriver     audio.AudioDriver
	onHotkeyChanged func() error // Callback to reload hotkey in main app
}

// New creates a new API handler
func New(cfg *config.Config, wiz *wizard.SetupWizard, onHotkeyChanged func() error) *Handler {
	return &Handler{
		config:          cfg,
		wizard:          wiz,
		audioDriver:     nil, // Will be set later via SetAudioDriver
		onHotkeyChanged: onHotkeyChanged,
	}
}

// SetAudioDriver sets the audio driver instance
// This is called after the audio driver is initialized in main.go
func (h *Handler) SetAudioDriver(driver audio.AudioDriver) {
	h.audioDriver = driver
}

// RegisterRoutes registers all API routes on the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/settings", h.handleSettings)
	mux.HandleFunc("/api/hotkey/validate", h.handleHotkeyValidate)
	mux.HandleFunc("/api/hotkey/register", h.handleHotkeyRegister)
	mux.HandleFunc("/api/devices", h.handleDevices)
	mux.HandleFunc("/api/models", h.handleModels)
	mux.HandleFunc("/api/models/rescan", h.handleModelsRescan)
	mux.HandleFunc("/api/models/browse", h.handleModelsBrowse)
	mux.HandleFunc("/api/models/validate", h.handleModelsValidate)
	mux.HandleFunc("/api/test/record", h.handleTestRecord)
	mux.HandleFunc("/api/permissions", h.handlePermissions)
}

// handleSettings handles GET and PUT /api/settings
func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getSettings(w, r)
	case http.MethodPut:
		h.putSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getSettings returns the current configuration
func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.config)
}

// putSettings updates the configuration
func (h *Handler) putSettings(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.config.Update(updates); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusBadRequest)
		return
	}

	// Save to file
	configPath := config.GetConfigPath()
	if err := h.config.Save(configPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	// 初回設定完了フラグを立てる
	if h.wizard != nil {
		if err := h.wizard.MarkSetupCompleted(); err != nil {
			// エラーログのみ、設定保存は成功しているので処理を継続
			fmt.Printf("Warning: Failed to mark setup completed: %v\n", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// handleHotkeyValidate handles POST /api/hotkey/validate
func (h *Handler) handleHotkeyValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request config.HotkeyConfig
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// HotkeyConfigからModifiersとKeyに変換
	mods := hotkeyConfigToModifiers(request)
	key := stringToKeyCode(request.Key)

	// 競合チェック
	conflicts := hotkey.CheckConflicts(mods, key)

	// レスポンス作成
	conflictNames := []string{}
	for _, c := range conflicts {
		conflictNames = append(conflictNames, c.Name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conflicts": conflictNames,
	})
}

// handleHotkeyRegister handles POST /api/hotkey/register
func (h *Handler) handleHotkeyRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var hotkey config.HotkeyConfig
	if err := json.NewDecoder(r.Body).Decode(&hotkey); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// デバッグログ: 受信したホットキー情報を出力
	fmt.Printf("DEBUG: Received hotkey config: Ctrl=%v, Shift=%v, Alt=%v, Cmd=%v, Key=%q\n",
		hotkey.Ctrl, hotkey.Shift, hotkey.Alt, hotkey.Cmd, hotkey.Key)

	// Validate hotkey configuration
	if hotkey.Key == "" {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if at least one modifier is set (recommended for safety)
	if !hotkey.Ctrl && !hotkey.Shift && !hotkey.Alt && !hotkey.Cmd {
		http.Error(w, "At least one modifier key (Ctrl/Shift/Alt/Cmd) is recommended", http.StatusBadRequest)
		return
	}

	// Update config
	h.config.Hotkey = hotkey

	// Save to file
	configPath := config.GetConfigPath()
	if err := h.config.Save(configPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload hotkey in the running application
	if h.onHotkeyChanged != nil {
		if err := h.onHotkeyChanged(); err != nil {
			// Log warning but don't fail the request (config is already saved)
			fmt.Printf("Warning: Failed to reload hotkey: %v\n", err)
			// Return partial success response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "partial",
				"message": fmt.Sprintf("Hotkey saved but reload failed: %v. Please restart the application.", err),
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Hotkey registered and applied successfully",
	})
}

// Device represents an audio device
type Device struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// convertAudioDevices converts audio.Device slice to api.Device slice
func convertAudioDevices(audioDevices []audio.Device) []Device {
	devices := make([]Device, 0, len(audioDevices))
	for _, dev := range audioDevices {
		devices = append(devices, Device{
			ID:        dev.ID,
			Name:      dev.Name,
			IsDefault: dev.IsDefault,
		})
	}
	return devices
}

// handleDevices handles GET /api/devices
func (h *Handler) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var devices []Device

	// Get actual devices from audio driver
	if h.audioDriver != nil {
		audioDevices, err := h.audioDriver.ListDevices()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list audio devices: %v", err), http.StatusInternalServerError)
			return
		}
		devices = convertAudioDevices(audioDevices)
	} else {
		// AudioDriver not initialized - create a temporary driver to list devices
		// This allows users to see and select devices even before granting microphone permission
		tempDriver, err := audio.NewPortAudioDriver()
		if err != nil {
			// If we can't create a driver, return system default only
			devices = []Device{
				{ID: -1, Name: "システムデフォルト", IsDefault: true},
			}
		} else {
			defer tempDriver.Close()
			audioDevices, err := tempDriver.ListDevices()
			if err != nil {
				// If we can't list devices, return system default only
				devices = []Device{
					{ID: -1, Name: "システムデフォルト", IsDefault: true},
				}
			} else {
				devices = convertAudioDevices(audioDevices)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
	})
}

// Model represents a Whisper model
type Model struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        string `json:"size"`
	Recommended bool   `json:"recommended"`
}

// handleModels handles GET /api/models
func (h *Handler) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := h.scanModels()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
	})
}

// handleModelsRescan handles POST /api/models/rescan
func (h *Handler) handleModelsRescan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := h.scanModels()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
	})
}

// scanModels scans the models directory and returns available models
func (h *Handler) scanModels() []Model {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Cannot get home directory, return empty list
		return []Model{}
	}

	modelsDir := filepath.Join(homeDir, "Library", "Application Support", "EzS2T-Whisper", "models")

	var models []Model

	// Check if directory exists
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		return models
	}

	// Read directory
	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		return models
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only include .bin or .gguf files
		if !config.IsValidModelExtension(entry.Name()) {
			continue
		}

		path := filepath.Join(modelsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		size := formatSize(info.Size())
		// Check if it's the recommended model (compare base name without extension)
		baseName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		recommended := baseName == "ggml-large-v3-turbo-q5_0"

		models = append(models, Model{
			Name:        entry.Name(),
			Path:        path,
			Size:        size,
			Recommended: recommended,
		})
	}

	return models
}

// formatSize formats bytes to human-readable size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// handleTestRecord handles POST /api/test/record
func (h *Handler) handleTestRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement actual test recording
	// For now, return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Test recording not yet implemented",
	})
}

// Permission represents a system permission status
type Permission struct {
	Granted bool `json:"granted"`
}

// handlePermissions handles GET /api/permissions
func (h *Handler) handlePermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get actual permission status
	// For now, return mock data
	permissions := map[string]Permission{
		"microphone":    {Granted: true},
		"accessibility": {Granted: true},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

// handleModelsBrowse handles POST /api/models/browse
// Opens a native file picker dialog using osascript (AppleScript)
func (h *Handler) handleModelsBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use osascript to open macOS file picker
	// AppleScript command to choose file with .bin or .gguf extension
	script := `
		set theFile to choose file with prompt "Whisperモデルファイル (.bin / .gguf) を選択してください" of type {"bin", "gguf"}
		return POSIX path of theFile
	`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		// User cancelled or error occurred
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 128 means user cancelled
			if exitErr.ExitCode() == 128 {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"cancelled": true,
				})
				return
			}
		}
		http.Error(w, fmt.Sprintf("Failed to open file picker: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the file path from output
	filePath := strings.TrimSpace(string(output))

	// Validate the selected file
	expandedPath, err := config.ExpandPath(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid file path: %v", err), http.StatusBadRequest)
		return
	}

	// Check if file exists and is a .bin or .gguf file
	info, err := os.Stat(expandedPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("File not found: %v", err), http.StatusBadRequest)
		return
	}

	if info.IsDir() {
		http.Error(w, "Selected path is a directory, not a file", http.StatusBadRequest)
		return
	}

	if !config.IsValidModelExtension(expandedPath) {
		http.Error(w, "File must have .bin or .gguf extension", http.StatusBadRequest)
		return
	}

	// Return the selected file path
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"path": filePath,
		"name": filepath.Base(filePath),
		"size": formatSize(info.Size()),
	})
}

// handleModelsValidate handles POST /api/models/validate
func (h *Handler) handleModelsValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Expand path
	expandedPath, err := config.ExpandPath(request.Path)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"message": fmt.Sprintf("パスの展開に失敗: %v", err),
		})
		return
	}

	// Check if file exists
	info, err := os.Stat(expandedPath)
	if os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"message": fmt.Sprintf("ファイルが見つかりません: %s", expandedPath),
		})
		return
	}
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"message": fmt.Sprintf("ファイルの確認に失敗: %v", err),
		})
		return
	}

	// Check if it's a regular file
	if info.IsDir() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"message": "指定されたパスはディレクトリです。ファイルを選択してください",
		})
		return
	}

	// Check file extension
	if !config.IsValidModelExtension(expandedPath) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":   false,
			"message": "モデルファイルは .bin または .gguf 拡張子である必要があります",
		})
		return
	}

	// Valid model file
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"message": "モデルファイルは有効です",
		"path":    expandedPath,
		"name":    filepath.Base(expandedPath),
		"size":    formatSize(info.Size()),
	})
}

// hotkeyConfigToModifiers は HotkeyConfig を golang.design/x/hotkey の Modifier スライスに変換
func hotkeyConfigToModifiers(hkConfig config.HotkeyConfig) []hk.Modifier {
	var mods []hk.Modifier
	if hkConfig.Ctrl {
		mods = append(mods, hk.ModCtrl)
	}
	if hkConfig.Shift {
		mods = append(mods, hk.ModShift)
	}
	if hkConfig.Alt {
		mods = append(mods, hk.ModOption)
	}
	if hkConfig.Cmd {
		mods = append(mods, hk.ModCmd)
	}
	return mods
}

// stringToKeyCode は文字列をキーコードに変換
func stringToKeyCode(keyStr string) hk.Key {
	// NBSP正規化: macOS IMEでスペースキーを押すとNBSP（U+00A0）が送信されることがあるため
	if keyStr == "\u00a0" {
		keyStr = "Space"
	}

	keyMap := map[string]hk.Key{
		"Space":  hk.KeySpace,
		"A":      hk.KeyA,
		"B":      hk.KeyB,
		"C":      hk.KeyC,
		"D":      hk.KeyD,
		"E":      hk.KeyE,
		"F":      hk.KeyF,
		"G":      hk.KeyG,
		"H":      hk.KeyH,
		"I":      hk.KeyI,
		"J":      hk.KeyJ,
		"K":      hk.KeyK,
		"L":      hk.KeyL,
		"M":      hk.KeyM,
		"N":      hk.KeyN,
		"O":      hk.KeyO,
		"P":      hk.KeyP,
		"Q":      hk.KeyQ,
		"R":      hk.KeyR,
		"S":      hk.KeyS,
		"T":      hk.KeyT,
		"U":      hk.KeyU,
		"V":      hk.KeyV,
		"W":      hk.KeyW,
		"X":      hk.KeyX,
		"Y":      hk.KeyY,
		"Z":      hk.KeyZ,
		"0":      hk.Key0,
		"1":      hk.Key1,
		"2":      hk.Key2,
		"3":      hk.Key3,
		"4":      hk.Key4,
		"5":      hk.Key5,
		"6":      hk.Key6,
		"7":      hk.Key7,
		"8":      hk.Key8,
		"9":      hk.Key9,
		"Escape": hk.KeyEscape,
		"Return": hk.KeyReturn,
		"Tab":    hk.KeyTab,
	}

	if key, ok := keyMap[keyStr]; ok {
		return key
	}

	// デフォルトはSpace
	return hk.KeySpace
}
