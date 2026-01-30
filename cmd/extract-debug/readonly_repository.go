package main

import (
	"context"
	"log/slog"
	"math/rand"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
)

// ReadOnlyRepository wraps a repository and prevents writes while capturing what would have been created
type ReadOnlyRepository struct {
	base                  graph.Repository
	logger                *slog.Logger
	capturedEntities      []*ent.DiscoveredEntity
	capturedRelationships []*graph.RelationshipInput
}

// NewReadOnlyRepository creates a new read-only repository wrapper
func NewReadOnlyRepository(base graph.Repository, logger *slog.Logger) *ReadOnlyRepository {
	return &ReadOnlyRepository{
		base:   base,
		logger: logger,
	}
}

// CreateEmail is blocked (read-only)
func (r *ReadOnlyRepository) CreateEmail(ctx context.Context, email *graph.EmailInput) (*ent.Email, error) {
	r.logger.Debug("Blocked CreateEmail call (read-only mode)", "message_id", email.MessageID)
	return &ent.Email{}, nil
}

// FindEmailByMessageID delegates to base repository (read operation)
func (r *ReadOnlyRepository) FindEmailByMessageID(ctx context.Context, messageID string) (*ent.Email, error) {
	return r.base.FindEmailByMessageID(ctx, messageID)
}

// CreateDiscoveredEntity captures the entity but doesn't persist it
func (r *ReadOnlyRepository) CreateDiscoveredEntity(ctx context.Context, entity *graph.EntityInput) (*ent.DiscoveredEntity, error) {
	r.logger.Debug("Captured entity (not persisted)", "type", entity.TypeCategory, "name", entity.Name)

	capturedEntity := &ent.DiscoveredEntity{
		ID:              rand.Intn(50), // Mock ID
		UniqueID:        entity.UniqueID,
		TypeCategory:    entity.TypeCategory,
		Name:            entity.Name,
		Properties:      entity.Properties,
		ConfidenceScore: entity.ConfidenceScore,
	}

	// Capture the entity input for later display
	r.capturedEntities = append(r.capturedEntities, capturedEntity)

	// Return a mock entity with the input data
	return capturedEntity, nil
}

// FindEntityByID delegates to base repository (read operation)
func (r *ReadOnlyRepository) FindEntityByID(ctx context.Context, id int) (*ent.DiscoveredEntity, error) {
	return r.base.FindEntityByID(ctx, id)
}

// FindEntityByUniqueID delegates to base repository (read operation)
func (r *ReadOnlyRepository) FindEntityByUniqueID(ctx context.Context, uniqueID string) (*ent.DiscoveredEntity, error) {
	return r.base.FindEntityByUniqueID(ctx, uniqueID)
}

// FindEntitiesByType delegates to base repository (read operation)
func (r *ReadOnlyRepository) FindEntitiesByType(ctx context.Context, typeCategory string) ([]*ent.DiscoveredEntity, error) {
	return r.base.FindEntitiesByType(ctx, typeCategory)
}

// GetDistinctEntityTypes delegates to base repository (read operation)
func (r *ReadOnlyRepository) GetDistinctEntityTypes(ctx context.Context) ([]string, error) {
	return r.base.GetDistinctEntityTypes(ctx)
}

// CreateRelationship captures the relationship but doesn't persist it
func (r *ReadOnlyRepository) CreateRelationship(ctx context.Context, rel *graph.RelationshipInput) (*ent.Relationship, error) {
	r.logger.Debug("Captured relationship (not persisted)", "type", rel.Type, "from", rel.FromType, "to", rel.ToType)

	// Capture the relationship input for later display
	r.capturedRelationships = append(r.capturedRelationships, rel)

	// Return a mock relationship with the input data
	return &ent.Relationship{
		ID:              rand.Intn(50), // Mock ID
		Type:            rel.Type,
		FromType:        rel.FromType,
		FromID:          rel.FromID,
		ToType:          rel.ToType,
		ToID:            rel.ToID,
		Timestamp:       rel.Timestamp,
		ConfidenceScore: rel.ConfidenceScore,
		Properties:      rel.Properties,
	}, nil
}

// FindRelationshipsByEntity delegates to base repository (read operation)
func (r *ReadOnlyRepository) FindRelationshipsByEntity(ctx context.Context, entityType string, entityID int) ([]*ent.Relationship, error) {
	return r.base.FindRelationshipsByEntity(ctx, entityType, entityID)
}

// GetDistinctRelationshipTypes delegates to base repository (read operation)
func (r *ReadOnlyRepository) GetDistinctRelationshipTypes(ctx context.Context) ([]string, error) {
	return r.base.GetDistinctRelationshipTypes(ctx)
}

// TraverseRelationships delegates to base repository (read operation)
func (r *ReadOnlyRepository) TraverseRelationships(ctx context.Context, fromID int, relType string, depth int) ([]*ent.DiscoveredEntity, error) {
	return r.base.TraverseRelationships(ctx, fromID, relType, depth)
}

// FindShortestPath delegates to base repository (read operation)
func (r *ReadOnlyRepository) FindShortestPath(ctx context.Context, fromID, toID int) ([]*ent.Relationship, error) {
	return r.base.FindShortestPath(ctx, fromID, toID)
}

// SimilaritySearch delegates to base repository (read operation)
func (r *ReadOnlyRepository) SimilaritySearch(ctx context.Context, embedding []float32, topK int, threshold float64) ([]*ent.DiscoveredEntity, error) {
	return r.base.SimilaritySearch(ctx, embedding, topK, threshold)
}

// Close delegates to base repository
func (r *ReadOnlyRepository) Close() error {
	return r.base.Close()
}

// GetClient delegates to base repository
func (r *ReadOnlyRepository) GetClient() *ent.Client {
	return r.base.GetClient()
}
