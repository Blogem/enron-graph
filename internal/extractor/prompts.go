package extractor

import (
	"fmt"
	"strings"
)

// EntityExtractionPrompt generates a prompt for extracting entities from email
// discoveredTypes contains previously identified entity types to guide the extraction
func EntityExtractionPrompt(from, to, subject, body string, discoveredTypes []string) string {
	// Truncate body if too long to save tokens
	maxBodyLength := 2000
	if len(body) > maxBodyLength {
		body = body[:maxBodyLength] + "...[truncated]"
	}

	// Build discovered types section
	discoveredTypesText := ""
	if len(discoveredTypes) > 0 {
		filteredTypes := filterCommonTypes(discoveredTypes)
		if len(filteredTypes) > 0 {
			discoveredTypesText = fmt.Sprintf(`
Previously discovered entity types in this dataset:
%s

IMPORTANT: Reuse these types when entities fit these categories. You can also discover new types if needed.

`, "- "+strings.Join(filteredTypes, "\n- "))
		}
	}

	return fmt.Sprintf(`Extract entities from this email and return JSON only, no explanation.

Email:
From: %s
To: %s
Subject: %s
Body:
%s
%s
Entity Extraction Guidelines:

You have FULL FLEXIBILITY to identify entity types that make sense for this content. Think creatively and identify:
- person: People mentioned or involved (include name, email if available)
- organization: Companies, departments, groups (include name, domain if available)  
- concept: Topics, themes, ideas being discussed (include keywords)
- project: Named projects or initiatives
- location: Places, offices, regions mentioned
- event: Meetings, conferences, deadlines
- document: Reports, contracts, files referenced
- product: Services, offerings, commodities
- regulation: Laws, policies, compliance matters
- technology: Systems, platforms, tools
- financial_instrument: Contracts, deals, financial products
- ANY OTHER TYPE that emerges from the content

Rules:
- Confidence > 0.9: Explicitly stated (e.g., email addresses, concrete facts)
- Confidence 0.7-0.9: Clear mentions in content
- Confidence 0.5-0.7: Implied or inferred
- Include only entities with confidence >= 0.7
- Be specific with types - use descriptive type names
- Each entity must have: type, name, confidence
- Optional: properties object for additional context

CRITICAL: Return ONLY a valid JSON object with an "entities" array. Example:

{
  "entities": [
    {"type": "person", "name": "John Doe", "properties": {"email": "john.doe@example.com"}, "confidence": 1.0},
    {"type": "organization", "name": "Enron Corp", "properties": {"domain": "enron.com"}, "confidence": 0.9},
    {"type": "project", "name": "California Power Trading", "properties": {"region": "California"}, "confidence": 0.85},
    {"type": "financial_instrument", "name": "Forward Contract Q4-2001", "properties": {"quarter": "Q4", "year": "2001"}, "confidence": 0.8}
  ]
}

Return ONLY valid JSON. Do not return individual objects separated by commas:`, from, to, subject, body, discoveredTypesText)
}

// filterCommonTypes removes standard types to highlight interesting custom types
func filterCommonTypes(types []string) []string {
	commonTypes := map[string]bool{
		"person":       true,
		"organization": true,
		"concept":      true,
	}

	var filtered []string
	for _, t := range types {
		if !commonTypes[t] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// ExtractionResult represents the structured output from entity extraction
type ExtractionResult struct {
	Entities []ExtractedEntity `json:"entities"`
}

// ExtractedEntity represents a single extracted entity with flexible type
type ExtractedEntity struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Confidence float64                `json:"confidence"`
}

// Legacy structures for backward compatibility
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

	// Check if response is an array (LLM sometimes returns just the array)
	if strings.HasPrefix(response, "[") {
		// Wrap array in entities object
		arrayStart := strings.Index(response, "[")
		arrayEnd := strings.LastIndex(response, "]")
		if arrayStart >= 0 && arrayEnd > arrayStart {
			arrayContent := response[arrayStart : arrayEnd+1]
			return fmt.Sprintf(`{"entities":%s}`, arrayContent)
		}
	}

	// Find JSON object boundaries
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start >= 0 && end > start {
		return response[start : end+1]
	}

	return response
}
