package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// chatHandler implements the Handler interface for processing chat queries
type chatHandler struct {
	llm       LLMClient
	repo      Repository
	formatter ResponseFormatter
}

// NewHandler creates a new chat handler
func NewHandler(llm LLMClient, repo Repository) Handler {
	return &chatHandler{
		llm:       llm,
		repo:      repo,
		formatter: NewResponseFormatter(),
	}
}

// llmResponse represents the structured response from the LLM
type llmResponse struct {
	Action       string                 `json:"action"`
	Entity       string                 `json:"entity,omitempty"`
	Source       string                 `json:"source,omitempty"`
	Target       string                 `json:"target,omitempty"`
	RelType      string                 `json:"rel_type,omitempty"`
	Relationship string                 `json:"relationship,omitempty"` // Alternative field name for rel_type
	Text         string                 `json:"text,omitempty"`
	Answer       string                 `json:"answer,omitempty"`
	EntityType   string                 `json:"entity_type,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// ProcessQuery processes a user query and returns a response
func (h *chatHandler) ProcessQuery(ctx context.Context, query string, chatContext Context) (string, error) {
	// Build prompt with conversation context
	promptContext := chatContext.BuildPromptContext(query)

	// Create system prompt with schema information
	systemPrompt := h.buildSystemPrompt()

	// Combine system prompt and user context
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, promptContext)

	// Call LLM to process the query
	llmOutput, err := h.llm.GenerateCompletion(ctx, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("LLM error: %w", err)
	}

	// Try to parse as JSON for structured query
	var llmResp llmResponse
	if err := json.Unmarshal([]byte(llmOutput), &llmResp); err != nil {
		// If not valid JSON, treat as direct text response
		response := llmOutput
		chatContext.AddQuery(query, response)
		return response, nil
	}

	// Execute the appropriate action based on LLM response
	response, err := h.executeAction(ctx, &llmResp, chatContext)
	if err != nil {
		// Handle errors gracefully - some errors like "entity not found" should return a user-friendly message
		if strings.Contains(err.Error(), "not found") {
			response = fmt.Sprintf("I couldn't find that entity. %s", err.Error())
			chatContext.AddQuery(query, response)
			return response, nil
		}
		return "", err
	}

	// Add to conversation history
	chatContext.AddQuery(query, response)

	return response, nil
}

// buildSystemPrompt creates the system prompt with schema information
func (h *chatHandler) buildSystemPrompt() string {
	return `You are a graph database assistant. You help users query a knowledge graph of emails and entities.

Available actions:
- entity_lookup: Find an entity by name (respond with JSON: {"action": "entity_lookup", "entity": "name"})
- relationship: Find relationships for an entity (respond with JSON: {"action": "relationship", "entity": "name", "rel_type": "SENT|RECEIVED|MENTIONS|COMMUNICATES_WITH"})
- path_finding: Find the shortest path between two entities (respond with JSON: {"action": "path_finding", "source": "name1", "target": "name2"})
- semantic_search: Search for entities by concept (respond with JSON: {"action": "semantic_search", "text": "search text"})
- aggregation: Count relationships (respond with JSON: {"action": "aggregation", "entity": "name", "rel_type": "SENT|RECEIVED"})

Entity types: person, organization, concept
Relationship types: SENT, RECEIVED, MENTIONS, COMMUNICATES_WITH

When the query is a simple question that can be answered directly, respond with JSON: {"action": "answer", "answer": "your response"}

Always respond with valid JSON.`
}

// executeAction executes the action specified by the LLM response
func (h *chatHandler) executeAction(ctx context.Context, resp *llmResponse, chatContext Context) (string, error) {
	// Support alternative field name for relationship type
	relType := resp.RelType
	if relType == "" {
		relType = resp.Relationship
	}

	switch resp.Action {
	case "entity_lookup":
		return h.executeEntityLookup(resp.Entity, chatContext)
	case "relationship", "traverse":
		return h.executeRelationship(resp.Entity, relType, chatContext)
	case "path_finding", "find_path":
		return h.executePathFinding(resp.Source, resp.Target, chatContext)
	case "semantic_search":
		return h.executeSemanticSearch(ctx, resp.Text)
	case "aggregation", "count":
		return h.executeAggregation(resp.Entity, relType, chatContext)
	case "answer":
		return resp.Answer, nil
	default:
		return "", fmt.Errorf("unknown action: %s", resp.Action)
	}
}

// executeEntityLookup finds an entity by name
func (h *chatHandler) executeEntityLookup(entityName string, chatContext Context) (string, error) {
	entity, err := h.repo.FindEntityByName(entityName)
	if err != nil {
		return "", fmt.Errorf("entity lookup failed: %w", err)
	}

	// Track the entity for pronoun resolution
	chatContext.TrackEntity(entity.Name, entity.Type, entity.ID)

	return h.formatter.FormatEntities([]*Entity{entity}), nil
}

// executeRelationship finds relationships for an entity
func (h *chatHandler) executeRelationship(entityName, relType string, chatContext Context) (string, error) {
	// First find the entity
	entity, err := h.repo.FindEntityByName(entityName)
	if err != nil {
		return "", fmt.Errorf("entity lookup failed: %w", err)
	}

	// Track the entity
	chatContext.TrackEntity(entity.Name, entity.Type, entity.ID)

	// Traverse relationships
	relatedEntities, err := h.repo.TraverseRelationships(entity.ID, relType)
	if err != nil {
		return "", fmt.Errorf("relationship traversal failed: %w", err)
	}

	// Track related entities
	for _, related := range relatedEntities {
		chatContext.TrackEntity(related.Name, related.Type, related.ID)
	}

	return h.formatter.FormatEntities(relatedEntities), nil
}

// executePathFinding finds the shortest path between two entities
func (h *chatHandler) executePathFinding(sourceName, targetName string, chatContext Context) (string, error) {
	// Find source entity
	source, err := h.repo.FindEntityByName(sourceName)
	if err != nil {
		return "", fmt.Errorf("source entity lookup failed: %w", err)
	}

	// Find target entity
	target, err := h.repo.FindEntityByName(targetName)
	if err != nil {
		return "", fmt.Errorf("target entity lookup failed: %w", err)
	}

	// Track entities
	chatContext.TrackEntity(source.Name, source.Type, source.ID)
	chatContext.TrackEntity(target.Name, target.Type, target.ID)

	// Find shortest path
	path, err := h.repo.FindShortestPath(source.ID, target.ID)
	if err != nil {
		return "", fmt.Errorf("path finding failed: %w", err)
	}

	return h.formatter.FormatPath(path), nil
}

// executeSemanticSearch performs semantic search for entities
func (h *chatHandler) executeSemanticSearch(ctx context.Context, searchText string) (string, error) {
	// Generate embedding for search text
	embedding, err := h.llm.GenerateEmbedding(ctx, searchText)
	if err != nil {
		return "", fmt.Errorf("embedding generation failed: %w", err)
	}

	// Search for similar entities
	entities, err := h.repo.SimilaritySearch(embedding, 10)
	if err != nil {
		return "", fmt.Errorf("similarity search failed: %w", err)
	}

	return h.formatter.FormatEntities(entities), nil
}

// executeAggregation counts relationships for an entity
func (h *chatHandler) executeAggregation(entityName, relType string, chatContext Context) (string, error) {
	// Find the entity
	entity, err := h.repo.FindEntityByName(entityName)
	if err != nil {
		return "", fmt.Errorf("entity lookup failed: %w", err)
	}

	// Track the entity
	chatContext.TrackEntity(entity.Name, entity.Type, entity.ID)

	// Count relationships
	count, err := h.repo.CountRelationships(entity.ID, relType)
	if err != nil {
		return "", fmt.Errorf("relationship counting failed: %w", err)
	}

	description := fmt.Sprintf("%s %s relationships for %s", relType, entity.Type, entity.Name)
	return h.formatter.FormatCount(count, description), nil
}

// responseFormatter implements the ResponseFormatter interface
type responseFormatter struct{}

// NewResponseFormatter creates a new response formatter
func NewResponseFormatter() ResponseFormatter {
	return &responseFormatter{}
}

// FormatEntities formats a list of entities into a readable string
func (f *responseFormatter) FormatEntities(entities []*Entity) string {
	if len(entities) == 0 {
		return "No entities found."
	}

	var builder strings.Builder

	if len(entities) == 1 {
		entity := entities[0]
		builder.WriteString(fmt.Sprintf("%s (%s)\n", entity.Name, entity.Type))
		if len(entity.Properties) > 0 {
			builder.WriteString("Properties:\n")
			for key, value := range entity.Properties {
				builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
		}
	} else {
		builder.WriteString(fmt.Sprintf("Found %d entities:\n", len(entities)))
		for i, entity := range entities {
			builder.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, entity.Name, entity.Type))
		}
	}

	return builder.String()
}

// FormatPath formats a path into a readable string
func (f *responseFormatter) FormatPath(path []*PathNode) string {
	if len(path) == 0 {
		return "No path found."
	}

	var builder strings.Builder
	builder.WriteString("Path found:\n")

	for i, node := range path {
		builder.WriteString(fmt.Sprintf("%s (%s)", node.Entity.Name, node.Entity.Type))
		if i < len(path)-1 {
			builder.WriteString(fmt.Sprintf(" -[%s]-> ", node.Relationship))
		}
	}
	builder.WriteString("\n")

	return builder.String()
}

// FormatCount formats a count result into a readable string
func (f *responseFormatter) FormatCount(count int, description string) string {
	return fmt.Sprintf("%s: %d\n", description, count)
}
