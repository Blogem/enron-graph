package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/internal/registry"
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
		e.logger.Debug("Header extraction complete", "entities", len(headerEntities))
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
		e.logger.Debug("Content extraction complete", "entities", len(llmEntities))
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

	// Get previously discovered relationship types to enrich the prompt
	discoveredRelationships, err := e.repo.GetDistinctRelationshipTypes(ctx)
	if err != nil {
		e.logger.Warn("Failed to get discovered relationships, proceeding without them", "error", err)
		discoveredRelationships = []string{}
	}

	// Generate extraction prompt with discovered types
	toStr := strings.Join(email.To, ", ")
	prompt := EntityExtractionPrompt(email.From, toStr, email.Subject, email.Body, discoveredTypes, discoveredRelationships)

	// Call LLM
	response, err := e.llmClient.GenerateCompletion(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Parse JSON response
	jsonStr := CleanJSONResponse("{" + response)
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

		// Special handling for persons with email
		uniqueID := generateUniqueID(entity.Type, entity.ID, entity.Properties)

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

	for _, rel := range result.Relationships {
		// find entity based on source and target IDs and create relationships in the graph
		var source, target *ent.DiscoveredEntity

		sourceUniqueID := generateUniqueID("", rel.SourceID, nil)
		targetUniqueID := generateUniqueID("", rel.TargetID, nil)

		for _, entity := range entities {
			if entity.UniqueID == generateUniqueID(entity.TypeCategory, rel.SourceID, entity.Properties) {
				source = entity
			}
			if entity.UniqueID == generateUniqueID(entity.TypeCategory, rel.TargetID, entity.Properties) {
				target = entity
			}
		}

		if source != nil && target != nil {
			_, err := e.createRelationship(ctx, &graph.RelationshipInput{
				Type:            rel.Predicate,
				FromType:        source.TypeCategory,
				FromID:          source.ID,
				ToType:          target.TypeCategory,
				ToID:            target.ID,
				Timestamp:       email.Date,
				ConfidenceScore: source.ConfidenceScore * target.ConfidenceScore,
				Properties: map[string]interface{}{
					"context": rel.Context,
				},
			})
			if err != nil {
				e.logger.Debug("Failed to create extracted relationship",
					"source_id", rel.SourceID,
					"target_id", rel.TargetID,
					"predicate", rel.Predicate,
					"error", err)
			}
		} else {
			// Log when entities aren't matched for relationships
			e.logger.Debug("Relationship entities not matched",
				"predicate", rel.Predicate,
				"source_id", rel.SourceID,
				"target_id", rel.TargetID,
				"source_matched", source != nil,
				"target_matched", target != nil,
				"looking_for_source", sourceUniqueID,
				"looking_for_target", targetUniqueID,
				"available_entities", len(entities))
		}
	}

	return entities, nil
}

// generateUniqueID generates a unique ID for an entity based on its type and properties
func generateUniqueID(typeCategory, ID string, entityProperties map[string]interface{}) string {
	if typeCategory == "person" {
		if email, ok := entityProperties["email"].(string); ok && email != "" {
			return email
		}
		return fmt.Sprintf("person:%s", strings.ToLower(strings.TrimSpace(ID)))
	}

	return fmt.Sprintf("%s:%s", typeCategory, strings.ToLower(strings.TrimSpace(ID)))
}

// createPersonEntity creates a person entity with email as unique ID
func (e *Extractor) createPersonEntity(ctx context.Context, email, name string, confidence float64) (*ent.DiscoveredEntity, error) {
	// Check if entity already exists
	existing, err := e.repo.FindEntityByUniqueID(ctx, email)
	if err == nil && existing != nil {
		e.logger.Debug("Found existing person entity", "email", email, "id", existing.ID)
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
		// TODO: length needs to match model embedding size
		embedding = make([]float32, 1024) // Empty embedding as fallback
	}

	// Create entity
	entity, err := e.repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:     email,
		TypeCategory: "person",
		Name:         name,
		Properties: map[string]interface{}{
			"email":  email,
			"source": "header",
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
		// TODO: length needs to match model embedding size
		embedding = make([]float32, 1024) // Empty embedding as fallback
	}

	// Create new entity
	if properties == nil {
		properties = make(map[string]interface{})
	}
	properties["source"] = "content"

	// Check if this type has been promoted (exists in registry)
	if createFn, exists := registry.PromotedTypes[typeCategory]; exists {
		e.logger.Debug("Using promoted type creator", "type", typeCategory)

		// Prepare data map for promoted type creator
		data := make(map[string]any)
		data["unique_id"] = uniqueID
		data["name"] = name
		data["confidence_score"] = confidence
		data["embedding"] = embedding

		// Add all custom properties
		for k, v := range properties {
			data[k] = v
		}

		// Add Ent client to context for registry creator functions
		ctxWithClient := context.WithValue(ctx, "entClient", e.repo.GetClient())

		// Call the registered creator function
		result, err := createFn(ctxWithClient, data)
		if err != nil {
			e.logger.Warn("Promoted type creator failed, falling back to DiscoveredEntity",
				"type", typeCategory,
				"error", err)
			// Fall through to use DiscoveredEntity as fallback
		} else {
			// Successfully created with promoted type
			// Note: We return a DiscoveredEntity pointer for compatibility
			// In the future, we might want to return a generic entity interface
			e.logger.Debug("Successfully created promoted entity", "type", typeCategory, "name", name)

			// TODO: this won't work, as FindEntityByUniqueID only returns DiscoveredEntity
			// We need to change FindEntityByUniqueID to return a generic entity interface
			// Then adjust the implementation to fetch from the correct table based on typeCategory
			// If no typeCategory is provided, loop through all promoted types and DiscoveredEntity to find the entity
			// For now, we need to handle the case where the promoted type was created
			// but we still need to return it in a compatible format
			// Since we can't directly convert the promoted type back to DiscoveredEntity,
			// we'll fetch it by uniqueID
			existing, fetchErr := e.repo.FindEntityByUniqueID(ctx, uniqueID)
			if fetchErr == nil && existing != nil {
				return existing, nil
			}

			// If we can't fetch it, log and fall through to DiscoveredEntity
			e.logger.Warn("Created promoted entity but could not fetch it",
				"type", typeCategory,
				"unique_id", uniqueID,
				"result", result)
		}
	}

	// Create as DiscoveredEntity (default path or fallback)
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
