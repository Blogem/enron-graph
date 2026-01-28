package extractor

import (
	"fmt"
	"regexp"
	"strings"
)

// EntityExtractionPrompt generates a prompt for extracting entities from email
// types contains previously identified entity types to guide the extraction
func EntityExtractionPrompt(from, to, subject, body string, types, relationships []string) string {
	maxBodyLength := 2000
	if len(body) > maxBodyLength {
		body = body[:maxBodyLength] + "...[truncated]"
	}

	return fmt.Sprintf(`### ROLE
You are a headless Knowledge Graph Extraction Service. You output ONLY valid JSON. No conversational filler, no preamble, no markdown formatting.

### ONTOLOGY
Types: [%s]
Relationships: [%s]

### INPUT EMAIL
From: %s
To: %s
Subject: %s
Content: %s

### TASK
1. Extract entities and relationships into the schema below.
2. Normalize names (e.g., "John Doe").
3. Use 'VERB_FORM' for predicates (e.g., 'WORKS_ON').
4. If a type is missing from the ontology, create a specific one.

### JSON SCHEMA
{
  "analysis": "1-sentence summary of the email intent",
  "entities": [{"id": "slug", "type": "type", "name": "Name", "properties": {}, "confidence": 0.0-1.0}],
  "relationships": [{"source_id": "slug", "target_id": "slug", "predicate": "VERB", "context": "reasoning"}]
}

### DATA OUTPUT
{`, // <--- The "Prefix Force" starts here
		strings.Join(types, ", "),
		strings.Join(relationships, ", "),
		from, to, subject, body)
}

// ExtractionResult represents the structured output from entity extraction
type ExtractionResult struct {
	Analysis      string                  `json:"analysis"`
	Entities      []ExtractedEntity       `json:"entities"`
	Relationships []ExtractedRelationship `json:"relationships"`
}

// ExtractedEntity represents a single extracted entity with flexible type
type ExtractedEntity struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Confidence float64                `json:"confidence"`
}

// ExtractedRelationship represents a relationship between two entities
type ExtractedRelationship struct {
	SourceID  string `json:"source_id"`
	TargetID  string `json:"target_id"`
	Predicate string `json:"predicate"`
	Context   string `json:"context"`
}

// // CleanJSONResponse attempts to extract JSON from LLM response
func CleanJSONResponse(response string) string {
	// 1. Try to find a markdown code block first
	re := regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
	match := re.FindStringSubmatch(response)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	// 2. If no code block, find the LARGEST valid-looking JSON object
	// Your current logic finds the FIRST '{', which in your case was a
	// property snippet in the prose.
	// Instead, try to find the last '{' that is followed by 'analysis'
	start := strings.LastIndex(response, "{\n  \"analysis\"")
	if start == -1 {
		start = strings.Index(response, "{")
	}

	end := strings.LastIndex(response, "}")
	if start >= 0 && end > start {
		return response[start : end+1]
	}

	return response
}
