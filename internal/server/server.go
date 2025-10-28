package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

//go:embed frontend/*
var frontendFS embed.FS

// Server manages the HTTP server for settings UI
type Server struct {
	httpServer *http.Server
	listener   net.Listener
	port       int
	mu         sync.Mutex
	running    bool
}

// Config holds server configuration
type Config struct {
	Port            int           // Port to listen on (0 = random)
	ReadTimeout     time.Duration // HTTP read timeout
	WriteTimeout    time.Duration // HTTP write timeout
	ShutdownTimeout time.Duration // Graceful shutdown timeout
}

// DefaultConfig returns the default server configuration
func DefaultConfig() Config {
	return Config{
		Port:            18765, // Default port
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

// New creates a new HTTP server
func New(config Config) *Server {
	return &Server{
		port: config.Port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	// Create listener on localhost only
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Serve frontend static files
	frontendSubFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		listener.Close()
		return fmt.Errorf("failed to create frontend sub-filesystem: %w", err)
	}

	// Add CORS middleware for localhost only
	handler := corsMiddleware(mux)

	// Serve static files
	mux.Handle("/", http.FileServer(http.FS(frontendSubFS)))

	// Create HTTP server
	s.httpServer = &http.Server{
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("HTTP server listening on http://127.0.0.1:%d", s.port)
		if err := s.httpServer.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	s.running = true
	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.running = false
	return nil
}

// Port returns the port the server is listening on
func (s *Server) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}

// URL returns the full URL to the server
func (s *Server) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.Port())
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// corsMiddleware adds CORS headers for localhost-only access
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only allow localhost origins
		origin := r.Header.Get("Origin")

		// Check if origin is localhost or 127.0.0.1
		if origin != "" {
			// Allow localhost and 127.0.0.1 origins
			if len(origin) >= 16 && (origin[:16] == "http://localhost" || origin[:16] == "http://127.0.0.1") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RegisterAPIHandler registers an API handler at the given path
func (s *Server) RegisterAPIHandler(path string, handler http.Handler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.httpServer == nil {
		return fmt.Errorf("server not initialized")
	}

	// Get the mux from the server handler
	mux, ok := s.httpServer.Handler.(*http.ServeMux)
	if !ok {
		// If it's wrapped in middleware, we need to get the underlying mux
		// For now, return an error
		return fmt.Errorf("cannot register handler after server started")
	}

	mux.Handle(path, handler)
	return nil
}
