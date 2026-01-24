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

// TraverseRelationships traverses relationships from an entity with BFS up to specified depth
func (r *entRepository) TraverseRelationships(ctx context.Context, fromID int, relType string, depth int) ([]*ent.DiscoveredEntity, error) {
	if depth <= 0 {
		return nil, nil
	}

	visited := make(map[int]bool)
	var allEntities []*ent.DiscoveredEntity
	currentLevel := []int{fromID}
	visited[fromID] = true

	// Breadth-first search for n-hop traversal
	for currentDepth := 0; currentDepth < depth && len(currentLevel) > 0; currentDepth++ {
		var nextLevel []int

		for _, currentID := range currentLevel {
			// Find relationships from this entity
			var rels []*ent.Relationship
			var err error

			if relType == "" {
				// Get all relationships if no type specified
				rels, err = r.client.Relationship.Query().
					Where(
						relationship.Or(
							relationship.FromIDEQ(currentID),
							relationship.ToIDEQ(currentID),
						),
					).
					All(ctx)
			} else {
				// Get relationships of specific type
				rels, err = r.client.Relationship.Query().
					Where(
						relationship.And(
							relationship.TypeEQ(relType),
							relationship.Or(
								relationship.FromIDEQ(currentID),
								relationship.ToIDEQ(currentID),
							),
						),
					).
					All(ctx)
			}

			if err != nil {
				return nil, fmt.Errorf("failed to query relationships: %w", err)
			}

			// Collect target entities
			for _, rel := range rels {
				var targetID int
				var targetType string

				// Determine the target entity (the one we're traversing to)
				if rel.FromID == currentID {
					targetID = rel.ToID
					targetType = rel.ToType
				} else {
					targetID = rel.FromID
					targetType = rel.FromType
				}

				// Only process discovered entities
				if targetType == "discovered_entity" && !visited[targetID] {
					entity, err := r.FindEntityByID(ctx, targetID)
					if err != nil {
						continue // Skip if entity not found
					}

					allEntities = append(allEntities, entity)
					visited[targetID] = true
					nextLevel = append(nextLevel, targetID)
				}
			}
		}

		currentLevel = nextLevel
	}

	return allEntities, nil
}

// FindShortestPath finds the shortest path between two entities using BFS
func (r *entRepository) FindShortestPath(ctx context.Context, fromID, toID int) ([]*ent.Relationship, error) {
	return r.findShortestPathBFS(ctx, fromID, toID)
}

// SimilaritySearch finds entities similar to the given embedding using pgvector
func (r *entRepository) SimilaritySearch(ctx context.Context, embedding []float32, topK int, threshold float64) ([]*ent.DiscoveredEntity, error) {
	// For POC, we'll use a simpler approach: return empty results
	// Full pgvector integration would require raw SQL with the driver
	// This would be implemented in production with:
	// 1. Get the underlying sql.DB from ent client
	// 2. Execute raw SQL query with pgvector operators
	// 3. Map results back to ent entities

	// Placeholder: return empty results for now
	// Full implementation would look like:
	/*
		db := r.client.Driver().(*sql.DB)
		query := `
			SELECT id, unique_id, type_category, name, properties, embedding, confidence_score, created_at
			FROM discovered_entities
			WHERE embedding IS NOT NULL
			ORDER BY embedding <-> $1
			LIMIT $2
		`
		rows, err := db.QueryContext(ctx, query, embeddingStr, topK)
		...
	*/

	return []*ent.DiscoveredEntity{}, nil
}

// Close closes the database connection
func (r *entRepository) Close() error {
	return r.client.Close()
}
