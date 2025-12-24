package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mcpgo "github.com/DR1N0/mcp-go"
	"github.com/DR1N0/mcp-go/transport/streamable"
)

// EchoArgs defines the arguments for the echo tool
type EchoArgs struct {
	Message string `json:"message" jsonschema:"required,description=The message to echo back"`
}

// AddArgs defines the arguments for the add tool
type AddArgs struct {
	A float64 `json:"a" jsonschema:"required,description=First number"`
	B float64 `json:"b" jsonschema:"required,description=Second number"`
}

func main() {
	// Create server with options
	server := mcpgo.NewServer(
		streamable.NewServerTransport("/mcp", ":8000"),
		mcpgo.WithName("clean-mcp-server"),
		mcpgo.WithVersion("1.0.0"),
	)

	// Register tools - schemas are automatically generated!
	if err := server.RegisterTool("echo", "Echoes back the provided message", echoTool); err != nil {
		log.Fatalf("Failed to register echo tool: %v", err)
	}

	if err := server.RegisterTool("add", "Adds two numbers together", addTool); err != nil {
		log.Fatalf("Failed to register add tool: %v", err)
	}

	// Start the server
	if err := server.Serve(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server is running on http://localhost:8000/mcp")
	log.Println("Press Ctrl+C to stop")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	server.Close()
}

// echoTool handles echo requests
func echoTool(args EchoArgs) (*mcpgo.ToolResponse, error) {
	return mcpgo.NewToolResponse(
		mcpgo.NewTextContent(fmt.Sprintf("Echo: %s", args.Message)),
	), nil
}

// addTool handles addition requests
func addTool(args AddArgs) (*mcpgo.ToolResponse, error) {
	result := args.A + args.B
	return mcpgo.NewToolResponse(
		mcpgo.NewTextContent(fmt.Sprintf("Result: %.2f", result)),
	), nil
}
