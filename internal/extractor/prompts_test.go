package extractor

import (
	"strings"
	"testing"
)

// TestEntityExtractionPrompt tests the prompt generation
func TestEntityExtractionPrompt(t *testing.T) {
	prompt := EntityExtractionPrompt(
		"jeff.skilling@enron.com",
		"kenneth.lay@enron.com",
		"Q4 Strategy",
		"We need to discuss energy trading.",
	)

	if !strings.Contains(prompt, "jeff.skilling@enron.com") {
		t.Error("Prompt should contain sender email")
	}
	if !strings.Contains(prompt, "Q4 Strategy") {
		t.Error("Prompt should contain subject")
	}
	if !strings.Contains(prompt, "JSON only") {
		t.Error("Prompt should request JSON output")
	}
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
		Persons: []PersonEntity{
			{Name: "John Doe", Email: "john@example.com", Confidence: 0.95},
		},
		Organizations: []OrganizationEntity{
			{Name: "Enron", Domain: "enron.com", Confidence: 0.9},
		},
		Concepts: []ConceptEntity{
			{Name: "Energy Trading", Keywords: []string{"energy", "trading"}, Confidence: 0.85},
		},
	}

	if len(result.Persons) != 1 {
		t.Errorf("Expected 1 person, got %d", len(result.Persons))
	}
	if result.Persons[0].Confidence != 0.95 {
		t.Errorf("Expected confidence 0.95, got %f", result.Persons[0].Confidence)
	}
	if len(result.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(result.Organizations))
	}
	if len(result.Concepts) != 1 {
		t.Errorf("Expected 1 concept, got %d", len(result.Concepts))
	}
}
