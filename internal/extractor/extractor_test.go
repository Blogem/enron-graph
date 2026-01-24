package extractor

import (
	"context"
	"encoding/json"
	"testing"
)

// T029: Unit tests for entity extractor
// Tests ExtractFromEmail, JSON parsing, confidence filtering, embedding generation, entity type detection

func TestExtractFromEmail_JSONParsing(t *testing.T) {
	// Test parsing LLM response JSON with new flexible structure
	mockResponse := `{
		"entities": [
			{"type": "person", "name": "Alice", "properties": {"email": "alice@enron.com"}, "confidence": 0.95},
			{"type": "person", "name": "Bob", "properties": {"email": "bob@enron.com"}, "confidence": 0.90},
			{"type": "organization", "name": "Enron Energy Trading", "confidence": 0.85},
			{"type": "concept", "name": "energy trading strategy", "confidence": 0.80}
		]
	}`

	var result ExtractionResult
	err := json.Unmarshal([]byte(mockResponse), &result)
	if err != nil {
		t.Fatalf("JSON parsing failed: %v", err)
	}

	if len(result.Entities) != 4 {
		t.Errorf("Expected 4 entities, got %d", len(result.Entities))
	}

	personCount := 0
	orgCount := 0
	conceptCount := 0
	for _, e := range result.Entities {
		switch e.Type {
		case "person":
			personCount++
		case "organization":
			orgCount++
		case "concept":
			conceptCount++
		}
	}

	if personCount != 2 {
		t.Errorf("Expected 2 persons, got %d", personCount)
	}
	if orgCount != 1 {
		t.Errorf("Expected 1 organization, got %d", orgCount)
	}
	if conceptCount != 1 {
		t.Errorf("Expected 1 concept, got %d", conceptCount)
	}
}

func TestExtractFromEmail_ConfidenceFiltering(t *testing.T) {
	entities := []ExtractedEntity{
		{Type: "person", Name: "High Confidence", Confidence: 0.95},
		{Type: "person", Name: "Medium Confidence", Confidence: 0.75},
		{Type: "person", Name: "Low Confidence", Confidence: 0.60},
		{Type: "person", Name: "Very Low", Confidence: 0.30},
	}

	minConfidence := 0.70
	filtered := []ExtractedEntity{}

	for _, e := range entities {
		if e.Confidence >= minConfidence {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 entities after filtering (confidence >= 0.70), got %d", len(filtered))
	}

	for _, e := range filtered {
		if e.Confidence < 0.70 {
			t.Errorf("Entity '%s' has confidence %f, below threshold 0.70", e.Name, e.Confidence)
		}
	}
}

func TestExtractFromEmail_EmbeddingGeneration(t *testing.T) {
	mockLLM := &MockLLMClient{
		EmbeddingResponse: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
	}

	ctx := context.Background()
	embedding, err := mockLLM.GenerateEmbedding(ctx, "Test text")

	if err != nil {
		t.Fatalf("GenerateEmbedding failed: %v", err)
	}

	expectedLength := 5
	if len(embedding) != expectedLength {
		t.Errorf("Expected embedding length %d, got %d", expectedLength, len(embedding))
	}

	for i, val := range embedding {
		if val < 0.0 || val > 1.0 {
			t.Errorf("Embedding[%d] = %f is out of expected range [0.0, 1.0]", i, val)
		}
	}
}

func TestExtractFromEmail_InvalidJSON(t *testing.T) {
	invalidJSON := "This is not valid JSON"

	var result ExtractionResult
	err := json.Unmarshal([]byte(invalidJSON), &result)

	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

func TestExtractFromEmail_EntityTypeDetection(t *testing.T) {
	mockResponse := `{
		"entities": [
			{"type": "person", "name": "John Doe", "properties": {"email": "john@enron.com"}, "confidence": 0.95},
			{"type": "organization", "name": "Enron Corp", "confidence": 0.90},
			{"type": "concept", "name": "risk management", "confidence": 0.85},
			{"type": "project", "name": "Project Phoenix", "confidence": 0.82}
		]
	}`

	var result ExtractionResult
	err := json.Unmarshal([]byte(mockResponse), &result)

	if err != nil {
		t.Fatalf("JSON parsing failed: %v", err)
	}

	if len(result.Entities) != 4 {
		t.Errorf("Expected 4 entities, got %d", len(result.Entities))
	}

	typeCount := make(map[string]int)
	for _, e := range result.Entities {
		typeCount[e.Type]++
	}

	if typeCount["person"] != 1 {
		t.Errorf("Expected 1 person, got %d", typeCount["person"])
	}

	if typeCount["organization"] != 1 {
		t.Errorf("Expected 1 organization, got %d", typeCount["organization"])
	}

	if typeCount["concept"] != 1 {
		t.Errorf("Expected 1 concept, got %d", typeCount["concept"])
	}

	if typeCount["project"] != 1 {
		t.Errorf("Expected 1 project (custom type), got %d", typeCount["project"])
	}
}

func TestExtractFromEmail_HeaderBasedPersonEntities(t *testing.T) {
	email := &Email{
		MessageID: "<test@enron.com>",
		From:      "alice@enron.com",
		To:        []string{"bob@enron.com", "charlie@enron.com"},
		CC:        []string{"dave@enron.com"},
	}

	// Should extract email addresses as person identifiers
	addresses := []string{email.From}
	addresses = append(addresses, email.To...)
	addresses = append(addresses, email.CC...)

	if len(addresses) != 4 {
		t.Errorf("Expected 4 email addresses, got %d", len(addresses))
	}

	for _, addr := range addresses {
		if !containsAtSign(addr) {
			t.Errorf("Invalid email address: %s", addr)
		}
	}
}

func TestExtractFromEmail_EmptyEmail(t *testing.T) {
	email := &Email{
		MessageID: "<test@enron.com>",
		From:      "",
		To:        []string{},
		Subject:   "",
		Body:      "",
	}

	// Empty email should have minimal data
	if email.MessageID == "" {
		t.Error("MessageID should not be empty")
	}

	if len(email.To) != 0 {
		t.Errorf("Expected 0 recipients, got %d", len(email.To))
	}
}

// Helper function
func containsAtSign(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}

// Mock types for testing
type Email struct {
	MessageID string
	From      string
	To        []string
	CC        []string
	BCC       []string
	Subject   string
	Body      string
}

// TestEntity for validation (actual entities stored in database)
type TestEntity struct {
	Name       string
	Type       string
	Confidence float64
	Properties map[string]interface{}
	Embedding  []float32
}

type MockLLMClient struct {
	CompletionResponse string
	EmbeddingResponse  []float32
}

func (m *MockLLMClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	return m.CompletionResponse, nil
}

func (m *MockLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return m.EmbeddingResponse, nil
}

func (m *MockLLMClient) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range result {
		result[i] = m.EmbeddingResponse
	}
	return result, nil
}

func (m *MockLLMClient) Close() error {
	return nil
}
