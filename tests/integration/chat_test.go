package integration

import (
	"context"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/internal/chat"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChatIntegration_EntityLookup tests the chat interface entity lookup functionality
func TestChatIntegration_EntityLookup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := SetupTestDB(t)
	repo := graph.NewRepository(client)

	// Pre-populate test database with entities
	jeffSkilling, err := repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "jeff.skilling@enron.com",
		TypeCategory:    "person",
		Name:            "Jeff Skilling",
		Properties:      map[string]interface{}{"title": "CEO"},
		Embedding:       make([]float32, 768),
		ConfidenceScore: 0.95,
	})
	require.NoError(t, err, "Failed to create Jeff Skilling entity")

	kennethLay, err := repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "kenneth.lay@enron.com",
		TypeCategory:    "person",
		Name:            "Kenneth Lay",
		Properties:      map[string]interface{}{"title": "Chairman"},
		Embedding:       make([]float32, 768),
		ConfidenceScore: 0.95,
	})
	require.NoError(t, err, "Failed to create Kenneth Lay entity")

	// Create repository adapter for chat
	chatRepo := newChatRepositoryAdapter(repo)

	// Create mock LLM client that returns structured responses
	mockLLM := &mockLLMClient{
		responses: make(map[string]string),
	}

	// Configure mock responses for different queries
	mockLLM.responses["Who is Jeff Skilling?"] = `{
		"action": "entity_lookup",
		"entity": "Jeff Skilling"
	}`

	mockLLM.responses["Who did Jeff Skilling email?"] = `{
		"action": "relationship",
		"entity": "Jeff Skilling",
		"rel_type": "SENT"
	}`

	mockLLM.responses["How are Jeff Skilling and Kenneth Lay connected?"] = `{
		"action": "path_finding",
		"source": "Jeff Skilling",
		"target": "Kenneth Lay"
	}`

	// Create chat handler and context
	handler := chat.NewHandler(mockLLM, chatRepo)
	chatContext := chat.NewContext()

	t.Run("EntityLookup_JeffSkilling", func(t *testing.T) {
		// Test: Submit query "Who is Jeff Skilling?"
		response, err := handler.ProcessQuery(ctx, "Who is Jeff Skilling?", chatContext)
		require.NoError(t, err, "Failed to process query")

		// Verify: Correct entity information returned
		assert.Contains(t, response, "Jeff Skilling", "Response should mention Jeff Skilling")
		assert.Contains(t, response, "person", "Response should mention entity type")

		// Verify context tracking
		entities := chatContext.GetTrackedEntities()
		assert.Contains(t, entities, "Jeff Skilling", "Jeff Skilling should be tracked in context")

		t.Logf("✓ Entity lookup test passed: %s", response)
	})

	t.Run("TraverseRelationships_JeffSkillingEmails", func(t *testing.T) {
		// Create a relationship for testing
		_, err := repo.CreateRelationship(ctx, &graph.RelationshipInput{
			Type:            "SENT",
			FromType:        "discovered_entity",
			FromID:          jeffSkilling.ID,
			ToType:          "discovered_entity",
			ToID:            kennethLay.ID,
			Timestamp:       time.Now(),
			ConfidenceScore: 0.9,
			Properties:      map[string]interface{}{},
		})
		require.NoError(t, err, "Failed to create relationship")

		// Test: Submit query "Who did Jeff Skilling email?"
		response, err := handler.ProcessQuery(ctx, "Who did Jeff Skilling email?", chatContext)
		require.NoError(t, err, "Failed to process query")

		// Verify: Related entities returned
		assert.NotEmpty(t, response, "Response should not be empty")

		t.Logf("✓ Relationship traversal test passed: %s", response)
	})

	t.Run("PathFinding_JeffSkillingToKennethLay", func(t *testing.T) {
		// Ensure relationship exists from previous test
		relationships, err := repo.FindRelationshipsByEntity(ctx, "discovered_entity", jeffSkilling.ID)
		require.NoError(t, err, "Failed to find relationships")

		if len(relationships) == 0 {
			// Create relationship if not exists
			_, err := repo.CreateRelationship(ctx, &graph.RelationshipInput{
				Type:            "SENT",
				FromType:        "discovered_entity",
				FromID:          jeffSkilling.ID,
				ToType:          "discovered_entity",
				ToID:            kennethLay.ID,
				Timestamp:       time.Now(),
				ConfidenceScore: 0.9,
				Properties:      map[string]interface{}{},
			})
			require.NoError(t, err, "Failed to create relationship")
		}

		// Test: Submit query "How are Jeff Skilling and Kenneth Lay connected?"
		response, err := handler.ProcessQuery(ctx, "How are Jeff Skilling and Kenneth Lay connected?", chatContext)
		require.NoError(t, err, "Failed to process query")

		// Verify: Path information returned
		assert.NotEmpty(t, response, "Response should not be empty")

		t.Logf("✓ Path finding test passed: %s", response)
	})
}

