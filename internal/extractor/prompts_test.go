package extractor

import (
	"strings"
	"testing"
)

// TestEntityExtractionPrompt tests the prompt generation
func TestEntityExtractionPrompt(t *testing.T) {
	// Test without discovered types
	prompt := EntityExtractionPrompt(
		"jeff.skilling@enron.com",
		"kenneth.lay@enron.com",
		"Q4 Strategy",
		"We need to discuss energy trading.",
		[]string{},
		[]string{},
	)

	if !strings.Contains(prompt, "jeff.skilling@enron.com") {
		t.Error("Prompt should contain sender email")
	}
	if !strings.Contains(prompt, "Q4 Strategy") {
		t.Error("Prompt should contain subject")
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("Prompt should request JSON output")
	}
	if !strings.Contains(prompt, "Knowledge Graph Extraction") {
		t.Error("Prompt should mention knowledge graph extraction")
	}
}

// TestEntityExtractionPromptWithDiscoveredTypes tests prompt with discovered types
func TestEntityExtractionPromptWithDiscoveredTypes(t *testing.T) {
	discoveredTypes := []string{"person", "organization", "project", "financial_instrument"}

	prompt := EntityExtractionPrompt(
		"jeff.skilling@enron.com",
		"kenneth.lay@enron.com",
		"Q4 Strategy",
		"We need to discuss energy trading.",
		discoveredTypes,
		[]string{},
	)

	if !strings.Contains(prompt, "Types:") {
		t.Error("Prompt should include types section")
	}

	// Check that custom types appear in the types list
	if !strings.Contains(prompt, "project") {
		t.Error("Prompt should include custom discovered types like 'project'")
	}
	if !strings.Contains(prompt, "financial_instrument") {
		t.Error("Prompt should include custom discovered types like 'financial_instrument'")
	}
}

// extractDiscoveredTypesSection extracts just the discovered types section from prompt
func extractDiscoveredTypesSection(prompt string) string {
	start := strings.Index(prompt, "Previously discovered entity types")
	if start == -1 {
		return ""
	}
	end := strings.Index(prompt[start:], "Entity Extraction Guidelines")
	if end == -1 {
		return prompt[start:]
	}
	return prompt[start : start+end]
}

// TestCleanJSONResponse tests JSON extraction from LLM response
func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Plain JSON",
			input:    `{"test": "value"}`,
			expected: `{"test": "value"}`,
		},
		{
			name:     "JSON with markdown",
			input:    "```json\n{\"test\": \"value\"}\n```",
			expected: `{"test": "value"}`,
		},
		{
			name:     "JSON with text before",
			input:    "Here is the result:\n{\"test\": \"value\"}",
			expected: `{"test": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanJSONResponse(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestExtractionResultStructure tests the extraction result types
func TestExtractionResultStructure(t *testing.T) {
	result := ExtractionResult{
		Entities: []ExtractedEntity{
			{Type: "person", Name: "John Doe", Properties: map[string]interface{}{"email": "john@example.com"}, Confidence: 0.95},
			{Type: "organization", Name: "Enron", Properties: map[string]interface{}{"domain": "enron.com"}, Confidence: 0.9},
			{Type: "concept", Name: "Energy Trading", Properties: map[string]interface{}{"keywords": []string{"energy", "trading"}}, Confidence: 0.85},
			{Type: "project", Name: "California Initiative", Properties: map[string]interface{}{"region": "West Coast"}, Confidence: 0.8},
		},
	}

	if len(result.Entities) != 4 {
		t.Errorf("Expected 4 entities, got %d", len(result.Entities))
	}

	if result.Entities[0].Type != "person" {
		t.Errorf("Expected type 'person', got '%s'", result.Entities[0].Type)
	}

	if result.Entities[0].Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", result.Entities[0].Confidence)
	}

	// Check that custom types are properly handled
	if result.Entities[3].Type != "project" {
		t.Errorf("Expected type 'project', got '%s'", result.Entities[3].Type)
	}
}
