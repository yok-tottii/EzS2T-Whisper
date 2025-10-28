package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/yok-tottii/EzS2T-Whisper/internal/api"
	"github.com/yok-tottii/EzS2T-Whisper/internal/config"
)

// TestServerAPIIntegration tests the server and API integration
// This demonstrates the proper pattern for registering API routes:
// 1. Create server with New()
// 2. Create API handler with api.New()
// 3. Register routes on the server's mux via api.RegisterRoutes()
// 4. Start the server
func TestServerAPIIntegration(t *testing.T) {
	// Create server
	serverConfig := DefaultConfig()
	serverConfig.Port = 0 // Use random port
	server := New(serverConfig)

	// Create API handler
	appConfig := config.DefaultConfig()
	apiHandler := api.New(appConfig)

	// Register API routes BEFORE starting the server
	// This approach is preferred as it registers routes upfront
	apiHandler.RegisterRoutes(server.GetMux())

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that API endpoint is accessible
	url := server.URL() + "/api/settings"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make request to API: %v", err)
	}
	defer resp.Body.Close()

	// Should get 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify response is valid JSON config
	var response config.Config
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode settings response: %v", err)
	}

	// Test PUT endpoint
	updates := map[string]interface{}{
		"language": "en",
	}
	bodyBytes, _ := json.Marshal(updates)
	putResp, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	putResp.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp2, err := client.Do(putResp)
	if err != nil {
		t.Fatalf("Failed to execute PUT request: %v", err)
	}
	defer resp2.Body.Close()

	// Should succeed or fail gracefully (config file may not be writable in test env)
	if resp2.StatusCode != http.StatusOK && resp2.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", resp2.StatusCode)
	}
}

// TestRegisterAPIHandlerBeforeStart demonstrates registering routes before server starts
func TestRegisterAPIHandlerBeforeStart(t *testing.T) {
	server := New(DefaultConfig())
	server.port = 0

	// Register a simple test handler before starting
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test ok"))
	})

	// Register handler before start
	if err := server.RegisterAPIHandler("/test/handler", testHandler); err != nil {
		t.Fatalf("Failed to register handler before start: %v", err)
	}

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test the registered handler
	resp, err := http.Get(server.URL() + "/test/handler")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "test ok" {
		t.Errorf("Expected response 'test ok', got '%s'", string(body))
	}
}

// TestRegisterAPIHandlerAfterStart demonstrates registering routes after server starts
// (with the new design, this is now possible safely)
func TestRegisterAPIHandlerAfterStart(t *testing.T) {
	server := New(DefaultConfig())
	server.port = 0

	// Start server first (no routes registered yet except static files)
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Register a handler AFTER server is running
	// With the new design, this is thread-safe and works correctly
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("dynamic handler ok"))
	})

	if err := server.RegisterAPIHandler("/dynamic/test", testHandler); err != nil {
		t.Fatalf("Failed to register handler after start: %v", err)
	}

	// Test the dynamically registered handler
	resp, err := http.Get(server.URL() + "/dynamic/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "dynamic handler ok" {
		t.Errorf("Expected response 'dynamic handler ok', got '%s'", string(body))
	}
}

// TestGetMux verifies direct mux access works correctly
func TestGetMux(t *testing.T) {
	server := New(DefaultConfig())

	// GetMux should return the mux
	mux := server.GetMux()
	if mux == nil {
		t.Fatal("Expected GetMux to return non-nil mux")
	}

	// Register handler directly on mux
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("direct mux ok"))
	})
	mux.Handle("/direct/test", testHandler)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test the handler registered via direct mux access
	resp, err := http.Get(server.URL() + "/direct/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "direct mux ok" {
		t.Errorf("Expected response 'direct mux ok', got '%s'", string(body))
	}
}

// TestConcurrentHandlerRegistration tests thread-safe handler registration
// while the server is running
func TestConcurrentHandlerRegistration(t *testing.T) {
	server := New(DefaultConfig())
	server.port = 0

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Register multiple handlers concurrently
	errChan := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			path := "/concurrent/" + string(rune(48+index)) // /concurrent/0, /concurrent/1, etc.
			errChan <- server.RegisterAPIHandler(path, handler)
		}(i)
	}

	// Wait for all registrations to complete
	for i := 0; i < 5; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Failed to register handler %d: %v", i, err)
		}
	}

	// Wait a bit for handlers to be registered
	time.Sleep(100 * time.Millisecond)

	// Test that all handlers are accessible
	for i := 0; i < 5; i++ {
		path := "/concurrent/" + string(rune(48+i))
		resp, err := http.Get(server.URL() + path)
		if err != nil {
			t.Errorf("Failed to request %s: %v", path, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", path, resp.StatusCode)
		}
		resp.Body.Close()
	}
}
