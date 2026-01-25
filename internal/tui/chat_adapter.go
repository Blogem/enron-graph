package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/chat"
	"github.com/Blogem/enron-graph/internal/graph"
)

// chatRepositoryAdapter adapts graph.Repository to chat.Repository
type chatRepositoryAdapter struct {
	repo graph.Repository
	ctx  context.Context
}

// newChatRepositoryAdapter creates a new repository adapter
func newChatRepositoryAdapter(repo graph.Repository) chat.Repository {
	return &chatRepositoryAdapter{
		repo: repo,
		ctx:  context.Background(),
	}
}

// FindEntityByName finds an entity by name
func (a *chatRepositoryAdapter) FindEntityByName(name string) (*chat.Entity, error) {
	// Query the ent client directly to search by name
	// We'll search for entities whose name contains the query (case-insensitive)
	entities, err := a.repo.FindEntitiesByType(a.ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}

	// Search for matching entity by name (case-insensitive)
	nameLower := strings.ToLower(name)
	for _, entity := range entities {
		if strings.Contains(strings.ToLower(entity.Name), nameLower) {
			return convertToEntity(entity), nil
		}
	}

	return nil, fmt.Errorf("entity not found: %s", name)
}

// TraverseRelationships traverses relationships from an entity
func (a *chatRepositoryAdapter) TraverseRelationships(entityID int, relType string) ([]*chat.Entity, error) {
	// Use the repository's TraverseRelationships method
	entities, err := a.repo.TraverseRelationships(a.ctx, entityID, relType, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to traverse relationships: %w", err)
	}

	// Convert ent.DiscoveredEntity to chat.Entity
	result := make([]*chat.Entity, 0, len(entities))
	for _, entity := range entities {
		result = append(result, convertToEntity(entity))
	}

	return result, nil
}

// FindShortestPath finds the shortest path between two entities
func (a *chatRepositoryAdapter) FindShortestPath(sourceID, targetID int) ([]*chat.PathNode, error) {
	// Use the repository's FindShortestPath method
	relationships, err := a.repo.FindShortestPath(a.ctx, sourceID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to find path: %w", err)
	}

	if len(relationships) == 0 {
		return nil, fmt.Errorf("no path found between entities")
	}

	// Build path nodes from relationships
	path := make([]*chat.PathNode, 0)

	// Add source entity
	sourceEntity, err := a.repo.FindEntityByID(a.ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find source entity: %w", err)
	}

	path = append(path, &chat.PathNode{
		Entity:       convertToEntity(sourceEntity),
		Relationship: "",
	})

	// Add intermediate and target entities
	for _, rel := range relationships {
		targetEntity, err := a.repo.FindEntityByID(a.ctx, rel.ToID)
		if err != nil {
			return nil, fmt.Errorf("failed to find entity in path: %w", err)
		}

		path = append(path, &chat.PathNode{
			Entity:       convertToEntity(targetEntity),
			Relationship: rel.Type,
		})
	}

	return path, nil
}

// SimilaritySearch performs similarity search
func (a *chatRepositoryAdapter) SimilaritySearch(embedding []float32, limit int) ([]*chat.Entity, error) {
	// Use the repository's SimilaritySearch method
	// Using a threshold of 0.7 for similarity
	entities, err := a.repo.SimilaritySearch(a.ctx, embedding, limit, 0.7)
	if err != nil {
		return nil, fmt.Errorf("failed to perform similarity search: %w", err)
	}

	// Convert ent.DiscoveredEntity to chat.Entity
	result := make([]*chat.Entity, 0, len(entities))
	for _, entity := range entities {
		result = append(result, convertToEntity(entity))
	}

	return result, nil
}

// CountRelationships counts relationships for an entity
func (a *chatRepositoryAdapter) CountRelationships(entityID int, relType string) (int, error) {
	// Find all relationships for the entity
	relationships, err := a.repo.FindRelationshipsByEntity(a.ctx, "discovered_entity", entityID)
	if err != nil {
		return 0, fmt.Errorf("failed to find relationships: %w", err)
	}

	// Count relationships matching the type (if specified)
	if relType == "" {
		return len(relationships), nil
	}

	count := 0
	for _, rel := range relationships {
		if rel.Type == relType {
			count++
		}
	}

	return count, nil
}

// convertToEntity converts ent.DiscoveredEntity to chat.Entity
func convertToEntity(entity *ent.DiscoveredEntity) *chat.Entity {
	if entity == nil {
		return nil
	}

	return &chat.Entity{
		ID:         entity.ID,
		Name:       entity.Name,
		Type:       entity.TypeCategory,
		Properties: entity.Properties,
	}
}
