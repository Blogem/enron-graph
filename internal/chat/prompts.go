package chat

import (
	"fmt"
	"strings"
)

// PromptTemplate holds the structure for chat prompts
type PromptTemplate struct {
	SystemRole          string
	AvailableOperations string
	Schema              string
	ResponseFormat      string
}

// DefaultPromptTemplate returns the default chat prompt template
func DefaultPromptTemplate() *PromptTemplate {
	return &PromptTemplate{
		SystemRole: "You are a graph database assistant helping users query a knowledge graph of Enron email communications and entity relationships.",
		AvailableOperations: `Available query operations:
1. entity_lookup: Find detailed information about a specific person, organization, or concept
   - Use when: User asks "Who is X?", "What is Y?", "Tell me about Z"
   - Returns: Entity details including type, properties, and relationships

2. relationship: Discover relationships and connections for an entity
   - Use when: User asks "Who did X email?", "Who emailed X?", "What did X mention?"
   - Returns: List of related entities

3. path_finding: Find how two entities are connected
   - Use when: User asks "How are X and Y connected?", "What's the relationship between X and Y?"
   - Returns: Shortest path showing relationship chain

4. semantic_search: Search for entities or emails by concept
   - Use when: User asks "Emails about X", "Find discussions about Y"
   - Returns: Semantically similar entities ranked by relevance

5. aggregation: Count relationships or entities
   - Use when: User asks "How many emails did X send?", "How many people did X email?"
   - Returns: Numeric count with description`,
		Schema: `Available entity types:
- person: Individual people (e.g., employees, executives)
- organization: Companies, departments, groups
- concept: Topics, subjects, themes discussed in emails

Available relationship types:
- SENT: person → email (person sent an email)
- RECEIVED: email → person (person received an email)
- MENTIONS: email → organization/concept (email mentions entity)
- COMMUNICATES_WITH: person ↔ person (bidirectional communication)`,
		ResponseFormat: `Response format:
- For direct questions you can answer, respond with plain text
- For queries requiring graph operations, respond with a structured action command
- Always be helpful, concise, and accurate`,
	}
}

// BuildPrompt builds a complete prompt from the template, context, and user query
func BuildPrompt(template *PromptTemplate, conversationHistory []HistoryEntry, trackedEntities map[string]TrackedEntity, userQuery string) string {
	var builder strings.Builder

	// System role
	builder.WriteString(template.SystemRole)
	builder.WriteString("\n\n")

	// Available operations
	builder.WriteString(template.AvailableOperations)
	builder.WriteString("\n\n")

	// Schema information
	builder.WriteString(template.Schema)
	builder.WriteString("\n\n")

	// Response format
	builder.WriteString(template.ResponseFormat)
	builder.WriteString("\n\n")

	// Conversation history
	if len(conversationHistory) > 0 {
		builder.WriteString("Previous conversation:\n")
		for _, entry := range conversationHistory {
			builder.WriteString(fmt.Sprintf("User: %s\n", entry.Query))
			builder.WriteString(fmt.Sprintf("Assistant: %s\n", entry.Response))
		}
		builder.WriteString("\n")
	}

	// Tracked entities for pronoun resolution
	if len(trackedEntities) > 0 {
		builder.WriteString("Recently mentioned entities:\n")
		for _, entity := range trackedEntities {
			builder.WriteString(fmt.Sprintf("- %s (%s)\n", entity.Name, entity.Type))
		}
		builder.WriteString("\n")
	}

	// User query
	builder.WriteString(fmt.Sprintf("User query: %s\n\n", userQuery))
	builder.WriteString("Provide a helpful response or execute the appropriate graph query operation.")

	return builder.String()
}

