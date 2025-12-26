package mcpgo_test

import (
	"encoding/json"
	"testing"

	mcpgo "github.com/DR1N0/mcp-go"
)

// Test argument structs
type SimpleArgs struct {
	Name string `json:"name" jsonschema:"required,description=Person's name"`
	Age  int    `json:"age" jsonschema:"description=Person's age"`
}

type NestedArgs struct {
	User    SimpleArgs `json:"user" jsonschema:"required"`
	Active  bool       `json:"active"`
	Tags    []string   `json:"tags" jsonschema:"description=User tags"`
	Ratings []int      `json:"ratings"`
}

type ComplexArgs struct {
	ID       string            `json:"id" jsonschema:"required,description=Unique identifier"`
	Metadata map[string]string `json:"metadata"`
	Count    *int              `json:"count"`
}

// Test handler functions
func simpleHandler(args SimpleArgs) (*mcpgo.ToolResponse, error) {
	return nil, nil
}

func nestedHandler(args NestedArgs) (*mcpgo.ToolResponse, error) {
	return nil, nil
}

func complexHandler(args ComplexArgs) (*mcpgo.ToolResponse, error) {
	return nil, nil
}

func TestGenerateSchema_Simple(t *testing.T) {
	schema, err := mcpgo.GenerateSchema(simpleHandler)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	// Check type
	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", schema["type"])
	}

	// Check properties exist
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties should be a map")
	}

	// Check name property
	nameProp, ok := props["name"].(map[string]interface{})
	if !ok {
		t.Fatal("Name property should exist")
	}
	if nameProp["type"] != "string" {
		t.Errorf("Name type should be 'string', got '%v'", nameProp["type"])
	}
	if nameProp["description"] != "Person's name" {
		t.Errorf("Name description incorrect: %v", nameProp["description"])
	}

	// Check age property
	ageProp, ok := props["age"].(map[string]interface{})
	if !ok {
		t.Fatal("Age property should exist")
	}
	if ageProp["type"] != "integer" {
		t.Errorf("Age type should be 'integer', got '%v'", ageProp["type"])
	}

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Required should be a string array")
	}

	// Check 'name' is required
	foundName := false
	for _, r := range required {
		if r == "name" {
			foundName = true
			break
		}
	}
	if !foundName {
		t.Error("Expected 'name' to be in required fields")
	}
}

func TestGenerateSchema_Nested(t *testing.T) {
	schema, err := mcpgo.GenerateSchema(nestedHandler)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties should be a map")
	}

	// Check nested user object
	userProp, ok := props["user"].(map[string]interface{})
	if !ok {
		t.Fatal("User property should exist")
	}
	if userProp["type"] != "object" {
		t.Errorf("User type should be 'object', got '%v'", userProp["type"])
	}

	// Check array
	tagsProp, ok := props["tags"].(map[string]interface{})
	if !ok {
		t.Fatal("Tags property should exist")
	}
	if tagsProp["type"] != "array" {
		t.Errorf("Tags type should be 'array', got '%v'", tagsProp["type"])
	}
}

func TestGenerateSchema_Required(t *testing.T) {
	schema, err := mcpgo.GenerateSchema(complexHandler)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Required should be a string array")
	}

	// Check 'id' is required
	found := false
	for _, r := range required {
		if r == "id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'id' to be in required fields")
	}
}

func TestGenerateSchema_Map(t *testing.T) {
	schema, err := mcpgo.GenerateSchema(complexHandler)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties should be a map")
	}

	// Check map field
	metadataProp, ok := props["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("Metadata property should exist")
	}
	if metadataProp["type"] != "object" {
		t.Errorf("Metadata type should be 'object', got '%v'", metadataProp["type"])
	}
}

func TestGenerateSchema_Pointer(t *testing.T) {
	schema, err := mcpgo.GenerateSchema(complexHandler)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties should be a map")
	}

	// Pointer fields should still be included
	countProp, ok := props["count"].(map[string]interface{})
	if !ok {
		t.Fatal("Count property should exist even though it's a pointer")
	}
	if countProp["type"] != "integer" {
		t.Errorf("Count type should be 'integer', got '%v'", countProp["type"])
	}
}

func TestGenerateSchema_JSONMarshalable(t *testing.T) {
	schema, err := mcpgo.GenerateSchema(simpleHandler)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	// Schema should be JSON-marshalable
	_, err = json.Marshal(schema)
	if err != nil {
		t.Errorf("Schema should be JSON-marshalable: %v", err)
	}
}

func TestGenerateSchema_Bool(t *testing.T) {
	schema, err := mcpgo.GenerateSchema(nestedHandler)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties should be a map")
	}

	activeProp, ok := props["active"].(map[string]interface{})
	if !ok {
		t.Fatal("Active property should exist")
	}
	if activeProp["type"] != "boolean" {
		t.Errorf("Active type should be 'boolean', got '%v'", activeProp["type"])
	}
}

func TestGenerateSchema_NonFunction(t *testing.T) {
	_, err := mcpgo.GenerateSchema("not a function")
	if err == nil {
		t.Error("Expected error when passing non-function")
	}
}

func TestGenerateSchema_NoArgs(t *testing.T) {
	noArgsHandler := func() (*mcpgo.ToolResponse, error) {
		return nil, nil
	}

	schema, err := mcpgo.GenerateSchema(noArgsHandler)
	if err != nil {
		t.Fatalf("GenerateSchema failed: %v", err)
	}

	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Properties should be a map")
	}

	if len(props) != 0 {
		t.Errorf("Expected empty properties for no-arg function, got %d properties", len(props))
	}
}
