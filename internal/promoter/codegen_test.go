package promoter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T079: Unit tests for ent schema codegen

func TestGenerateEntSchemaFile(t *testing.T) {
	tempDir := t.TempDir()

	schema := SchemaDefinition{
		Type: "person",
		Properties: map[string]PropertyDefinition{
			"email": {Type: "string", Required: true},
			"name":  {Type: "string", Required: true},
		},
	}

	err := GenerateEntSchemaFile(schema, tempDir)
	if err != nil {
		t.Fatalf("GenerateEntSchemaFile failed: %v", err)
	}

	// Check file was created
	filePath := filepath.Join(tempDir, "person.go")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Schema file was not created")
	}

	// Read generated content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	output := string(content)

	// Verify basic structure
	if !strings.Contains(output, "package schema") {
		t.Error("Generated schema missing 'package schema' declaration")
	}

	if !strings.Contains(output, "type Person struct") {
		t.Error("Generated schema missing type Person struct")
	}

	if !strings.Contains(output, "Fields()") {
		t.Error("Generated schema missing Fields() method")
	}
}

func TestMapFieldType(t *testing.T) {
	tests := []struct {
		jsonType string
		expected string
	}{
		{"string", "String"},
		{"integer", "Int"},
		{"number", "Float"},
		{"boolean", "Bool"},
		{"unknown", "String"},
	}

	for _, tt := range tests {
		t.Run(tt.jsonType, func(t *testing.T) {
			result := MapFieldType(tt.jsonType)
			if result != tt.expected {
				t.Errorf("MapFieldType(%s) = %s, want %s", tt.jsonType, result, tt.expected)
			}
		})
	}
}

func TestConvertValidationRules(t *testing.T) {
	tests := []struct {
		name               string
		rules              []ValidationRule
		expectedValidators []string
	}{
		{
			name: "email validation",
			rules: []ValidationRule{
				{Type: "format", Value: "email"},
			},
			expectedValidators: []string{
				"Match(regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`))",
			},
		},
		{
			name: "min length validation",
			rules: []ValidationRule{
				{Type: "minLength", Value: 3},
			},
			expectedValidators: []string{
				"MinLen(3)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ConvertValidationRules(tt.rules)

			if len(output) != len(tt.expectedValidators) {
				t.Errorf("Expected %d validators, got %d", len(tt.expectedValidators), len(output))
			}

			for i, expected := range tt.expectedValidators {
				if i < len(output) && output[i] != expected {
					t.Errorf("Expected validator '%s', got '%s'", expected, output[i])
				}
			}
		})
	}
}

func TestGenerateFieldDefinitions(t *testing.T) {
	schema := SchemaDefinition{
		Type: "test",
		Properties: map[string]PropertyDefinition{
			"email": {Type: "string", Required: true},
			"age":   {Type: "integer", Required: false},
		},
	}

	fields := GenerateFieldDefinitions(schema)

	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}

	// Check field types are mapped correctly
	for _, field := range fields {
		if field.Name == "email" && field.Type != "String" {
			t.Errorf("Expected email field type to be String, got %s", field.Type)
		}
		if field.Name == "age" && field.Type != "Int" {
			t.Errorf("Expected age field type to be Int, got %s", field.Type)
		}
	}
}
