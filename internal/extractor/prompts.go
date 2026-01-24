package extractor

import (
	"fmt"
	"strings"
)

// EntityExtractionPrompt generates a prompt for extracting entities from email
func EntityExtractionPrompt(from, to, subject, body string) string {
	// Truncate body if too long to save tokens
	maxBodyLength := 2000
	if len(body) > maxBodyLength {
		body = body[:maxBodyLength] + "...[truncated]"
	}

	return fmt.Sprintf(`Extract entities from this email and return JSON only, no explanation.

Email:
From: %s
To: %s
Subject: %s
Body:
%s

Extract:
1. Persons: People mentioned or communicated with (name, email if available, confidence 0.0-1.0)
2. Organizations: Companies, departments, groups (name, domain if available, confidence 0.0-1.0)
3. Concepts: Topics, projects, themes (name, keywords, confidence 0.0-1.0)
4. Other: Any other significant entities (type, name, properties, confidence 0.0-1.0)

Rules:
- Confidence > 0.9: Explicitly stated (e.g., email addresses in headers)
- Confidence 0.7-0.9: Clear mentions in content
- Confidence 0.5-0.7: Implied or inferred
- Confidence < 0.5: Uncertain (exclude these)
- Include only entities with confidence >= 0.7
- For persons from headers (From/To), use confidence 1.0 with their email
- For organizations, try to extract domain from email addresses
- For concepts, identify main topics discussed

Return JSON format:
{
  "persons": [{"name": "John Doe", "email": "john.doe@example.com", "confidence": 1.0}],
  "organizations": [{"name": "Enron Corp", "domain": "enron.com", "confidence": 0.9}],
  "concepts": [{"name": "Energy Trading", "keywords": ["trading", "energy", "deals"], "confidence": 0.8}],
  "other": [{"type": "project", "name": "Project X", "properties": {"status": "active"}, "confidence": 0.7}]
}

JSON only, no explanation:`, from, to, subject, body)
}

// ExtractionResult represents the structured output from entity extraction
type ExtractionResult struct {
	Persons       []PersonEntity       `json:"persons"`
	Organizations []OrganizationEntity `json:"organizations"`
	Concepts      []ConceptEntity      `json:"concepts"`
	Other         []GenericEntity      `json:"other"`
}

// PersonEntity represents a person entity
type PersonEntity struct {
	Name       string  `json:"name"`
	Email      string  `json:"email,omitempty"`
	Confidence float64 `json:"confidence"`
}

// OrganizationEntity represents an organization entity
type OrganizationEntity struct {
	Name       string  `json:"name"`
	Domain     string  `json:"domain,omitempty"`
	Confidence float64 `json:"confidence"`
}

// ConceptEntity represents a concept/topic entity
type ConceptEntity struct {
	Name       string   `json:"name"`
	Keywords   []string `json:"keywords,omitempty"`
	Confidence float64  `json:"confidence"`
}

// GenericEntity represents any other entity type
type GenericEntity struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Confidence float64                `json:"confidence"`
}

// CleanJSONResponse attempts to extract JSON from LLM response
func CleanJSONResponse(response string) string {
	// Remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Find JSON object boundaries
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start >= 0 && end > start {
		return response[start : end+1]
	}

	return response
}
