# Server API Design Refactoring

## Overview

The HTTP server and API handler design has been refactored to support flexible route registration patterns while maintaining thread safety and proper separation of concerns.

## Problem Statement

The original implementation had a design issue where `RegisterAPIHandler` attempted to inspect `s.httpServer.Handler` after the server started. This caused two problems:

1. **Dynamic Registration After Start Failed**: The method assumed the handler was a bare `*http.ServeMux`, but in production it's wrapped with middleware (CORS), making type assertion fail
2. **No Separation of Concerns**: Routes registration and server initialization were tightly coupled

## Solution: Dedicated ServeMux

The refactored design introduces a dedicated `mux` field on the `Server` struct, initialized upfront and used consistently:

### Key Changes

#### 1. Server Struct Enhancement
```go
type Server struct {
    httpServer *http.Server
    listener   net.Listener
    port       int
    mux        *http.ServeMux  // NEW: Dedicated multiplexer
    mu         sync.Mutex
    running    bool
}
```

#### 2. Initialization in `New()`
```go
func New(config Config) *Server {
    return &Server{
        port: config.Port,
        mux:  http.NewServeMux(),  // Initialize mux upfront
    }
}
```

#### 3. Simplified `RegisterAPIHandler()`
```go
func (s *Server) RegisterAPIHandler(path string, handler http.Handler) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.mux == nil {
        return fmt.Errorf("server mux not initialized")
    }

    s.mux.Handle(path, handler)
    return nil
}
```

#### 4. Updated `Start()` to Use Pre-Existing Mux
```go
func (s *Server) Start() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.running {
        return fmt.Errorf("server already running")
    }

    // ... listener setup ...

    // Register static files on the PRE-EXISTING mux
    s.mux.Handle("/", http.FileServer(http.FS(frontendSubFS)))

    // Wrap mux with middleware
    handler := corsMiddleware(s.mux)

    // Create HTTP server with wrapped handler
    s.httpServer = &http.Server{
        Handler:      handler,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    // ... start server ...
}
```

#### 5. New `GetMux()` Method
```go
func (s *Server) GetMux() *http.ServeMux {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.mux
}
```

## Usage Patterns

### Pattern 1: Register Routes Before Starting (Recommended)

```go
// Create server
server := server.New(server.DefaultConfig())

// Create API handler
apiHandler := api.New(config.DefaultConfig())

// Register API routes on the server's mux
apiHandler.RegisterRoutes(server.GetMux())

// Start the server
if err := server.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}
defer server.Stop()
```

**Advantages:**
- Simplest and most straightforward
- All routes registered upfront
- Clear initialization sequence
- Perfect for standard application startup

### Pattern 2: Register Routes After Starting

```go
// Create and start server
server := server.New(server.DefaultConfig())
if err := server.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}

// Register routes dynamically after server is running
testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Dynamic handler"))
})

if err := server.RegisterAPIHandler("/dynamic/route", testHandler); err != nil {
    log.Printf("Failed to register handler: %v", err)
}
```

**Advantages:**
- Supports lazy loading of routes
- Can register routes based on runtime conditions
- Thread-safe during server operation

### Pattern 3: Direct Mux Access

```go
server := server.New(server.DefaultConfig())

// Get direct access to the mux with locking
mux := server.GetMux()
mux.HandleFunc("/api/custom", func(w http.ResponseWriter, r *http.Request) {
    // Handle request
})

// Start server
if err := server.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}
```

**Advantages:**
- Lowest overhead for simple handlers
- Full flexibility with `http.ServeMux`
- Mux access is properly locked

## Thread Safety

All operations are protected by `sync.Mutex`:

1. **`RegisterAPIHandler()`**: Acquires lock before modifying mux
2. **`GetMux()`**: Acquires lock before returning mux reference
3. **`Start()`**: Acquires lock during initialization and state changes
4. **`Stop()`**: Acquires lock during shutdown

## API Registration Pattern

The `api.Handler` type provides a convenient registration method:

```go
// api/api.go
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/api/settings", h.handleSettings)
    mux.HandleFunc("/api/hotkey/validate", h.handleHotkeyValidate)
    // ... more routes ...
}
```

This pattern allows:
- Decoupling API handler from server implementation
- Reusable route registration logic
- Easy testing with test mux instances

## Testing

The refactoring includes comprehensive integration tests in `server_api_integration_test.go`:

1. **TestServerAPIIntegration**: Full server + API integration
2. **TestRegisterAPIHandlerBeforeStart**: Routes registered before startup
3. **TestRegisterAPIHandlerAfterStart**: Dynamic route registration
4. **TestGetMux**: Direct mux access with locking
5. **TestConcurrentHandlerRegistration**: Thread-safety verification

All tests pass successfully, demonstrating:
- Correct initialization and handler registration
- Proper thread safety with concurrent registration
- Works with both pre-start and post-start registration
- CORS middleware properly applied

## Middleware Stack

The final handler chain is:

```
corsMiddleware
    ↓
http.ServeMux (s.mux)
    ├── /api/* endpoints (registered by api.Handler.RegisterRoutes)
    ├── /static/* endpoints (could be added)
    └── / static file serving (registered in Start())
```

The middleware wrapping occurs in `Start()`:
```go
handler := corsMiddleware(s.mux)
s.httpServer = &http.Server{
    Handler: handler,
    // ...
}
```

## Migration Guide

If you have code using the old design:

### Old Code
```go
server := server.New(config)
server.Start()
server.RegisterAPIHandler("/api/test", handler) // This used to fail!
```

### New Code
```go
server := server.New(config)

// Option 1: Use api.Handler for standard routes
apiHandler := api.New(appConfig)
apiHandler.RegisterRoutes(server.GetMux())

// Option 2: Or register routes manually
server.RegisterAPIHandler("/api/test", handler)

server.Start()
```

## Benefits of This Design

1. **Flexibility**: Routes can be registered before or after server starts
2. **Thread Safety**: All mux access is properly protected by mutex
3. **Testability**: Easy to create test servers with specific routes
4. **Cleanliness**: Separate concerns between server and API handler
5. **Maintainability**: Clear initialization sequence and usage patterns
6. **Extensibility**: Can add custom middleware or route groups easily
7. **Middleware Support**: Works correctly with wrapped handlers

## Future Enhancements

Possible improvements building on this foundation:

1. **Route Groups**: Group related routes with common prefix/middleware
2. **Before/After Hooks**: Register lifecycle hooks for routes
3. **Route Validation**: Validate routes don't conflict during registration
4. **Dynamic Reloading**: Hot-reload routes without server restart
5. **Middleware Chain**: More sophisticated middleware composition

## See Also

- `internal/server/server.go`: Server implementation
- `internal/api/api.go`: API handler implementation
- `internal/server/server_api_integration_test.go`: Integration tests
