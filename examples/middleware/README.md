# HTTP Middleware Example

This example demonstrates how to use HTTP middleware with mcp-go to add cross-cutting concerns like authentication, logging, and CORS to your MCP servers.

## Overview

The middleware feature allows you to chain standard Go HTTP middleware functions with your MCP server transport. Middleware is applied in **reverse order**, so the last middleware added becomes the outermost wrapper.

## What's Included

- **Server** (`server/main.go`) - MCP server with authentication, logging, and CORS middleware
- **Client** (`client/main.go`) - Test client that demonstrates both successful and failed authentication

## Middleware Chain

```
Request Flow:
1. CORS middleware (outermost - applied first)
2. Logging middleware  
3. Authentication middleware (innermost - applied last)
4. MCP Handler
```

## Running the Example

### 1. Start the Server

```bash
cd examples/middleware/server
go run main.go
```

The server will start on `http://localhost:8080/mcp` with middleware enabled.

### 2. Run the Test Client

In a separate terminal:

```bash
cd examples/middleware/client  
go run main.go
```

The client will run two tests:
- **Test 1**: Call server without auth (should fail with 401)
- **Test 2**: Call server with valid Bearer token (should succeed)

### 3. Manual Testing with curl

**Without authentication (will fail):**
```bash
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
```

**With authentication (will succeed):**
```bash
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer demo-token' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
```

## Creating Custom Middleware

Middleware functions follow the standard Go HTTP middleware pattern:

```go
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing
        log.Println("Before request")
        
        // Call next handler
        next.ServeHTTP(w, r)
        
        // Post-processing  
        log.Println("After request")
    })
}
```

## Common Middleware Patterns

### Authentication

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !validateToken(token) {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("Invalid credentials"))
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Rate Limiting

```go
import "golang.org/x/time/rate"

func rateLimitMiddleware(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(10, 100) // 10 req/sec, burst 100
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Request ID Tracking

```go
import "github.com/google/uuid"

func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        w.Header().Set("X-Request-ID", requestID)
        
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### OpenTelemetry Tracing

```go
import "go.opentelemetry.io/otel"

func telemetryMiddleware(next http.Handler) http.Handler {
    tracer := otel.Tracer("mcp-server")
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, span := tracer.Start(r.Context(), "mcp.request")
        defer span.End()
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Prometheus Metrics

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "mcp_requests_total"},
        []string{"method", "status"},
    )
)

func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rw := &responseWriter{ResponseWriter: w}
        next.ServeHTTP(rw, r)
        
        requestCounter.WithLabelValues(
            r.Method,
            fmt.Sprintf("%d", rw.status),
        ).Inc()
    })
}
```

## Usage in Your Code

```go
import (
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/streamable"
    "github.com/DR1N0/mcp-go/types"
)

// Create your middleware functions
func authMiddleware(next http.Handler) http.Handler { /* ... */ }
func loggingMiddleware(next http.Handler) http.Handler { /* ... */ }
func corsMiddleware(next http.Handler) http.Handler { /* ... */ }

// Create transport with middleware
transport := streamable.NewServerTransport("/mcp", ":8080").
    WithMiddleware(authMiddleware).
    WithMiddleware(loggingMiddleware).
    WithMiddleware(corsMiddleware)

// Create server
server := mcpgo.NewServer(transport)
server.Serve()
```

## Middleware Type

The middleware type is defined in the `types` package:

```go
import "github.com/DR1N0/mcp-go/types"

type HTTPMiddleware = types.HTTPMiddleware
// Equivalent to: func(http.Handler) http.Handler
```

## Supported Transports

Middleware is available on:
- ✅ Streamable HTTP transport (`transport/streamable`)
- ✅ SSE transport (`transport/sse`)
- ❌ Stdio transport (not applicable - no HTTP layer)

## Production Considerations

1. **Order Matters**: Middleware executes in reverse order of declaration
2. **Error Handling**: Always handle errors gracefully in middleware
3. **Performance**: Avoid blocking operations in middleware
4. **Security**: Validate all inputs, especially auth tokens
5. **Logging**: Use structured logging for better observability
6. **Metrics**: Track request counts, latencies, and error rates

## Learn More

- [Chi Router Middleware](https://github.com/go-chi/chi#middleware) - Inspiration for this pattern
- [Go HTTP Middleware Patterns](https://www.alexedwards.net/blog/making-and-using-middleware)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
