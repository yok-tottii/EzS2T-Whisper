package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/YOURUSERNAME/EzS2T-Whisper/internal/config"
)

// Handler manages API endpoints
type Handler struct {
	config *config.Config
}

// New creates a new API handler
func New(cfg *config.Config) *Handler {
	return &Handler{
		config: cfg,
	}
}

// RegisterRoutes registers all API routes on the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/settings", h.handleSettings)
	mux.HandleFunc("/api/hotkey/validate", h.handleHotkeyValidate)
	mux.HandleFunc("/api/hotkey/register", h.handleHotkeyRegister)
	mux.HandleFunc("/api/devices", h.handleDevices)
	mux.HandleFunc("/api/models", h.handleModels)
	mux.HandleFunc("/api/models/rescan", h.handleModelsRescan)
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

	// TODO: Implement actual conflict detection
	// For now, return no conflicts
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conflicts": []string{},
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

	// Update config
	h.config.Hotkey = hotkey

	// Save to file
	configPath := config.GetConfigPath()
	if err := h.config.Save(configPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// Device represents an audio device
type Device struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// handleDevices handles GET /api/devices
func (h *Handler) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get actual devices from audio driver
	// For now, return mock data
	devices := []Device{
		{ID: 0, Name: "Built-in Microphone", IsDefault: true},
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
	homeDir, _ := os.UserHomeDir()
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

		// Only include .gguf files
		if filepath.Ext(entry.Name()) != ".gguf" {
			continue
		}

		path := filepath.Join(modelsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		size := formatSize(info.Size())
		recommended := entry.Name() == "ggml-large-v3-turbo-q5_0.gguf"

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
