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
	// Get previously discovered entity types to enrich the prompt
	discoveredTypes, err := e.repo.GetDistinctEntityTypes(ctx)
	if err != nil {
		e.logger.Warn("Failed to get discovered types, proceeding without them", "error", err)
		discoveredTypes = []string{}
	}

	// Generate extraction prompt with discovered types
	toStr := strings.Join(email.To, ", ")
	prompt := EntityExtractionPrompt(email.From, toStr, email.Subject, email.Body, discoveredTypes)

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

	// Process all entities uniformly
	for _, entity := range result.Entities {
		if entity.Confidence < 0.7 {
			continue
		}

		// Generate unique ID based on type and name
		var uniqueID string

		// Special handling for persons with email
		if entity.Type == "person" {
			if email, ok := entity.Properties["email"].(string); ok && email != "" {
				uniqueID = email
			} else {
				uniqueID = fmt.Sprintf("person:%s", strings.ToLower(strings.TrimSpace(entity.Name)))
			}
		} else {
			uniqueID = fmt.Sprintf("%s:%s", entity.Type, strings.ToLower(strings.TrimSpace(entity.Name)))
		}

		created, err := e.createOrUpdateEntity(ctx, uniqueID, entity.Type, entity.Name, entity.Properties, entity.Confidence)
		if err != nil {
			e.logger.Debug("Failed to create entity",
				"type", entity.Type,
				"name", entity.Name,
				"error", err)
			continue
		}
		entities = append(entities, created)
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
	entity, err := e.repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:     email,
		TypeCategory: "person",
		Name:         name,
		Properties: map[string]interface{}{
			"email": email,
		},
		Embedding:       embedding,
		ConfidenceScore: confidence,
	})

	// If creation failed due to duplicate key, fetch the existing entity
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		existing, fetchErr := e.repo.FindEntityByUniqueID(ctx, email)
		if fetchErr == nil && existing != nil {
			return existing, nil
		}
	}

	return entity, err
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
	entity, err := e.repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        uniqueID,
		TypeCategory:    typeCategory,
		Name:            name,
		Properties:      properties,
		Embedding:       embedding,
		ConfidenceScore: confidence,
	})

	// If creation failed due to duplicate key, fetch the existing entity
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		existing, fetchErr := e.repo.FindEntityByUniqueID(ctx, uniqueID)
		if fetchErr == nil && existing != nil {
			return existing, nil
		}
	}

	return entity, err
}

// ExtractionSummary summarizes the extraction results
type ExtractionSummary struct {
	EntitiesCreated      int
	RelationshipsCreated int
}
