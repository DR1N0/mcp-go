package streamable

import (
	"net/http"

	"github.com/DR1N0/mcp-go/transport"
)

// HTTPHandler provides HTTP-specific functionality
type HTTPHandler interface {
	// ServeHTTP handles HTTP requests for the MCP endpoint
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// ServerTransport combines Transport with HTTP handling for streamable HTTP servers
type ServerTransport interface {
	transport.ServerTransport
	HTTPHandler
}

// ClientTransport is a streamable HTTP client transport
type ClientTransport interface {
	transport.ClientTransport
}