// BuildExtractionPrompt builds a prompt for entity extraction from emails
func BuildExtractionPrompt(emailSubject, emailBody string, senderEmail string, recipientEmails []string) string {
	var builder strings.Builder

	builder.WriteString("Extract entities and relationships from the following email.\n\n")
	builder.WriteString("Email details:\n")
	builder.WriteString(fmt.Sprintf("Subject: %s\n", emailSubject))
	builder.WriteString(fmt.Sprintf("From: %s\n", senderEmail))
	if len(recipientEmails) > 0 {
		builder.WriteString(fmt.Sprintf("To: %s\n", strings.Join(recipientEmails, ", ")))
	}
	builder.WriteString("\n")
	builder.WriteString("Body:\n")
	builder.WriteString(emailBody)
	builder.WriteString("\n\n")

	builder.WriteString("Extract the following:\n")
	builder.WriteString("1. People mentioned (not including sender/recipients unless explicitly discussed)\n")
	builder.WriteString("2. Organizations mentioned\n")
	builder.WriteString("3. Concepts or topics discussed\n\n")

	builder.WriteString("For each entity, provide:\n")
	builder.WriteString("- name: The entity name\n")
	builder.WriteString("- type: person, organization, or concept\n")
	builder.WriteString("- confidence: 0.0-1.0 (how certain you are this is a distinct entity)\n")
	builder.WriteString("- properties: Any relevant attributes (e.g., role, department, location)\n\n")

	builder.WriteString("Return as JSON array: [{\"name\": \"...\", \"type\": \"...\", \"confidence\": 0.0, \"properties\": {...}}]")

	return builder.String()
}

// BuildDisambiguationPrompt builds a prompt to handle ambiguous queries
func BuildDisambiguationPrompt(query string, options []string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Your query '%s' could have multiple interpretations:\n\n", query))

	for i, option := range options {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, option))
	}

	builder.WriteString("\nPlease specify which interpretation you meant, or rephrase your query to be more specific.")

	return builder.String()
}

// BuildErrorPrompt builds a user-friendly error message
func BuildErrorPrompt(errorType string, details string) string {
	errorMessages := map[string]string{
		"entity_not_found": "I couldn't find that entity in the knowledge graph. Please check the name and try again.",
		"no_path":          "I couldn't find a connection between those entities. They may not be related in the available data.",
		"invalid_query":    "I didn't understand that query. Please try rephrasing it.",
		"llm_error":        "I'm having trouble processing your request right now. Please try again.",
		"timeout":          "The query is taking too long to process. Please try a simpler query or be more specific.",
		"database_error":   "I'm having trouble accessing the knowledge graph. Please try again later.",
	}

	message, ok := errorMessages[errorType]
	if !ok {
		message = "An unexpected error occurred."
	}

	if details != "" {
		return fmt.Sprintf("%s\n\nDetails: %s", message, details)
	}

	return message
}

// SummarizeResults builds a natural language summary of query results
func SummarizeResults(resultType string, count int, entities []string) string {
	switch resultType {
	case "entity_lookup":
		if count == 0 {
			return "No matching entities found."
		}
		if count == 1 {
			return fmt.Sprintf("Found 1 entity: %s", entities[0])
		}
		return fmt.Sprintf("Found %d entities: %s", count, strings.Join(entities[:min(3, len(entities))], ", "))

	case "relationship":
		if count == 0 {
			return "No relationships found."
		}
		if count == 1 {
			return fmt.Sprintf("Found 1 related entity: %s", entities[0])
		}
		return fmt.Sprintf("Found %d related entities including: %s", count, strings.Join(entities[:min(3, len(entities))], ", "))

	case "path":
		if count == 0 {
			return "No connection path found."
		}
		return fmt.Sprintf("Found connection path with %d hops: %s", count, strings.Join(entities, " → "))

	case "search":
		if count == 0 {
			return "No matching results found."
		}
		return fmt.Sprintf("Found %d semantically similar entities", count)

	case "aggregation":
		return fmt.Sprintf("Count: %d", count)

	default:
		return fmt.Sprintf("Query completed with %d results.", count)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
