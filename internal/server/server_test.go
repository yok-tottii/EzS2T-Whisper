package server

import (
	"io"
	"net/http"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Port != 18765 {
		t.Errorf("Expected port 18765, got %d", config.Port)
	}

	if config.ReadTimeout != 10*time.Second {
		t.Errorf("Expected ReadTimeout 10s, got %v", config.ReadTimeout)
	}

	if config.WriteTimeout != 10*time.Second {
		t.Errorf("Expected WriteTimeout 10s, got %v", config.WriteTimeout)
	}

	if config.ShutdownTimeout != 5*time.Second {
		t.Errorf("Expected ShutdownTimeout 5s, got %v", config.ShutdownTimeout)
	}
}

func TestNew(t *testing.T) {
	config := DefaultConfig()
	server := New(config)

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.port != config.Port {
		t.Errorf("Expected port %d, got %d", config.Port, server.port)
	}

	if server.running {
		t.Error("Expected server to not be running initially")
	}
}

func TestStartStop(t *testing.T) {
	config := DefaultConfig()
	config.Port = 0 // Use random port
	server := New(config)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Check that server is running
	if !server.IsRunning() {
		t.Error("Expected server to be running")
	}

	// Check that port was assigned
	port := server.Port()
	if port == 0 {
		t.Error("Expected non-zero port")
	}

	// Try to start again (should fail)
	if err := server.Start(); err == nil {
		t.Error("Expected error when starting already running server")
	}

	// Give server a moment to fully start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	if err := server.Stop(); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Check that server is stopped
	if server.IsRunning() {
		t.Error("Expected server to be stopped")
	}

	// Stop again (should succeed, no-op)
	if err := server.Stop(); err != nil {
		t.Errorf("Expected no error when stopping already stopped server: %v", err)
	}
}

func TestURL(t *testing.T) {
	config := DefaultConfig()
	config.Port = 12345
	server := New(config)

	expectedURL := "http://127.0.0.1:12345"
	if server.URL() != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, server.URL())
	}
}

func TestServerServesFrontend(t *testing.T) {
	config := DefaultConfig()
	config.Port = 0 // Use random port
	server := New(config)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Give server a moment to fully start
	time.Sleep(100 * time.Millisecond)

	// Make HTTP request to root
	url := server.URL() + "/"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check that we got HTML content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}

	// Check for HTML content
	bodyStr := string(body)
	if len(bodyStr) < 100 {
		t.Errorf("Expected substantial HTML content, got %d bytes", len(bodyStr))
	}
}

func TestCORSMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with CORS middleware
	handler := corsMiddleware(testHandler)

	// Test OPTIONS request
	req, err := http.NewRequest("OPTIONS", "http://127.0.0.1:8080/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Origin", "http://127.0.0.1:8080")

	recorder := &testResponseWriter{
		headers: make(http.Header),
	}

	handler.ServeHTTP(recorder, req)

	// Check CORS headers
	if recorder.headers.Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header to be set")
	}
}

// testResponseWriter is a simple ResponseWriter for testing
type testResponseWriter struct {
	headers    http.Header
	statusCode int
	body       []byte
}

func (w *testResponseWriter) Header() http.Header {
	return w.headers
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestMultipleStartStop(t *testing.T) {
	config := DefaultConfig()
	config.Port = 0

	for i := 0; i < 3; i++ {
		server := New(config)

		if err := server.Start(); err != nil {
			t.Fatalf("Iteration %d: Failed to start server: %v", i, err)
		}

		time.Sleep(50 * time.Millisecond)

		if err := server.Stop(); err != nil {
			t.Fatalf("Iteration %d: Failed to stop server: %v", i, err)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func TestPort(t *testing.T) {
	config := DefaultConfig()
	config.Port = 19999
	server := New(config)

	// Before start, should return configured port
	if server.Port() != 19999 {
		t.Errorf("Expected port 19999 before start, got %d", server.Port())
	}

	// Start with port 0 to get random port
	config.Port = 0
	server = New(config)

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// After start with port 0, should return assigned port
	port := server.Port()
	if port == 0 {
		t.Error("Expected non-zero port after start")
	}
}
