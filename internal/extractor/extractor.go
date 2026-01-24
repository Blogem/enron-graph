package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
)

// Extractor handles entity extraction from emails
type Extractor struct {
	llmClient llm.Client
	repo      graph.Repository
	logger    *slog.Logger
}

// NewExtractor creates a new entity extractor
func NewExtractor(llmClient llm.Client, repo graph.Repository, logger *slog.Logger) *Extractor {
	return &Extractor{
		llmClient: llmClient,
		repo:      repo,
		logger:    logger,
	}
}

// ExtractFromEmail extracts entities and relationships from an email
func (e *Extractor) ExtractFromEmail(ctx context.Context, email *ent.Email) (*ExtractionSummary, error) {
	summary := &ExtractionSummary{}

	// Step 1: Extract high-confidence person entities from email headers
	headerEntities, err := e.extractFromHeaders(ctx, email)
	if err != nil {
		e.logger.Warn("Failed to extract from headers", "error", err)
	} else {
		summary.EntitiesCreated += len(headerEntities)
	}

	// Step 2: Use LLM to extract entities from email content
	llmEntities, err := e.extractFromContent(ctx, email)
	if err != nil {
		e.logger.Warn("Failed to extract from content",
			"message_id", email.MessageID,
			"error", err)
		// Continue with header entities even if LLM fails
	} else {
		summary.EntitiesCreated += len(llmEntities)
	}

	// Step 3: Create relationships between entities and email
	allEntities := append(headerEntities, llmEntities...)
	relationships, err := e.createRelationships(ctx, email, allEntities)
	if err != nil {
		e.logger.Warn("Failed to create relationships", "error", err)
	} else {
		summary.RelationshipsCreated += len(relationships)
	}

	return summary, nil
}

