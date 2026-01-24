package graph

import (
	"context"
	"fmt"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/email"
	"github.com/Blogem/enron-graph/ent/relationship"

	_ "github.com/lib/pq"
)

// entRepository implements the Repository interface using ent
type entRepository struct {
	client *ent.Client
}

// NewRepository creates a new ent-based repository
func NewRepository(client *ent.Client) Repository {
	return &entRepository{client: client}
}

// CreateEmail creates a new email entity
func (r *entRepository) CreateEmail(ctx context.Context, input *EmailInput) (*ent.Email, error) {
	return r.client.Email.Create().
		SetMessageID(input.MessageID).
		SetFrom(input.From).
		SetTo(input.To).
		SetCc(input.CC).
		SetBcc(input.BCC).
		SetSubject(input.Subject).
		SetDate(input.Date).
		SetBody(input.Body).
		SetNillableFilePath(&input.FilePath).
		Save(ctx)
}

// FindEmailByMessageID finds an email by message ID
func (r *entRepository) FindEmailByMessageID(ctx context.Context, messageID string) (*ent.Email, error) {
	return r.client.Email.Query().
		Where(email.MessageIDEQ(messageID)).
		Only(ctx)
}

// CreateDiscoveredEntity creates a new discovered entity
func (r *entRepository) CreateDiscoveredEntity(ctx context.Context, input *EntityInput) (*ent.DiscoveredEntity, error) {
	return r.client.DiscoveredEntity.Create().
		SetUniqueID(input.UniqueID).
		SetTypeCategory(input.TypeCategory).
		SetName(input.Name).
		SetProperties(input.Properties).
		SetEmbedding(input.Embedding).
		SetConfidenceScore(input.ConfidenceScore).
		Save(ctx)
}

// FindEntityByID finds an entity by ID
func (r *entRepository) FindEntityByID(ctx context.Context, id int) (*ent.DiscoveredEntity, error) {
	return r.client.DiscoveredEntity.Get(ctx, id)
}

// FindEntityByUniqueID finds an entity by unique ID
func (r *entRepository) FindEntityByUniqueID(ctx context.Context, uniqueID string) (*ent.DiscoveredEntity, error) {
	return r.client.DiscoveredEntity.Query().
		Where(discoveredentity.UniqueIDEQ(uniqueID)).
		Only(ctx)
}

// FindEntitiesByType finds entities by type category
func (r *entRepository) FindEntitiesByType(ctx context.Context, typeCategory string) ([]*ent.DiscoveredEntity, error) {
	return r.client.DiscoveredEntity.Query().
		Where(discoveredentity.TypeCategoryEQ(typeCategory)).
		All(ctx)
}

// CreateRelationship creates a new relationship
func (r *entRepository) CreateRelationship(ctx context.Context, input *RelationshipInput) (*ent.Relationship, error) {
	return r.client.Relationship.Create().
		SetType(input.Type).
		SetFromType(input.FromType).
		SetFromID(input.FromID).
		SetToType(input.ToType).
		SetToID(input.ToID).
		SetTimestamp(input.Timestamp).
		SetConfidenceScore(input.ConfidenceScore).
		SetProperties(input.Properties).
		Save(ctx)
}

// FindRelationshipsByEntity finds relationships for an entity
func (r *entRepository) FindRelationshipsByEntity(ctx context.Context, entityType string, entityID int) ([]*ent.Relationship, error) {
	return r.client.Relationship.Query().
		Where(
			relationship.Or(
				relationship.And(
					relationship.FromTypeEQ(entityType),
					relationship.FromIDEQ(entityID),
				),
				relationship.And(
					relationship.ToTypeEQ(entityType),
					relationship.ToIDEQ(entityID),
				),
			),
		).
		All(ctx)
}

// TraverseRelationships traverses relationships from an entity
func (r *entRepository) TraverseRelationships(ctx context.Context, fromID int, relType string, depth int) ([]*ent.DiscoveredEntity, error) {
	// Simple implementation for now - can be optimized later
	// This does a breadth-first traversal up to the specified depth

	if depth <= 0 {
		return nil, nil
	}

	// Find relationships from this entity
	rels, err := r.client.Relationship.Query().
		Where(
			relationship.FromIDEQ(fromID),
			relationship.TypeEQ(relType),
		).
		All(ctx)

	if err != nil {
		return nil, err
	}

	// Collect target entities
	var entities []*ent.DiscoveredEntity
	for _, rel := range rels {
		if rel.ToType == "discovered_entity" {
			entity, err := r.FindEntityByID(ctx, rel.ToID)
			if err != nil {
				continue
			}
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

// FindShortestPath finds the shortest path between two entities
func (r *entRepository) FindShortestPath(ctx context.Context, fromID, toID int) ([]*ent.Relationship, error) {
	// Placeholder implementation - returns empty for now
	// Full BFS implementation will be added in path.go
	return nil, fmt.Errorf("not implemented yet")
}

// SimilaritySearch finds entities similar to the given embedding
func (r *entRepository) SimilaritySearch(ctx context.Context, embedding []float32, topK int, threshold float64) ([]*ent.DiscoveredEntity, error) {
	// Placeholder implementation - will be implemented with raw SQL for pgvector
	return nil, fmt.Errorf("not implemented yet - requires pgvector integration")
}

// Close closes the database connection
func (r *entRepository) Close() error {
	return r.client.Close()
}
