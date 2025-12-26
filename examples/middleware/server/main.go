package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	mcpgo "github.com/DR1N0/mcp-go"
	"github.com/DR1N0/mcp-go/transport/streamable"
)

// authMiddleware demonstrates authentication middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Authorization header required"))
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid authorization format"))
			return
		}

		log.Printf("✓ Authenticated request from: %s", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("→ [%s] %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type EchoArgs struct {
	Message string `json:"message" jsonschema:"required,description=Message to echo"`
}

func echoTool(args EchoArgs) (*mcpgo.ToolResponse, error) {
	return mcpgo.NewToolResponse(
		mcpgo.NewTextContent(fmt.Sprintf("Echo: %s", args.Message)),
	), nil
}

func main() {
	fmt.Println("=================================================================================")
	fmt.Println("MCP Server with HTTP Middleware")
	fmt.Println("=================================================================================")
	fmt.Println()

	// Create transport with middleware chain
	transport := streamable.NewServerTransport("/mcp", ":8080").
		WithMiddleware(authMiddleware).
		WithMiddleware(loggingMiddleware).
		WithMiddleware(corsMiddleware)

	server := mcpgo.NewServer(
		transport,
		mcpgo.WithName("middleware-example"),
		mcpgo.WithVersion("1.0.0"),
	)

	if err := server.RegisterTool("echo", "Echoes a message", echoTool); err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}

	if err := server.Serve(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Println("✅ Server started with middleware:")
	fmt.Println("   • CORS enabled")
	fmt.Println("   • Request logging")
	fmt.Println("   • Bearer token auth")
	fmt.Println()
	fmt.Println("Running on: http://localhost:8080/mcp")
	fmt.Println()
	fmt.Println("Test with:")
	fmt.Println("  curl -X POST http://localhost:8080/mcp \\")
	fmt.Println("    -H 'Authorization: Bearer demo-token' \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{}}'")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	server.Close()
}