// extractFromHeaders extracts person entities from email headers
func (e *Extractor) extractFromHeaders(ctx context.Context, email *ent.Email) ([]*ent.DiscoveredEntity, error) {
	var entities []*ent.DiscoveredEntity

	// Extract sender
	if email.From != "" {
		entity, err := e.createPersonEntity(ctx, email.From, email.From, 1.0)
		if err != nil {
			e.logger.Warn("Failed to create sender entity", "email", email.From, "error", err)
		} else {
			entities = append(entities, entity)
		}
	}

	// Extract recipients (To, CC, BCC)
	allRecipients := append(email.To, email.Cc...)
	allRecipients = append(allRecipients, email.Bcc...)

	for _, recipient := range allRecipients {
		if recipient == "" {
			continue
		}

		entity, err := e.createPersonEntity(ctx, recipient, recipient, 1.0)
		if err != nil {
			e.logger.Debug("Failed to create recipient entity", "email", recipient, "error", err)
		} else {
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

// extractFromContent uses LLM to extract entities from email content
func (e *Extractor) extractFromContent(ctx context.Context, email *ent.Email) ([]*ent.DiscoveredEntity, error) {
	// Generate extraction prompt
	toStr := strings.Join(email.To, ", ")
	prompt := EntityExtractionPrompt(email.From, toStr, email.Subject, email.Body)

	// Call LLM
	response, err := e.llmClient.GenerateCompletion(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Parse JSON response
	jsonStr := CleanJSONResponse(response)
	var result ExtractionResult

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		e.logger.Warn("Failed to parse LLM response as JSON",
			"response", jsonStr,
			"error", err)
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var entities []*ent.DiscoveredEntity

	// Process persons (skip if already created from headers)
	for _, person := range result.Persons {
		if person.Confidence < 0.7 {
			continue
		}

		uniqueID := person.Email
		if uniqueID == "" {
			// Generate unique ID from name if no email
			uniqueID = fmt.Sprintf("person:%s", strings.ToLower(strings.TrimSpace(person.Name)))
		}

		entity, err := e.createOrUpdateEntity(ctx, uniqueID, "person", person.Name, map[string]interface{}{
			"email": person.Email,
		}, person.Confidence)

		if err != nil {
			e.logger.Debug("Failed to create person entity", "name", person.Name, "error", err)
			continue
		}
		entities = append(entities, entity)
	}

	// Process organizations
	for _, org := range result.Organizations {
		if org.Confidence < 0.7 {
			continue
		}

		uniqueID := fmt.Sprintf("org:%s", strings.ToLower(strings.TrimSpace(org.Name)))

		entity, err := e.createOrUpdateEntity(ctx, uniqueID, "organization", org.Name, map[string]interface{}{
			"domain": org.Domain,
		}, org.Confidence)

		if err != nil {
			e.logger.Debug("Failed to create organization entity", "name", org.Name, "error", err)
			continue
		}
		entities = append(entities, entity)
	}

	// Process concepts
	for _, concept := range result.Concepts {
		if concept.Confidence < 0.7 {
			continue
		}

		uniqueID := fmt.Sprintf("concept:%s", strings.ToLower(strings.TrimSpace(concept.Name)))

		entity, err := e.createOrUpdateEntity(ctx, uniqueID, "concept", concept.Name, map[string]interface{}{
			"keywords": concept.Keywords,
		}, concept.Confidence)

		if err != nil {
			e.logger.Debug("Failed to create concept entity", "name", concept.Name, "error", err)
			continue
		}
		entities = append(entities, entity)
	}

	// Process other entities
	for _, other := range result.Other {
		if other.Confidence < 0.7 {
			continue
		}

		uniqueID := fmt.Sprintf("%s:%s", other.Type, strings.ToLower(strings.TrimSpace(other.Name)))

		entity, err := e.createOrUpdateEntity(ctx, uniqueID, other.Type, other.Name, other.Properties, other.Confidence)

		if err != nil {
			e.logger.Debug("Failed to create entity", "type", other.Type, "name", other.Name, "error", err)
			continue
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

// createPersonEntity creates a person entity with email as unique ID
func (e *Extractor) createPersonEntity(ctx context.Context, email, name string, confidence float64) (*ent.DiscoveredEntity, error) {
	// Check if entity already exists
	existing, err := e.repo.FindEntityByUniqueID(ctx, email)
	if err == nil && existing != nil {
		return existing, nil
	}

	// Extract name from email if name is same as email
	if name == email && strings.Contains(email, "@") {
		parts := strings.Split(email, "@")
		name = parts[0]
	}

	// Generate embedding for the name
	embedding, err := e.llmClient.GenerateEmbedding(ctx, name)
	if err != nil {
		e.logger.Warn("Failed to generate embedding", "name", name, "error", err)
		embedding = make([]float32, 1024) // Empty embedding as fallback
	}

	// Create entity
	return e.repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:     email,
		TypeCategory: "person",
		Name:         name,
		Properties: map[string]interface{}{
			"email": email,
		},
		Embedding:       embedding,
		ConfidenceScore: confidence,
	})
}

// createOrUpdateEntity creates or updates an entity with deduplication
func (e *Extractor) createOrUpdateEntity(ctx context.Context, uniqueID, typeCategory, name string, properties map[string]interface{}, confidence float64) (*ent.DiscoveredEntity, error) {
	// Check if entity already exists
	existing, err := e.repo.FindEntityByUniqueID(ctx, uniqueID)
	if err == nil && existing != nil {
		// Entity exists, return it (could update confidence/properties here)
		return existing, nil
	}

	// Generate embedding for the name
	embedding, err := e.llmClient.GenerateEmbedding(ctx, name)
	if err != nil {
		e.logger.Warn("Failed to generate embedding", "name", name, "error", err)
		embedding = make([]float32, 1024) // Empty embedding as fallback
	}

	// Create new entity
	return e.repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        uniqueID,
		TypeCategory:    typeCategory,
		Name:            name,
		Properties:      properties,
		Embedding:       embedding,
		ConfidenceScore: confidence,
	})
}

// ExtractionSummary summarizes the extraction results
type ExtractionSummary struct {
	EntitiesCreated      int
	RelationshipsCreated int
}
