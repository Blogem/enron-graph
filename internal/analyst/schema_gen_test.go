package analyst

import (
	"encoding/json"
	"testing"
)

// T078: Unit tests for schema generator
// Tests required property inference (>90% presence), optional property inference (30-90% presence),
// data type inference from samples, validation rule generation, JSON schema output format

func TestInferRequiredProperties(t *testing.T) {
	tests := []struct {
		name               string
		entities           []EntitySample
		threshold          float64
		expectedProperties []string
	}{
		{
			name: "properties above 90% threshold",
			entities: []EntitySample{
				{Properties: map[string]interface{}{"email": "alice@enron.com", "name": "Alice", "title": "VP"}},
				{Properties: map[string]interface{}{"email": "bob@enron.com", "name": "Bob", "title": "Director"}},
				{Properties: map[string]interface{}{"email": "charlie@enron.com", "name": "Charlie", "title": "Manager"}},
				{Properties: map[string]interface{}{"email": "david@enron.com", "name": "David"}}, // missing title
			},
			threshold:          0.90,
			expectedProperties: []string{"email", "name"}, // 100% presence
		},
		{
			name: "all properties required",
			entities: []EntitySample{
				{Properties: map[string]interface{}{"id": 1, "email": "test1@test.com"}},
				{Properties: map[string]interface{}{"id": 2, "email": "test2@test.com"}},
				{Properties: map[string]interface{}{"id": 3, "email": "test3@test.com"}},
			},
			threshold:          0.90,
			expectedProperties: []string{"id", "email"},
		},
		{
			name: "no properties meet threshold",
			entities: []EntitySample{
				{Properties: map[string]interface{}{"name": "Alice", "email": "alice@enron.com"}},
				{Properties: map[string]interface{}{"name": "Bob"}},
				{Properties: map[string]interface{}{"email": "charlie@enron.com"}},
			},
			threshold:          0.90,
			expectedProperties: []string{}, // name: 67%, email: 67%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			required := InferRequiredProperties(tt.entities, tt.threshold)

			if len(required) != len(tt.expectedProperties) {
				t.Errorf("Expected %d required properties, got %d", len(tt.expectedProperties), len(required))
			}

			requiredMap := make(map[string]bool)
			for _, prop := range required {
				requiredMap[prop] = true
			}

			for _, expected := range tt.expectedProperties {
				if !requiredMap[expected] {
					t.Errorf("Expected required property '%s' not found", expected)
				}
			}
		})
	}
}

func TestInferOptionalProperties(t *testing.T) {
	tests := []struct {
		name               string
		entities           []EntitySample
		minThreshold       float64
		maxThreshold       float64
		expectedProperties []string
	}{
		{
			name: "properties in 30-90% range",
			entities: []EntitySample{
				{Properties: map[string]interface{}{"name": "Alice", "email": "alice@enron.com", "title": "VP"}},
				{Properties: map[string]interface{}{"name": "Bob", "email": "bob@enron.com"}},
				{Properties: map[string]interface{}{"name": "Charlie", "email": "charlie@enron.com", "title": "Director"}},
				{Properties: map[string]interface{}{"name": "David", "email": "david@enron.com"}},
			},
			minThreshold:       0.30,
			maxThreshold:       0.90,
			expectedProperties: []string{"title"}, // 50% presence (2/4)
		},
		{
			name: "multiple optional properties",
			entities: []EntitySample{
				{Properties: map[string]interface{}{"name": "Alice", "phone": "123", "department": "Sales"}},
				{Properties: map[string]interface{}{"name": "Bob", "phone": "456"}},
				{Properties: map[string]interface{}{"name": "Charlie", "department": "IT"}},
				{Properties: map[string]interface{}{"name": "David"}},
			},
			minThreshold:       0.30,
			maxThreshold:       0.90,
			expectedProperties: []string{"phone", "department"}, // both 50%
		},
		{
			name: "no optional properties",
			entities: []EntitySample{
				{Properties: map[string]interface{}{"name": "Alice", "email": "alice@enron.com"}},
				{Properties: map[string]interface{}{"name": "Bob", "email": "bob@enron.com"}},
			},
			minThreshold:       0.30,
			maxThreshold:       0.90,
			expectedProperties: []string{}, // name & email are 100% (above max threshold)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optional := InferOptionalProperties(tt.entities, tt.minThreshold, tt.maxThreshold)

			if len(optional) != len(tt.expectedProperties) {
				t.Errorf("Expected %d optional properties, got %d", len(tt.expectedProperties), len(optional))
			}

			optionalMap := make(map[string]bool)
			for _, prop := range optional {
				optionalMap[prop] = true
			}

			for _, expected := range tt.expectedProperties {
				if !optionalMap[expected] {
					t.Errorf("Expected optional property '%s' not found", expected)
				}
			}
		})
	}
}

