package graph

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/email"
	"github.com/Blogem/enron-graph/ent/relationship"

	_ "github.com/lib/pq"
)

// entRepository implements the Repository interface using ent
type entRepository struct {
	client *ent.Client
	db     *sql.DB
	logger *slog.Logger
}

// NewRepository creates a new ent-based repository
func NewRepository(client *ent.Client, logger *slog.Logger) Repository {
	return &entRepository{
		client: client,
		db:     nil, // No SQL DB for raw queries
		logger: logger,
	}
}

// NewRepositoryWithDB creates a new ent-based repository with a direct SQL connection
// The SQL connection is needed for raw pgvector queries
func NewRepositoryWithDB(client *ent.Client, db *sql.DB, logger *slog.Logger) Repository {
	return &entRepository{
		client: client,
		db:     db,
		logger: logger,
	}
}

// GetClient returns the underlying Ent client
func (r *entRepository) GetClient() *ent.Client {
	return r.client
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
// If typeCategory is empty, returns all entities
func (r *entRepository) FindEntitiesByType(ctx context.Context, typeCategory string) ([]*ent.DiscoveredEntity, error) {
	query := r.client.DiscoveredEntity.Query()
	if typeCategory != "" {
		query = query.Where(discoveredentity.TypeCategoryEQ(typeCategory))
	}
	return query.All(ctx)
}

// GetDistinctEntityTypes returns all unique entity type categories
func (r *entRepository) GetDistinctEntityTypes(ctx context.Context) ([]string, error) {
	discoveredTypes, err := r.client.DiscoveredEntity.Query().
		GroupBy(discoveredentity.FieldTypeCategory).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct types: %w", err)
	}

	// If no SQL DB available, return only discovered types
	if r.db == nil {
		r.logger.Info("No SQL DB connection, returning only discovered entity types")
		return discoveredTypes, nil
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_type = 'BASE TABLE'
		AND table_name NOT IN ('relationships', 'discovered_entities', 'schema_promotions')
		ORDER BY table_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table names: %w", err)
	}
	defer rows.Close()

	var promotedTypes []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		promotedTypes = append(promotedTypes, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table names: %w", err)
	}

	return append(discoveredTypes, promotedTypes...), nil
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
	// Support both schema types: specific types (person, organization, etc.) and generic "discovered_entity"
	return r.client.Relationship.Query().
		Where(
			relationship.Or(
				// Match on specific entity type (test DB schema)
				relationship.And(
					relationship.FromTypeEQ(entityType),
					relationship.FromIDEQ(entityID),
				),
				relationship.And(
					relationship.ToTypeEQ(entityType),
					relationship.ToIDEQ(entityID),
				),
				// Match on discovered_entity (production DB schema)
				relationship.And(
					relationship.FromTypeEQ("discovered_entity"),
					relationship.FromIDEQ(entityID),
				),
				relationship.And(
					relationship.ToTypeEQ("discovered_entity"),
					relationship.ToIDEQ(entityID),
				),
			),
		).
		All(ctx)
}

// GetDistinctRelationshipTypes returns all unique relationship types
func (r *entRepository) GetDistinctRelationshipTypes(ctx context.Context) ([]string, error) {
	relationshipTypes, err := r.client.Relationship.Query().
		GroupBy(relationship.FieldType).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct relationship types: %w", err)
	}
	return relationshipTypes, nil
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
	// Convert embedding to JSON array string for pgvector
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Check if we have access to the underlying database
	if r.db == nil {
		return nil, fmt.Errorf("database connection not available for raw SQL queries")
	}

	// Build the query with pgvector distance operator
	// Convert JSONB to vector via text casting
	query := `
		SELECT id, unique_id, type_category, name, properties, embedding, confidence_score, created_at
		FROM discovered_entities
		WHERE embedding IS NOT NULL
		ORDER BY embedding::text::vector <-> $1::vector
		LIMIT $2
	`

	// Execute the raw SQL query
	rows, err := r.db.QueryContext(ctx, query, string(embeddingJSON), topK)
	if err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}
	defer rows.Close()

	// Parse results into ent entities
	entities := make([]*ent.DiscoveredEntity, 0)
	for rows.Next() {
		var (
			id              int
			uniqueID        string
			typeCategory    string
			name            string
			propertiesJSON  []byte
			embeddingJSON   []byte
			confidenceScore float64
			createdAt       sql.NullTime
		)

		if err := rows.Scan(&id, &uniqueID, &typeCategory, &name, &propertiesJSON, &embeddingJSON, &confidenceScore, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse properties JSON
		var properties map[string]interface{}
		if len(propertiesJSON) > 0 {
			if err := json.Unmarshal(propertiesJSON, &properties); err != nil {
				return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
			}
		}

		// Create ent entity (simplified, doesn't include edges)
		entity := &ent.DiscoveredEntity{
			ID:              id,
			UniqueID:        uniqueID,
			TypeCategory:    typeCategory,
			Name:            name,
			Properties:      properties,
			ConfidenceScore: confidenceScore,
		}

		if createdAt.Valid {
			entity.CreatedAt = createdAt.Time
		}

		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Apply threshold filter if specified
	// For cosine distance: 0 = identical, 2 = opposite
	// threshold of 0.7 similarity means distance <= 0.3
	if threshold > 0 {
		maxDistance := 1.0 - threshold
		filteredQuery := fmt.Sprintf(`
			SELECT id, unique_id, type_category, name, properties, embedding, confidence_score, created_at,
			       embedding::text::vector <-> $1::vector as distance
			FROM discovered_entities
			WHERE embedding IS NOT NULL
			  AND embedding::text::vector <-> $1::vector <= %f
			ORDER BY distance
			LIMIT $2
		`, maxDistance)

		rows2, err := r.db.QueryContext(ctx, filteredQuery, string(embeddingJSON), topK)
		if err != nil {
			return nil, fmt.Errorf("failed to execute filtered similarity search: %w", err)
		}
		defer rows2.Close()

		filtered := make([]*ent.DiscoveredEntity, 0)
		for rows2.Next() {
			var (
				id              int
				uniqueID        string
				typeCategory    string
				name            string
				propertiesJSON  []byte
				embeddingJSON   []byte
				confidenceScore float64
				createdAt       sql.NullTime
				distance        float64
			)

			if err := rows2.Scan(&id, &uniqueID, &typeCategory, &name, &propertiesJSON, &embeddingJSON, &confidenceScore, &createdAt, &distance); err != nil {
				return nil, fmt.Errorf("failed to scan filtered row: %w", err)
			}

			var properties map[string]interface{}
			if len(propertiesJSON) > 0 {
				if err := json.Unmarshal(propertiesJSON, &properties); err != nil {
					return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
				}
			}

			entity := &ent.DiscoveredEntity{
				ID:              id,
				UniqueID:        uniqueID,
				TypeCategory:    typeCategory,
				Name:            name,
				Properties:      properties,
				ConfidenceScore: confidenceScore,
			}

			if createdAt.Valid {
				entity.CreatedAt = createdAt.Time
			}

			filtered = append(filtered, entity)
		}

		if err := rows2.Err(); err != nil {
			return nil, fmt.Errorf("error iterating filtered rows: %w", err)
		}

		return filtered, nil
	}

	return entities, nil
}

// Close closes the database connection
func (r *entRepository) Close() error {
	return r.client.Close()
}
