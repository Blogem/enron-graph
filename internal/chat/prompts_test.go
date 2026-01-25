package chat

import (
	"strings"
	"testing"
	"time"
)

// TestDefaultPromptTemplate tests the default prompt template structure
func TestDefaultPromptTemplate(t *testing.T) {
	template := DefaultPromptTemplate()

	if template.SystemRole == "" {
		t.Error("SystemRole should not be empty")
	}
	if template.AvailableOperations == "" {
		t.Error("AvailableOperations should not be empty")
	}
	if template.Schema == "" {
		t.Error("Schema should not be empty")
	}
	if template.ResponseFormat == "" {
		t.Error("ResponseFormat should not be empty")
	}

	// Check that system role mentions graph database
	if !strings.Contains(strings.ToLower(template.SystemRole), "graph") {
		t.Error("SystemRole should mention graph database")
	}

	// Check that schema includes entity types
	if !strings.Contains(template.Schema, "person") ||
		!strings.Contains(template.Schema, "organization") ||
		!strings.Contains(template.Schema, "concept") {
		t.Error("Schema should include all entity types")
	}

	// Check that operations are documented
	if !strings.Contains(template.AvailableOperations, "entity_lookup") ||
		!strings.Contains(template.AvailableOperations, "relationship") ||
		!strings.Contains(template.AvailableOperations, "path_finding") {
		t.Error("AvailableOperations should include all query types")
	}
}

// TestBuildPrompt tests building a complete prompt
func TestBuildPrompt(t *testing.T) {
	template := DefaultPromptTemplate()
	history := []HistoryEntry{
		{
			Query:     "Who is Jeff Skilling?",
			Response:  "Jeff Skilling is the CEO...",
			Timestamp: time.Now(),
		},
	}
	entities := map[string]TrackedEntity{
		"Jeff Skilling": {
			Name: "Jeff Skilling",
			Type: "person",
			ID:   1,
		},
	}
	query := "What did he do?"

	prompt := BuildPrompt(template, history, entities, query)

	// Check that all components are included
	if !strings.Contains(prompt, template.SystemRole) {
		t.Error("Prompt should include system role")
	}
	if !strings.Contains(prompt, "Previous conversation") {
		t.Error("Prompt should include conversation history")
	}
	if !strings.Contains(prompt, "Jeff Skilling") {
		t.Error("Prompt should include tracked entities")
	}
	if !strings.Contains(prompt, query) {
		t.Error("Prompt should include user query")
	}
}

// TestBuildPromptWithoutHistory tests prompt building without history
func TestBuildPromptWithoutHistory(t *testing.T) {
	template := DefaultPromptTemplate()
	query := "Who is Kenneth Lay?"

	prompt := BuildPrompt(template, nil, nil, query)

	// Should include system components but not history
	if !strings.Contains(prompt, template.SystemRole) {
		t.Error("Prompt should include system role")
	}
	if strings.Contains(prompt, "Previous conversation") {
		t.Error("Prompt should not include conversation history when empty")
	}
	if !strings.Contains(prompt, query) {
		t.Error("Prompt should include user query")
	}
}

// TestBuildExtractionPrompt tests extraction prompt building
func TestBuildExtractionPrompt(t *testing.T) {
	subject := "Q3 Earnings Discussion"
	body := "We need to discuss the quarterly earnings with the accounting team."
	sender := "jeff.skilling@enron.com"
	recipients := []string{"kenneth.lay@enron.com", "andrew.fastow@enron.com"}

	prompt := BuildExtractionPrompt(subject, body, sender, recipients)

	// Check that email details are included
	if !strings.Contains(prompt, subject) {
		t.Error("Prompt should include email subject")
	}
	if !strings.Contains(prompt, body) {
		t.Error("Prompt should include email body")
	}
	if !strings.Contains(prompt, sender) {
		t.Error("Prompt should include sender")
	}
	if !strings.Contains(prompt, recipients[0]) {
		t.Error("Prompt should include recipients")
	}

	// Check that extraction instructions are present
	if !strings.Contains(prompt, "Extract") {
		t.Error("Prompt should include extraction instructions")
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("Prompt should request JSON format")
	}
}

// TestBuildDisambiguationPrompt tests disambiguation prompt building
func TestBuildDisambiguationPrompt(t *testing.T) {
	query := "Who is John?"
	options := []string{
		"Did you mean John Smith?",
		"Did you mean John Doe?",
	}

	prompt := BuildDisambiguationPrompt(query, options)

	// Check that query is included
	if !strings.Contains(prompt, query) {
		t.Error("Prompt should include original query")
	}

	// Check that all options are listed
	for _, option := range options {
		if !strings.Contains(prompt, option) {
			t.Errorf("Prompt should include option: %s", option)
		}
	}

	// Check for clarification request
	if !strings.Contains(strings.ToLower(prompt), "specify") ||
		!strings.Contains(strings.ToLower(prompt), "rephrase") {
		t.Error("Prompt should ask for clarification")
	}
}

