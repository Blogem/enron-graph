package graph

import (
	"context"

	"github.com/Blogem/enron-graph/ent"
)

// MockRepository is a mock implementation of Repository for testing
type MockRepository struct {
	emails           []*ent.Email
	entities         []*ent.DiscoveredEntity
	relationships    []*ent.Relationship
	entityTypes      []string
	createEmailFunc  func(ctx context.Context, email *EmailInput) (*ent.Email, error)
	createEntityFunc func(ctx context.Context, entity *EntityInput) (*ent.DiscoveredEntity, error)
	createRelFunc    func(ctx context.Context, rel *RelationshipInput) (*ent.Relationship, error)
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		emails:        []*ent.Email{},
		entities:      []*ent.DiscoveredEntity{},
		relationships: []*ent.Relationship{},
		entityTypes:   []string{},
	}
}

func (m *MockRepository) CreateEmail(ctx context.Context, email *EmailInput) (*ent.Email, error) {
	if m.createEmailFunc != nil {
		return m.createEmailFunc(ctx, email)
	}
	e := &ent.Email{ID: len(m.emails) + 1}
	m.emails = append(m.emails, e)
	return e, nil
}

func (m *MockRepository) FindEmailByMessageID(ctx context.Context, messageID string) (*ent.Email, error) {
	return nil, nil
}

func (m *MockRepository) CreateDiscoveredEntity(ctx context.Context, entity *EntityInput) (*ent.DiscoveredEntity, error) {
	if m.createEntityFunc != nil {
		return m.createEntityFunc(ctx, entity)
	}
	e := &ent.DiscoveredEntity{ID: len(m.entities) + 1}
	m.entities = append(m.entities, e)
	return e, nil
}

func (m *MockRepository) FindEntityByID(ctx context.Context, id int, typeHint ...string) (*ent.DiscoveredEntity, error) {
	for _, entity := range m.entities {
		if entity.ID == id {
			return entity, nil
		}
	}
	// Return a default entity to avoid nil pointer errors in tests
	return &ent.DiscoveredEntity{
		ID:              id,
		TypeCategory:    "person",
		Name:            "Test Entity",
		Properties:      map[string]interface{}{},
		ConfidenceScore: 0.95,
	}, nil
}

func (m *MockRepository) FindEntityByUniqueID(ctx context.Context, uniqueID string, typeHint ...string) (*ent.DiscoveredEntity, error) {
	return nil, nil
}

func (m *MockRepository) FindEntitiesByType(ctx context.Context, typeCategory string, typeHint ...string) ([]*ent.DiscoveredEntity, error) {
	return m.entities, nil
}

func (m *MockRepository) GetDistinctEntityTypes(ctx context.Context) ([]string, error) {
	return m.entityTypes, nil
}

func (m *MockRepository) CreateRelationship(ctx context.Context, rel *RelationshipInput) (*ent.Relationship, error) {
	if m.createRelFunc != nil {
		return m.createRelFunc(ctx, rel)
	}
	r := &ent.Relationship{ID: len(m.relationships) + 1}
	m.relationships = append(m.relationships, r)
	return r, nil
}

func (m *MockRepository) FindRelationshipsByEntity(ctx context.Context, entityType string, entityID int) ([]*ent.Relationship, error) {
	return m.relationships, nil
}

func (m *MockRepository) GetDistinctRelationshipTypes(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (m *MockRepository) TraverseRelationships(ctx context.Context, fromID int, relType string, depth int) ([]*ent.DiscoveredEntity, error) {
	return nil, nil
}

func (m *MockRepository) FindShortestPath(ctx context.Context, fromID, toID int) ([]*ent.Relationship, error) {
	return nil, nil
}

func (m *MockRepository) SimilaritySearch(ctx context.Context, embedding []float32, topK int, threshold float64) ([]*ent.DiscoveredEntity, error) {
	return nil, nil
}

func (m *MockRepository) Close() error {
	return nil
}

func (m *MockRepository) GetClient() *ent.Client {
	return nil // Mock doesn't have a real client
}
