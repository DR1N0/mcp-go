
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