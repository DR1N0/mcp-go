package mcpgo

import (
	"fmt"
	"reflect"
	"strings"
)

// GenerateSchema generates a JSON schema from a function signature
func GenerateSchema(handler interface{}) (map[string]interface{}, error) {
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		return nil, fmt.Errorf("handler must be a function")
	}

	// Determine the argument type
	numIn := handlerType.NumIn()
	var argType reflect.Type

	// Check if first param is context
	if numIn > 0 && handlerType.In(0).Implements(reflect.TypeOf((*interface{ Deadline() })(nil)).Elem()) {
		if numIn > 1 {
			argType = handlerType.In(1)
		}
	} else if numIn > 0 {
		argType = handlerType.In(0)
	}

	// If no arguments, return empty schema
	if argType == nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}, nil
	}

	// Generate schema from struct
	return generateSchemaFromType(argType)
}

// generateSchemaFromType generates JSON schema from a Go type
func generateSchemaFromType(t reflect.Type) (map[string]interface{}, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("argument must be a struct")
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}

	properties := make(map[string]interface{})
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		// Generate field schema
		fieldSchema := generateFieldSchema(field)

		// Check if required (from jsonschema tag or non-pointer type)
		jsonSchemaTag := field.Tag.Get("jsonschema")
		isRequired := strings.Contains(jsonSchemaTag, "required") || field.Type.Kind() != reflect.Ptr

		// Add description if present
		if desc := extractDescription(jsonSchemaTag); desc != "" {
			fieldSchema["description"] = desc
		}

		properties[fieldName] = fieldSchema

		if isRequired {
			required = append(required, fieldName)
		}
	}

	schema["properties"] = properties
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema, nil
}

// generateFieldSchema generates JSON schema for a struct field
func generateFieldSchema(field reflect.StructField) map[string]interface{} {
	t := field.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := make(map[string]interface{})

	switch t.Kind() {
	case reflect.String:
		schema["type"] = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		// Could add items schema here
	case reflect.Map:
		schema["type"] = "object"
	case reflect.Struct:
		schema["type"] = "object"
	default:
		schema["type"] = "string"
	}

	return schema
}

// extractDescription extracts description from jsonschema tag
func extractDescription(tag string) string {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "description=") {
			return strings.TrimPrefix(part, "description=")
		}
	}
	return ""
}
