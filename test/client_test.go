package mcpgo_test

import (
	"context"
	"encoding/json"
	"testing"

	mcpgo "github.com/DR1N0/mcp-go"
	"github.com/DR1N0/mcp-go/transport"
	"github.com/DR1N0/mcp-go/types"
)

func TestClient_Creation(t *testing.T) {
	mockTransport := transport.NewMock()
	client := mcpgo.NewClient(mockTransport)

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

func TestMockTransport_SendAndReceive(t *testing.T) {
	mock := transport.NewMock()

	// Set up message handler
	received := false
	mock.SetMessageHandler(func(ctx context.Context, msg *types.BaseJSONRPCMessage) {
		received = true
		if msg.Method != "test" {
			t.Errorf("Expected method 'test', got '%s'", msg.Method)
		}
	})

	// Send a message
	testMsg := &types.BaseJSONRPCMessage{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "test",
	}

	err := mock.Send(context.Background(), testMsg)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Verify message was stored
	msgs := mock.GetSentMessages()
	if len(msgs) != 1 {
		t.Errorf("Expected 1 sent message, got %d", len(msgs))
	}

	// Simulate receiving a message
	mock.SimulateReceive(context.Background(), testMsg)

	if !received {
		t.Error("Message handler was not called")
	}
}

func TestMockTransport_StartAndClose(t *testing.T) {
	mock := transport.NewMock()

	// Test start
	err := mock.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !mock.IsStarted() {
		t.Error("Transport should be started")
	}

	// Test close
	err = mock.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if !mock.IsClosed() {
		t.Error("Transport should be closed")
	}
}

func TestMockTransport_ErrorHandling(t *testing.T) {
	mock := transport.NewMock()

	// Test close handler
	closeReceived := false
	mock.SetCloseHandler(func() {
		closeReceived = true
	})

	// Close should trigger close handler
	mock.Close()

	if !closeReceived {
		t.Error("Close handler was not called")
	}
}

func TestMockTransport_ClearMessages(t *testing.T) {
	mock := transport.NewMock()

	// Send some messages
	msg := &types.BaseJSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "test",
	}

	mock.Send(context.Background(), msg)
	mock.Send(context.Background(), msg)

	if len(mock.GetSentMessages()) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(mock.GetSentMessages()))
	}

	// Clear messages
	mock.ClearSentMessages()

	if len(mock.GetSentMessages()) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(mock.GetSentMessages()))
	}
}