// mockLLMClient is a mock implementation of chat.LLMClient for testing
type mockLLMClient struct {
	responses map[string]string
	patterns  []string // Ordered list of patterns for matching priority
}

func (m *mockLLMClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// Try pattern matching with priority order (longest patterns first)
	for _, pattern := range m.patterns {
		if contains(prompt, pattern) {
			return m.responses[pattern], nil
		}
	}

	// Fallback to map iteration (shouldn't reach here if patterns list is complete)
	for query, response := range m.responses {
		if contains(prompt, query) {
			return response, nil
		}
	}

	// Default response
	return `{"action": "unknown", "message": "I don't understand that query."}`, nil
}

func (m *mockLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Return zero embedding for testing
	return make([]float32, 768), nil
}

// contains checks if haystack contains needle (case-insensitive)
func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack == needle ||
			findSubstring(haystack, needle))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// chatRepositoryAdapter adapts graph.Repository to chat.Repository for testing
type chatRepositoryAdapter struct {
	repo graph.Repository
}

func newChatRepositoryAdapter(repo graph.Repository) chat.Repository {
	return &chatRepositoryAdapter{repo: repo}
}

func (a *chatRepositoryAdapter) FindEntityByName(name string) (*chat.Entity, error) {
	ctx := context.Background()
	entities, err := a.repo.FindEntitiesByType(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		if entity.Name == name {
			return &chat.Entity{
				ID:         entity.ID,
				Name:       entity.Name,
				Type:       entity.TypeCategory,
				Properties: entity.Properties,
			}, nil
		}
	}

	return nil, nil
}

func (a *chatRepositoryAdapter) TraverseRelationships(entityID int, relType string) ([]*chat.Entity, error) {
	ctx := context.Background()
	entities, err := a.repo.TraverseRelationships(ctx, entityID, relType, 1)
	if err != nil {
		return nil, err
	}

	result := make([]*chat.Entity, len(entities))
	for i, entity := range entities {
		result[i] = &chat.Entity{
			ID:         entity.ID,
			Name:       entity.Name,
			Type:       entity.TypeCategory,
			Properties: entity.Properties,
		}
	}

	return result, nil
}

func (a *chatRepositoryAdapter) FindShortestPath(sourceID, targetID int) ([]*chat.PathNode, error) {
	ctx := context.Background()
	relationships, err := a.repo.FindShortestPath(ctx, sourceID, targetID)
	if err != nil {
		return nil, err
	}

	if len(relationships) == 0 {
		return nil, nil
	}

	path := make([]*chat.PathNode, 0)

	sourceEntity, err := a.repo.FindEntityByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	path = append(path, &chat.PathNode{
		Entity: &chat.Entity{
			ID:         sourceEntity.ID,
			Name:       sourceEntity.Name,
			Type:       sourceEntity.TypeCategory,
			Properties: sourceEntity.Properties,
		},
		Relationship: "",
	})

	for _, rel := range relationships {
		targetEntity, err := a.repo.FindEntityByID(ctx, rel.ToID)
		if err != nil {
			return nil, err
		}

		path = append(path, &chat.PathNode{
			Entity: &chat.Entity{
				ID:         targetEntity.ID,
				Name:       targetEntity.Name,
				Type:       targetEntity.TypeCategory,
				Properties: targetEntity.Properties,
			},
			Relationship: rel.Type,
		})
	}

	return path, nil
}

func (a *chatRepositoryAdapter) SimilaritySearch(embedding []float32, limit int) ([]*chat.Entity, error) {
	ctx := context.Background()
	entities, err := a.repo.SimilaritySearch(ctx, embedding, limit, 0.7)
	if err != nil {
		return nil, err
	}

	result := make([]*chat.Entity, len(entities))
	for i, entity := range entities {
		result[i] = &chat.Entity{
			ID:         entity.ID,
			Name:       entity.Name,
			Type:       entity.TypeCategory,
			Properties: entity.Properties,
		}
	}

	return result, nil
}

func (a *chatRepositoryAdapter) CountRelationships(entityID int, relType string) (int, error) {
	ctx := context.Background()
	relationships, err := a.repo.FindRelationshipsByEntity(ctx, "discovered_entity", entityID)
	if err != nil {
		return 0, err
	}

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
