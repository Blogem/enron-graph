package main

import (
	"context"
	"fmt"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/relationship"
	"github.com/Blogem/enron-graph/internal/chat"
)

// chatAdapter implements chat.Repository interface using ent client
type chatAdapter struct {
	client *ent.Client
	ctx    context.Context
}

// newChatAdapter creates a new chat repository adapter
func newChatAdapter(client *ent.Client, ctx context.Context) chat.Repository {
	return &chatAdapter{
		client: client,
		ctx:    ctx,
	}
}

// FindEntityByName finds an entity by name in the database
func (a *chatAdapter) FindEntityByName(name string) (*chat.Entity, error) {
	// Try to find in discovered entities first
	entity, err := a.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.NameEQ(name)).
		First(a.ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("entity not found: %s", name)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Convert to chat.Entity
	return &chat.Entity{
		ID:   entity.ID,
		Name: entity.Name,
		Type: entity.TypeCategory,
		Properties: map[string]interface{}{
			"unique_id":        entity.UniqueID,
			"confidence_score": entity.ConfidenceScore,
			"properties":       entity.Properties,
		},
	}, nil
}

// TraverseRelationships finds all entities connected by a specific relationship type
func (a *chatAdapter) TraverseRelationships(entityID int, relType string) ([]*chat.Entity, error) {
	// Query relationships where the entity is the source
	rels, err := a.client.Relationship.
		Query().
		Where(relationship.FromIDEQ(entityID)).
		Where(relationship.TypeEQ(relType)).
		All(a.ctx)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Get unique target IDs
	targetIDs := make([]int, 0, len(rels))
	for _, rel := range rels {
		targetIDs = append(targetIDs, rel.ToID)
	}

	if len(targetIDs) == 0 {
		return []*chat.Entity{}, nil
	}

	// Fetch target entities
	entities, err := a.client.DiscoveredEntity.
		Query().
		Where(discoveredentity.IDIn(targetIDs...)).
		All(a.ctx)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Convert to chat.Entity slice
	result := make([]*chat.Entity, len(entities))
	for i, entity := range entities {
		result[i] = &chat.Entity{
			ID:   entity.ID,
			Name: entity.Name,
			Type: entity.TypeCategory,
			Properties: map[string]interface{}{
				"unique_id":        entity.UniqueID,
				"confidence_score": entity.ConfidenceScore,
				"properties":       entity.Properties,
			},
		}
	}

	return result, nil
}

// FindShortestPath finds the shortest path between two entities (stub implementation)
func (a *chatAdapter) FindShortestPath(sourceID, targetID int) ([]*chat.PathNode, error) {
	// This is a simplified BFS implementation for finding shortest path
	// In a production system, this might use a more efficient graph algorithm

	// Verify both entities exist
	source, err := a.client.DiscoveredEntity.Get(a.ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("source entity not found: %w", err)
	}

	target, err := a.client.DiscoveredEntity.Get(a.ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("target entity not found: %w", err)
	}

	// For now, just check if there's a direct relationship
	directRel, err := a.client.Relationship.
		Query().
		Where(relationship.FromIDEQ(sourceID)).
		Where(relationship.ToIDEQ(targetID)).
		First(a.ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if directRel != nil {
		// Direct path exists
		return []*chat.PathNode{
			{
				Entity: &chat.Entity{
					ID:   source.ID,
					Name: source.Name,
					Type: source.TypeCategory,
					Properties: map[string]interface{}{
						"unique_id":        source.UniqueID,
						"confidence_score": source.ConfidenceScore,
					},
				},
				Relationship: directRel.Type,
			},
			{
				Entity: &chat.Entity{
					ID:   target.ID,
					Name: target.Name,
					Type: target.TypeCategory,
					Properties: map[string]interface{}{
						"unique_id":        target.UniqueID,
						"confidence_score": target.ConfidenceScore,
					},
				},
				Relationship: "",
			},
		}, nil
	}

	// No direct path found (more complex path finding would go here)
	return nil, fmt.Errorf("no path found between entities")
}

// SimilaritySearch searches for entities similar to the given embedding (stub implementation)
func (a *chatAdapter) SimilaritySearch(embedding []float32, limit int) ([]*chat.Entity, error) {
	// In a production system, this would use vector similarity search
	// For now, just return the most common entities
	entities, err := a.client.DiscoveredEntity.
		Query().
		Order(ent.Desc(discoveredentity.FieldConfidenceScore)).
		Limit(limit).
		All(a.ctx)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Convert to chat.Entity slice
	result := make([]*chat.Entity, len(entities))
	for i, entity := range entities {
		result[i] = &chat.Entity{
			ID:   entity.ID,
			Name: entity.Name,
			Type: entity.TypeCategory,
			Properties: map[string]interface{}{
				"unique_id":        entity.UniqueID,
				"confidence_score": entity.ConfidenceScore,
				"properties":       entity.Properties,
			},
		}
	}

	return result, nil
}

// CountRelationships counts the number of relationships of a specific type for an entity
func (a *chatAdapter) CountRelationships(entityID int, relType string) (int, error) {
	count, err := a.client.Relationship.
		Query().
		Where(relationship.FromIDEQ(entityID)).
		Where(relationship.TypeEQ(relType)).
		Count(a.ctx)

	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}

	return count, nil
}