func TestInferDataTypes(t *testing.T) {
	tests := []struct {
		name             string
		samples          []interface{}
		expectedDataType string
	}{
		{
			name:             "integer values",
			samples:          []interface{}{1, 2, 3, 100},
			expectedDataType: "integer",
		},
		{
			name:             "float values",
			samples:          []interface{}{1.5, 2.7, 3.14},
			expectedDataType: "number",
		},
		{
			name:             "string values",
			samples:          []interface{}{"alice", "bob", "charlie"},
			expectedDataType: "string",
		},
		{
			name:             "boolean values",
			samples:          []interface{}{true, false, true},
			expectedDataType: "boolean",
		},
		{
			name:             "mixed numeric types default to number",
			samples:          []interface{}{1, 2.5, 3},
			expectedDataType: "number",
		},
		{
			name:             "email pattern detected",
			samples:          []interface{}{"alice@enron.com", "bob@enron.com"},
			expectedDataType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataType := InferDataType(tt.samples)

			if dataType != tt.expectedDataType {
				t.Errorf("Expected data type '%s', got '%s'", tt.expectedDataType, dataType)
			}
		})
	}
}

func TestGenerateValidationRules(t *testing.T) {
	tests := []struct {
		name          string
		property      string
		dataType      string
		samples       []interface{}
		expectedRules []ValidationRule
	}{
		{
			name:     "email validation",
			property: "email",
			dataType: "string",
			samples:  []interface{}{"alice@enron.com", "bob@enron.com"},
			expectedRules: []ValidationRule{
				{Type: "format", Value: "email"},
			},
		},
		{
			name:     "string length validation",
			property: "name",
			dataType: "string",
			samples:  []interface{}{"Alice", "Bob", "Charlie"},
			expectedRules: []ValidationRule{
				{Type: "minLength", Value: 1},
				{Type: "maxLength", Value: 100},
			},
		},
		{
			name:     "integer range validation",
			property: "age",
			dataType: "integer",
			samples:  []interface{}{25, 30, 45, 50},
			expectedRules: []ValidationRule{
				{Type: "minimum", Value: 0},
			},
		},
		{
			name:     "no specific rules for generic string",
			property: "description",
			dataType: "string",
			samples:  []interface{}{"test", "another test"},
			expectedRules: []ValidationRule{
				{Type: "minLength", Value: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := GenerateValidationRules(tt.property, tt.dataType, tt.samples)

			// Check if expected rules are present (don't require exact match)
			ruleMap := make(map[string]interface{})
			for _, rule := range rules {
				ruleMap[rule.Type] = rule.Value
			}

			for _, expectedRule := range tt.expectedRules {
				value, ok := ruleMap[expectedRule.Type]
				if !ok {
					t.Errorf("Expected validation rule '%s' not found", expectedRule.Type)
					continue
				}

				// Type-specific comparison
				if expectedRule.Type == "format" {
					if value != expectedRule.Value {
						t.Errorf("Rule '%s': expected value '%v', got '%v'", expectedRule.Type, expectedRule.Value, value)
					}
				}
			}
		})
	}
}

func TestGenerateJSONSchema(t *testing.T) {
	tests := []struct {
		name           string
		typeName       string
		entities       []EntitySample
		expectedSchema SchemaDefinition
	}{
		{
			name:     "person schema with required and optional fields",
			typeName: "person",
			entities: []EntitySample{
				{Properties: map[string]interface{}{"email": "alice@enron.com", "name": "Alice", "title": "VP"}},
				{Properties: map[string]interface{}{"email": "bob@enron.com", "name": "Bob", "title": "Director"}},
				{Properties: map[string]interface{}{"email": "charlie@enron.com", "name": "Charlie"}},
			},
			expectedSchema: SchemaDefinition{
				Type: "person",
				Properties: map[string]PropertyDefinition{
					"email": {Type: "string", Required: true},
					"name":  {Type: "string", Required: true},
					"title": {Type: "string", Required: false}, // 67% presence
				},
			},
		},
		{
			name:     "organization schema",
			typeName: "organization",
			entities: []EntitySample{
				{Properties: map[string]interface{}{"name": "Enron", "industry": "energy"}},
				{Properties: map[string]interface{}{"name": "Dynegy", "industry": "energy"}},
				{Properties: map[string]interface{}{"name": "ExxonMobil", "industry": "oil"}},
			},
			expectedSchema: SchemaDefinition{
				Type: "organization",
				Properties: map[string]PropertyDefinition{
					"name":     {Type: "string", Required: true},
					"industry": {Type: "string", Required: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateJSONSchema(tt.typeName, tt.entities)

			if schema.Type != tt.expectedSchema.Type {
				t.Errorf("Expected type '%s', got '%s'", tt.expectedSchema.Type, schema.Type)
			}

			if len(schema.Properties) != len(tt.expectedSchema.Properties) {
				t.Errorf("Expected %d properties, got %d", len(tt.expectedSchema.Properties), len(schema.Properties))
			}

			for propName, expectedProp := range tt.expectedSchema.Properties {
				actualProp, ok := schema.Properties[propName]
				if !ok {
					t.Errorf("Expected property '%s' not found in schema", propName)
					continue
				}

				if actualProp.Type != expectedProp.Type {
					t.Errorf("Property '%s': expected type '%s', got '%s'", propName, expectedProp.Type, actualProp.Type)
				}

				if actualProp.Required != expectedProp.Required {
					t.Errorf("Property '%s': expected required=%v, got required=%v", propName, expectedProp.Required, actualProp.Required)
				}
			}
		})
	}
}

func TestJSONSchemaOutputFormat(t *testing.T) {
	schema := SchemaDefinition{
		Type: "person",
		Properties: map[string]PropertyDefinition{
			"email": {
				Type:     "string",
				Required: true,
				ValidationRules: []ValidationRule{
					{Type: "format", Value: "email"},
				},
			},
			"name": {
				Type:     "string",
				Required: true,
			},
		},
	}

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled SchemaDefinition
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON schema: %v", err)
	}

	if unmarshaled.Type != schema.Type {
		t.Errorf("Expected type '%s', got '%s' after unmarshal", schema.Type, unmarshaled.Type)
	}

	if len(unmarshaled.Properties) != len(schema.Properties) {
		t.Errorf("Expected %d properties, got %d after unmarshal", len(schema.Properties), len(unmarshaled.Properties))
	}

	// Verify valid JSON structure
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	if err != nil {
		t.Fatalf("Schema is not valid JSON: %v", err)
	}

	if jsonMap["type"] != "person" {
		t.Errorf("JSON schema missing or incorrect 'type' field")
	}

	properties, ok := jsonMap["properties"].(map[string]interface{})
	if !ok || len(properties) != 2 {
		t.Errorf("JSON schema missing or incorrect 'properties' field")
	}
}
