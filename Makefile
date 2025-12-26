.PHONY: vendor
vendor: ## Update vendor dependencies
	@echo "Updating module dependencies..."
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor

.PHONY: test
test: ## Run all tests
	@echo "Running all go tests..."
	go test -v ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin
	@echo "✅ Cleaned"

.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

# Examples

# Streamable HTTP
.PHONY: server-streamable
server-streamable: ## Run the MCP server with streamable HTTP transport
	@echo "Starting Streamable HTTP MCP server..."
	go run ./examples/streamable_http/server/main.go

.PHONY: client-streamable
client-streamable: ## Run the MCP client with streamable HTTP transport
	@echo "Starting Streamable HTTP MCP client..."
	go run ./examples/streamable_http/clients/go/main.go

.PHONY: client-streamable-python
client-streamable-python: ## Run the MCP client with streamable HTTP transport (Python)
	@echo "Starting Streamable HTTP MCP client (Python)..."
	uv run ./examples/streamable_http/clients/python/main.py

# Stdio
.PHONY: build-stdio-server
build-stdio-server: ## Build the MCP stdio server binary
	@echo "Building Stdio MCP server binary..."
	go build -o ./bin/stdio_server ./examples/stdio/server/main.go
	
.PHONY: server-stdio
server-stdio: ## Run the MCP server with stdio transport, for manual testing mainly
	@echo "Starting Stdio MCP server..."
	go run ./examples/stdio/server/main.go

.PHONY: client-stdio
client-stdio: ## Run the MCP client with stdio transport
	@echo "Starting Stdio MCP client..."
	go run ./examples/stdio/clients/go/main.go

.PHONY: client-stdio-python
client-stdio-python: ## Run the MCP client with stdio transport (Python)
	@echo "Starting Stdio MCP client (Python)..."
	uv run ./examples/stdio/clients/python/main.py

# SSE
.PHONY: server-sse
server-sse: ## Run the MCP server with SSE transport
	@echo "Starting SSE MCP server..."
	go run ./examples/sse/server/main.go

.PHONY: client-sse
client-sse: ## Run the MCP client with SSE transport
	@echo "Starting SSE MCP client..."
	go run ./examples/sse/clients/go/main.go

.PHONY: client-sse-python
client-sse-python: ## Run the MCP client with SSE transport (Python)
	@echo "Starting SSE MCP client (Python)..."
	uv run ./examples/sse/clients/python/main.py

# Middleware
.PHONY: server-middleware
server-middleware: ## Run the middleware example server with authentication
	@echo "Starting Middleware Example server..."
	go run ./examples/middleware/server/main.go

.PHONY: client-middleware
client-middleware: ## Run the middleware example client
	@echo "Starting Middleware Example client..."
	go run ./examples/middleware/client/main.go
