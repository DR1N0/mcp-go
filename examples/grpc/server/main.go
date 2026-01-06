package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	mcpgo "github.com/DR1N0/mcp-go"
	"github.com/DR1N0/mcp-go/transport/grpc"
)

// Tool argument structs
type EchoArgs struct {
	Message string `json:"message" jsonschema:"required,description=The message to echo back"`
}

type AddArgs struct {
	A int `json:"a" jsonschema:"required,description=First number"`
	B int `json:"b" jsonschema:"required,description=Second number"`
}

// Prompt argument struct
type GreetingArgs struct {
	Name string `json:"name" jsonschema:"required,description=Name to greet"`
}

// Weather argument struct
type WeatherArgs struct {
	Latitude  float64 `json:"latitude" jsonschema:"required,description=The latitude of the location"`
	Longitude float64 `json:"longitude" jsonschema:"required,description=The longitude of the location"`
}

func main() {
	// Create server with gRPC transport
	server := mcpgo.NewServer(
		grpc.NewServerTransport(),
		mcpgo.WithName("grpc-example-server"),
		mcpgo.WithVersion("1.0.0"),
	)

	// Register Tools
	if err := server.RegisterTool("echo", "Echoes back the provided message", echoTool); err != nil {
		log.Fatalf("Failed to register echo tool: %v", err)
	}

	if err := server.RegisterTool("add", "Adds two numbers together", addTool); err != nil {
		log.Fatalf("Failed to register add tool: %v", err)
	}

	if err := server.RegisterTool("get_weather", "Get the weather forecast for temperature, wind speed and relative humidity", weatherTool); err != nil {
		log.Fatalf("Failed to register weather tool: %v", err)
	}

	// Register Prompts
	if err := server.RegisterPrompt("greeting", "Generates a greeting prompt", greetingPrompt); err != nil {
		log.Fatalf("Failed to register greeting prompt: %v", err)
	}

	// Register Resources
	if err := server.RegisterResource(
		"lyrics://never-gonna-give-you-up",
		"Never Gonna Give You Up Lyrics",
		"The lyrics of Never Gonna Give You Up by Rick Astley",
		"text/plain",
		lyricResource,
	); err != nil {
		log.Fatalf("Failed to register lyrics resource: %v", err)
	}

	if err := server.RegisterResource(
		"config://server",
		"Server Configuration",
		"Configuration details of the MCP server",
		"application/json",
		configResource,
	); err != nil {
		log.Fatalf("Failed to register config resource: %v", err)
	}

	// Start server
	if err := server.Serve(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("✅ MCP gRPC Server started successfully!")
	log.Println("   - gRPC endpoint: localhost:50051")
	log.Println("   - Protocol: bidirectional streaming")
	log.Println("   - Tools: echo, add, get_weather")
	log.Println("   - Prompts: greeting")
	log.Println("   - Resources: config://server, lyrics://never-gonna-give-you-up")
	log.Println("\nPress Ctrl+C to stop...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down server...")
	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}
}

// Tool handlers

func echoTool(args EchoArgs) (*mcpgo.ToolResponse, error) {
	return mcpgo.NewToolResponse(
		mcpgo.NewTextContent(fmt.Sprintf("Echo: %s", args.Message)),
	), nil
}

func addTool(args AddArgs) (*mcpgo.ToolResponse, error) {
	result := args.A + args.B
	return mcpgo.NewToolResponse(
		mcpgo.NewTextContent(fmt.Sprintf("%d + %d = %d", args.A, args.B, result)),
	), nil
}

func weatherTool(args WeatherArgs) (*mcpgo.ToolResponse, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,wind_speed_10m&hourly=temperature_2m,relative_humidity_2m,wind_speed_10m",
		args.Latitude, args.Longitude)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	output, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return mcpgo.NewToolResponse(
		mcpgo.NewTextContent(string(output)),
	), nil
}

// Prompt handlers

func greetingPrompt(args GreetingArgs) (*mcpgo.PromptResponse, error) {
	return mcpgo.NewPromptResponse(
		"A friendly greeting prompt",
		mcpgo.NewPromptMessage(
			mcpgo.NewTextContent(fmt.Sprintf("Hello %s! How can I assist you today?", args.Name)),
			mcpgo.RoleAssistant,
		),
	), nil
}

// Resource handlers

func lyricResource() (*mcpgo.ResourceResponse, error) {
	lyrics := `Never gonna give you up
Never gonna let you down
Never gonna run around and desert you
Never gonna make you cry
Never gonna say goodbye
Never gonna tell a lie and hurt you`
	return mcpgo.NewResourceResponse(
		mcpgo.NewTextResource("lyrics://never-gonna-give-you-up", lyrics, "text/plain"),
	), nil
}

func configResource() (*mcpgo.ResourceResponse, error) {
	configJSON := `{
	"server": "grpc-example-server",
	"version": "1.0.0",
	"features": ["tools", "prompts", "resources"],
	"transport": "grpc",
	"port": 50051
}`
	return mcpgo.NewResourceResponse(
		mcpgo.NewTextResource("config://server", configJSON, "application/json"),
	), nil
}