// TestBuildErrorPrompt tests error message building
func TestBuildErrorPrompt(t *testing.T) {
	tests := []struct {
		name       string
		errorType  string
		details    string
		wantPhrase string
	}{
		{
			name:       "entity not found",
			errorType:  "entity_not_found",
			details:    "",
			wantPhrase: "couldn't find",
		},
		{
			name:       "no path found",
			errorType:  "no_path",
			details:    "",
			wantPhrase: "connection",
		},
		{
			name:       "invalid query",
			errorType:  "invalid_query",
			details:    "",
			wantPhrase: "didn't understand",
		},
		{
			name:       "with details",
			errorType:  "llm_error",
			details:    "Connection timeout",
			wantPhrase: "Details:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := BuildErrorPrompt(tt.errorType, tt.details)

			if !strings.Contains(strings.ToLower(message), strings.ToLower(tt.wantPhrase)) {
				t.Errorf("Error message should contain phrase %q, got: %s", tt.wantPhrase, message)
			}

			if tt.details != "" && !strings.Contains(message, tt.details) {
				t.Errorf("Error message should include details: %s", tt.details)
			}
		})
	}
}

// TestSummarizeResults tests result summarization
func TestSummarizeResults(t *testing.T) {
	tests := []struct {
		name       string
		resultType string
		count      int
		entities   []string
		wantPhrase string
	}{
		{
			name:       "entity lookup - single",
			resultType: "entity_lookup",
			count:      1,
			entities:   []string{"Jeff Skilling"},
			wantPhrase: "Found 1 entity",
		},
		{
			name:       "entity lookup - multiple",
			resultType: "entity_lookup",
			count:      3,
			entities:   []string{"Jeff Skilling", "Kenneth Lay", "Andrew Fastow"},
			wantPhrase: "Found 3 entities",
		},
		{
			name:       "relationship - none",
			resultType: "relationship",
			count:      0,
			entities:   []string{},
			wantPhrase: "No relationships found",
		},
		{
			name:       "path",
			resultType: "path",
			count:      2,
			entities:   []string{"Jeff Skilling", "Email", "Kenneth Lay"},
			wantPhrase: "→",
		},
		{
			name:       "search",
			resultType: "search",
			count:      10,
			entities:   []string{},
			wantPhrase: "10 semantically similar",
		},
		{
			name:       "aggregation",
			resultType: "aggregation",
			count:      42,
			entities:   []string{},
			wantPhrase: "Count: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := SummarizeResults(tt.resultType, tt.count, tt.entities)

			if !strings.Contains(summary, tt.wantPhrase) {
				t.Errorf("Summary should contain %q, got: %s", tt.wantPhrase, summary)
			}
		})
	}
}

// TestPromptTemplateConsistency tests that prompts are consistent
func TestPromptTemplateConsistency(t *testing.T) {
	template1 := DefaultPromptTemplate()
	template2 := DefaultPromptTemplate()

	// Templates should be identical when created multiple times
	if template1.SystemRole != template2.SystemRole {
		t.Error("Default templates should have consistent SystemRole")
	}
	if template1.AvailableOperations != template2.AvailableOperations {
		t.Error("Default templates should have consistent AvailableOperations")
	}
}

// TestPromptLength tests that prompts don't exceed reasonable lengths
func TestPromptLength(t *testing.T) {
	template := DefaultPromptTemplate()

	// Create a large history
	history := make([]HistoryEntry, 5)
	for i := 0; i < 5; i++ {
		history[i] = HistoryEntry{
			Query:     "Sample query " + string(rune(i)),
			Response:  "Sample response " + string(rune(i)),
			Timestamp: time.Now(),
		}
	}

	entities := make(map[string]TrackedEntity)
	for i := 0; i < 10; i++ {
		name := "Entity " + string(rune(i))
		entities[name] = TrackedEntity{
			Name: name,
			Type: "person",
			ID:   i,
		}
	}

	prompt := BuildPrompt(template, history, entities, "Test query")

	// Prompt should be under 10KB for reasonable LLM context
	if len(prompt) > 10000 {
		t.Errorf("Prompt length %d exceeds reasonable limit of 10000 characters", len(prompt))
	}
}
