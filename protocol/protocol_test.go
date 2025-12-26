package protocol

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DR1N0/mcp-go/transport"
	"github.com/DR1N0/mcp-go/types"
)

func TestProtocol_RequestResponse(t *testing.T) {
	mock := transport.NewMock()
	proto := NewProtocol()

	if err := proto.Connect(mock); err != nil {
		t.Fatalf("Failed to connect protocol: %v", err)
	}
	defer proto.Close()

	// Simulate response
	go func() {
		time.Sleep(50 * time.Millisecond)
		msgs := mock.GetSentMessages()
		if len(msgs) > 0 {
			response := &types.BaseJSONRPCMessage{
				JSONRPC: "2.0",
				ID:      msgs[0].ID,
				Result:  json.RawMessage(`{"status": "ok"}`),
			}
			mock.SimulateReceive(context.Background(), response)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := proto.Request(ctx, "test/method", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestProtocol_ErrorResponse(t *testing.T) {
	mock := transport.NewMock()
	proto := NewProtocol()

	if err := proto.Connect(mock); err != nil {
		t.Fatalf("Failed to connect protocol: %v", err)
	}
	defer proto.Close()

	// Simulate error response
	go func() {
		time.Sleep(50 * time.Millisecond)
		msgs := mock.GetSentMessages()
		if len(msgs) > 0 {
			response := &types.BaseJSONRPCMessage{
				JSONRPC: "2.0",
				ID:      msgs[0].ID,
				Error: &types.RPCError{
					Code:    -32600,
					Message: "Invalid Request",
				},
			}
			mock.SimulateReceive(context.Background(), response)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := proto.Request(ctx, "test/method", nil)
	if err == nil {
		t.Error("Expected error response")
	}

	if err != nil && err.Error() != "RPC error -32600: Invalid Request" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestProtocol_Timeout(t *testing.T) {
	mock := transport.NewMock()
	proto := NewProtocol()

	if err := proto.Connect(mock); err != nil {
		t.Fatalf("Failed to connect protocol: %v", err)
	}
	defer proto.Close()

	// Don't simulate any response - let it timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := proto.Request(ctx, "test/method", nil)
	if err == nil {
		t.Error("Expected timeout error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got: %v", err)
	}
}

func TestProtocol_Notification(t *testing.T) {
	mock := transport.NewMock()
	proto := NewProtocol()

	if err := proto.Connect(mock); err != nil {
		t.Fatalf("Failed to connect protocol: %v", err)
	}
	defer proto.Close()

	err := proto.Notification("test/notify", map[string]string{"data": "test"})
	if err != nil {
		t.Fatalf("Notification failed: %v", err)
	}

	// Verify message was sent
	msgs := mock.GetSentMessages()
	if len(msgs) != 1 {
		t.Errorf("Expected 1 message, got %d", len(msgs))
	}

	if msgs[0].Method != "test/notify" {
		t.Errorf("Expected method 'test/notify', got '%s'", msgs[0].Method)
	}

	if msgs[0].ID != nil {
		t.Error("Notification should not have an ID")
	}
}

func TestProtocol_IncomingRequest(t *testing.T) {
	mock := transport.NewMock()
	proto := NewProtocol()

	if err := proto.Connect(mock); err != nil {
		t.Fatalf("Failed to connect protocol: %v", err)
	}
	defer proto.Close()

	// Register request handler
	handlerCalled := false
	proto.SetRequestHandler("test/method", func(ctx context.Context, params interface{}) (interface{}, error) {
		handlerCalled = true
		return map[string]string{"result": "success"}, nil
	})

	// Simulate incoming request
	request := &types.BaseJSONRPCMessage{
		JSONRPC: "2.0",
		ID:      int64(1),
		Method:  "test/method",
		Params:  json.RawMessage(`{}`),
	}

	mock.SimulateReceive(context.Background(), request)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	if !handlerCalled {
		t.Error("Handler was not called")
	}

	// Check that response was sent
	msgs := mock.GetSentMessages()
	foundResponse := false
	for _, msg := range msgs {
		if msg.ID != nil && msg.Result != nil {
			foundResponse = true
			break
		}
	}

	if !foundResponse {
		t.Error("No response was sent")
	}
}

func TestProtocol_IncomingNotification(t *testing.T) {
	mock := transport.NewMock()
	proto := NewProtocol()

	if err := proto.Connect(mock); err != nil {
		t.Fatalf("Failed to connect protocol: %v", err)
	}
	defer proto.Close()

	// Register notification handler
	handlerCalled := false
	proto.SetNotificationHandler("test/notify", func(params interface{}) error {
		handlerCalled = true
		return nil
	})

	// Simulate incoming notification (no ID)
	notification := &types.BaseJSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "test/notify",
		Params:  json.RawMessage(`{}`),
	}

	mock.SimulateReceive(context.Background(), notification)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	if !handlerCalled {
		t.Error("Notification handler was not called")
	}
}

func TestProtocol_ConcurrentRequests(t *testing.T) {
	mock := transport.NewMock()
	proto := NewProtocol()

	if err := proto.Connect(mock); err != nil {
		t.Fatalf("Failed to connect protocol: %v", err)
	}
	defer proto.Close()

	// Respond to all requests
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			msgs := mock.GetSentMessages()
			for _, msg := range msgs {
				if msg.ID != nil {
					response := &types.BaseJSONRPCMessage{
						JSONRPC: "2.0",
						ID:      msg.ID,
						Result:  json.RawMessage(`{"ok": true}`),
					}
					mock.SimulateReceive(context.Background(), response)
				}
			}
			mock.ClearSentMessages()
		}
	}()

	// Send concurrent requests
	const numRequests = 5
	done := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, err := proto.Request(ctx, "test/method", nil)
			done <- err
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Concurrent request %d failed: %v", i, err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}

func TestProtocol_MethodNotFound(t *testing.T) {
	mock := transport.NewMock()
	proto := NewProtocol()

	if err := proto.Connect(mock); err != nil {
		t.Fatalf("Failed to connect protocol: %v", err)
	}
	defer proto.Close()

	// Simulate request for unregistered method
	request := &types.BaseJSONRPCMessage{
		JSONRPC: "2.0",
		ID:      int64(1),
		Method:  "unknown/method",
		Params:  json.RawMessage(`{}`),
	}

	mock.SimulateReceive(context.Background(), request)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Check that error response was sent
	msgs := mock.GetSentMessages()
	foundError := false
	for _, msg := range msgs {
		if msg.Error != nil && msg.Error.Code == -32601 {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected method not found error")
	}
}
