package graph

import (
	"context"
	"time"

	"github.com/Blogem/enron-graph/ent"
)

// Repository defines the interface for graph operations
type Repository interface {
	// Email operations
	CreateEmail(ctx context.Context, email *EmailInput) (*ent.Email, error)
	FindEmailByMessageID(ctx context.Context, messageID string) (*ent.Email, error)

	// Entity operations
	CreateDiscoveredEntity(ctx context.Context, entity *EntityInput) (*ent.DiscoveredEntity, error)

	// FindEntityByID finds an entity by its database ID.
	// Optional typeHint parameter enables direct table lookup when entity type is known.
	// Fallback strategy: type hint → relationships inference → generic discovery.
	FindEntityByID(ctx context.Context, id int, typeHint ...string) (*ent.DiscoveredEntity, error)

	// FindEntityByUniqueID finds an entity by its unique identifier.
	// Optional typeHint parameter enables O(1) direct table lookup when entity type is known.
	// Fallback strategy: type hint → discovered_entities → relationships inference → parallel search.
	FindEntityByUniqueID(ctx context.Context, uniqueID string, typeHint ...string) (*ent.DiscoveredEntity, error)

	// FindEntitiesByType returns all entities of a given type.
	// Optional typeHint parameter enables querying promoted tables directly.
	// If type is promoted, queries promoted table; otherwise queries discovered_entities.
	FindEntitiesByType(ctx context.Context, typeCategory string, typeHint ...string) ([]*ent.DiscoveredEntity, error)

	GetDistinctEntityTypes(ctx context.Context) ([]string, error)

	// Relationship operations
	CreateRelationship(ctx context.Context, rel *RelationshipInput) (*ent.Relationship, error)
	FindRelationshipsByEntity(ctx context.Context, entityType string, entityID int) ([]*ent.Relationship, error)
	GetDistinctRelationshipTypes(ctx context.Context) ([]string, error)

	// Graph traversal
	TraverseRelationships(ctx context.Context, fromID int, relType string, depth int) ([]*ent.DiscoveredEntity, error)
	FindShortestPath(ctx context.Context, fromID, toID int) ([]*ent.Relationship, error)

	// Vector search
	SimilaritySearch(ctx context.Context, embedding []float32, topK int, threshold float64) ([]*ent.DiscoveredEntity, error)

	// Get the underlying Ent client (for registry creators)
	GetClient() *ent.Client

	// Close the database connection
	Close() error
}

// EmailInput represents input data for creating an email
type EmailInput struct {
	MessageID string
	From      string
	To        []string
	CC        []string
	BCC       []string
	Subject   string
	Date      time.Time
	Body      string
	FilePath  string
}

// EntityInput represents input data for creating a discovered entity
type EntityInput struct {
	UniqueID        string
	TypeCategory    string
	Name            string
	Properties      map[string]interface{}
	Embedding       []float32
	ConfidenceScore float64
}

// RelationshipInput represents input data for creating a relationship
type RelationshipInput struct {
	Type            string
	FromType        string
	FromID          int
	ToType          string
	ToID            int
	Timestamp       time.Time
	ConfidenceScore float64
	Properties      map[string]interface{}
}
