package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/internal/api"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIIntegration tests T068: API integration test
// - Setup: Pre-populate test database with entities and relationships
// - Test: Query entities via API
// - Verify: Correct entities returned
// - Test: Find shortest path between known entities
// - Verify: Correct path returned
// - Test: Semantic search for concept
// - Verify: Similar entities returned
func TestAPIIntegration(t *testing.T) {
	// Setup: Create test database with clean schema
	client := SetupTestDB(t)

	ctx := context.Background()

	// Create repository
	repo := graph.NewRepository(client)

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Initialize LLM client (using Ollama)
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Pre-populate test database with entities and relationships
	setupTestEntitiesAndRelationships(t, ctx, client, repo, llmClient)

	// Create API handler and router
	handler := api.NewHandlerWithLLM(repo, llmClient)
	router := setupAPIRouter(handler)

	// Test: Query entities via API
	t.Run("QueryEntities", func(t *testing.T) {
		// Test GET /entities?type=person
		req := httptest.NewRequest("GET", "/api/v1/entities?type=person", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Parse response
		var response api.SearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Failed to parse response")

		// Verify: Correct entities returned
		assert.Greater(t, len(response.Entities), 0, "Expected at least one person entity")
		assert.Greater(t, response.Total, 0, "Expected total count > 0")

		// Verify all returned entities are persons
		for _, entity := range response.Entities {
			assert.Equal(t, "person", entity.TypeCategory, "Expected all entities to be of type person")
		}

		t.Logf("Found %d person entities", len(response.Entities))
	})

	// Test: Query specific entity by name
	t.Run("QueryEntityByName", func(t *testing.T) {
		// Test GET /entities?name=jeff
		req := httptest.NewRequest("GET", "/api/v1/entities?type=person&name=jeff", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Parse response
		var response api.SearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Failed to parse response")

		// Verify: Entities with "jeff" in name returned
		assert.Greater(t, len(response.Entities), 0, "Expected at least one entity with name containing 'jeff'")

		t.Logf("Found %d entities matching 'jeff'", len(response.Entities))
	})

	// Test: Get entity by ID
	t.Run("GetEntityByID", func(t *testing.T) {
		// First, get an entity ID from search
		entities, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.TypeCategoryEQ("person")).
			Limit(1).
			All(ctx)
		require.NoError(t, err, "Failed to query entity")
		require.Greater(t, len(entities), 0, "No entities found for test")

		entityID := entities[0].ID

		// Test GET /entities/:id
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d", entityID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Parse response
		var response api.EntityResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Failed to parse response")

		// Verify: Correct entity returned
		assert.Equal(t, entityID, response.ID, "Expected entity ID to match")
		assert.NotEmpty(t, response.Name, "Expected entity to have a name")

		t.Logf("Retrieved entity: %s (ID: %d)", response.Name, response.ID)
	})

	// Test: Get entity relationships
	t.Run("GetEntityRelationships", func(t *testing.T) {
		// Get an entity that has relationships
		entities, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.TypeCategoryEQ("person")).
			Limit(1).
			All(ctx)
		require.NoError(t, err, "Failed to query entity")
		require.Greater(t, len(entities), 0, "No entities found for test")

		entityID := entities[0].ID

		// Test GET /entities/:id/relationships
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/relationships", entityID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Parse response
		var response api.RelationshipsResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Failed to parse response")

		// Verify: Relationships returned
		assert.Equal(t, entityID, response.EntityID, "Expected entity ID to match")
		// Note: May have 0 relationships if entity was just created
		t.Logf("Entity %d has %d relationships", entityID, len(response.Relationships))
	})

	// Test: Find shortest path between known entities
	t.Run("FindShortestPath", func(t *testing.T) {
		// Get two entities that potentially have a path between them
		entities, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.TypeCategoryEQ("person")).
			Limit(2).
			All(ctx)
		require.NoError(t, err, "Failed to query entities")
		if len(entities) < 2 {
			t.Skip("Not enough entities to test path finding")
		}

		sourceID := entities[0].ID
		targetID := entities[1].ID

		// Test POST /entities/path
		pathReq := api.PathRequest{
			SourceID: sourceID,
			TargetID: targetID,
			MaxDepth: 5,
		}
		reqBody, _ := json.Marshal(pathReq)
		req := httptest.NewRequest("POST", "/api/v1/entities/path", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Note: May return 404 if no path exists, which is acceptable
		if w.Code == http.StatusOK {
			// Parse response
			var response api.PathResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Failed to parse response")

			// Verify: Correct path returned
			assert.Equal(t, sourceID, response.SourceID, "Expected source ID to match")
			assert.Equal(t, targetID, response.TargetID, "Expected target ID to match")
			assert.Greater(t, len(response.Path), 0, "Expected path to have elements")

			t.Logf("Found path of length %d between entities %d and %d", response.PathLength, sourceID, targetID)
		} else if w.Code == http.StatusNotFound {
			t.Logf("No path found between entities %d and %d (acceptable)", sourceID, targetID)
		} else {
			t.Fatalf("Unexpected status code: %d", w.Code)
		}
	})

	// Test: Semantic search for concept
	t.Run("SemanticSearch", func(t *testing.T) {
		// Test POST /entities/search
		searchReq := api.SearchRequest{
			Query:         "energy trading",
			Limit:         10,
			MinSimilarity: 0.5,
		}
		reqBody, _ := json.Marshal(searchReq)
		req := httptest.NewRequest("POST", "/api/v1/entities/search", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Semantic search requires Ollama to be running
		if w.Code == http.StatusOK {
			// Parse response
			var response api.SemanticSearchResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Failed to parse response")

			// Verify: Similar entities returned
			assert.Equal(t, "energy trading", response.Query, "Expected query to match")
			t.Logf("Found %d similar entities for query 'energy trading'", len(response.Results))

			// Check similarity scores
			for _, result := range response.Results {
				assert.GreaterOrEqual(t, result.Similarity, 0.5, "Expected similarity >= min threshold")
				assert.LessOrEqual(t, result.Similarity, 1.0, "Expected similarity <= 1.0")
			}
		} else {
			// Ollama might not be available in test environment
			t.Logf("Semantic search returned status %d (Ollama may not be available)", w.Code)
		}
	})

	// Test: Get entity neighbors (traversal)
	t.Run("GetEntityNeighbors", func(t *testing.T) {
		// Get an entity that might have neighbors
		entities, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.TypeCategoryEQ("person")).
			Limit(1).
			All(ctx)
		require.NoError(t, err, "Failed to query entity")
		require.Greater(t, len(entities), 0, "No entities found for test")

		entityID := entities[0].ID

		// Test GET /entities/:id/neighbors?depth=1
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/neighbors?depth=1", entityID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Parse response
		var response api.NeighborsResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Failed to parse response")

		// Verify: Neighbors response structure
		assert.Equal(t, entityID, response.SourceEntityID, "Expected source entity ID to match")
		assert.Equal(t, 1, response.Depth, "Expected depth to be 1")
		t.Logf("Entity %d has %d neighbors at depth 1", entityID, len(response.Neighbors))
	})
}

// setupTestEntitiesAndRelationships creates test entities and relationships in the database
func setupTestEntitiesAndRelationships(t *testing.T, ctx context.Context, client *ent.Client, repo graph.Repository, llmClient llm.Client) {
	t.Helper()

	// Create sample emails
	date1, _ := time.Parse(time.RFC3339, "2001-05-15T10:00:00Z")
	email1, err := repo.CreateEmail(ctx, &graph.EmailInput{
		MessageID: "msg-test-1@enron.com",
		From:      "jeff.skilling@enron.com",
		To:        []string{"ken.lay@enron.com", "andrew.fastow@enron.com"},
		CC:        []string{"rebecca.mark@enron.com"},
		Subject:   "Energy Trading Discussion",
		Body:      "Let's discuss the energy trading strategy for Q2. Enron Corporation is positioned well in the California market.",
		Date:      date1,
	})
	require.NoError(t, err, "Failed to create test email 1")

	date2, _ := time.Parse(time.RFC3339, "2001-06-20T14:30:00Z")
	email2, err := repo.CreateEmail(ctx, &graph.EmailInput{
		MessageID: "msg-test-2@enron.com",
		From:      "ken.lay@enron.com",
		To:        []string{"jeff.skilling@enron.com"},
		Subject:   "Re: Energy Trading Discussion",
		Body:      "Agreed. The California market presents significant opportunities for Enron.",
		Date:      date2,
	})
	require.NoError(t, err, "Failed to create test email 2")

	date3, _ := time.Parse(time.RFC3339, "2001-07-10T09:15:00Z")
	email3, err := repo.CreateEmail(ctx, &graph.EmailInput{
		MessageID: "msg-test-3@enron.com",
		From:      "andrew.fastow@enron.com",
		To:        []string{"jeff.skilling@enron.com", "ken.lay@enron.com"},
		Subject:   "Financial Structures",
		Body:      "Reviewing the special purpose entities for off-balance sheet transactions.",
		Date:      date3,
	})
	require.NoError(t, err, "Failed to create test email 3")

	// Create person entities with embeddings
	embedding1 := generateMockEmbedding(t, "Jeff Skilling CEO")
	_, err = repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "jeff.skilling@enron.com",
		TypeCategory:    "person",
		Name:            "Jeff Skilling",
		Properties:      map[string]interface{}{"title": "CEO", "email": "jeff.skilling@enron.com"},
		Embedding:       embedding1,
		ConfidenceScore: 0.98,
	})
	require.NoError(t, err, "Failed to create entity: Jeff Skilling")

	embedding2 := generateMockEmbedding(t, "Ken Lay Chairman")
	_, err = repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "ken.lay@enron.com",
		TypeCategory:    "person",
		Name:            "Ken Lay",
		Properties:      map[string]interface{}{"title": "Chairman", "email": "ken.lay@enron.com"},
		Embedding:       embedding2,
		ConfidenceScore: 0.97,
	})
	require.NoError(t, err, "Failed to create entity: Ken Lay")

	embedding3 := generateMockEmbedding(t, "Andrew Fastow CFO")
	_, err = repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "andrew.fastow@enron.com",
		TypeCategory:    "person",
		Name:            "Andrew Fastow",
		Properties:      map[string]interface{}{"title": "CFO", "email": "andrew.fastow@enron.com"},
		Embedding:       embedding3,
		ConfidenceScore: 0.96,
	})
	require.NoError(t, err, "Failed to create entity: Andrew Fastow")

	embedding4 := generateMockEmbedding(t, "Rebecca Mark Executive")
	_, err = repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "rebecca.mark@enron.com",
		TypeCategory:    "person",
		Name:            "Rebecca Mark",
		Properties:      map[string]interface{}{"email": "rebecca.mark@enron.com"},
		Embedding:       embedding4,
		ConfidenceScore: 0.95,
	})
	require.NoError(t, err, "Failed to create entity: Rebecca Mark")

	// Create organization entities
	embedding5 := generateMockEmbedding(t, "Enron Corporation energy company")
	enronEntity, err := repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "enron-corporation",
		TypeCategory:    "organization",
		Name:            "Enron Corporation",
		Properties:      map[string]interface{}{"industry": "Energy"},
		Embedding:       embedding5,
		ConfidenceScore: 0.99,
	})
	require.NoError(t, err, "Failed to create entity: Enron Corporation")

	// Create concept entities
	embedding6 := generateMockEmbedding(t, "energy trading commodities market")
	_, err = repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "concept-energy-trading",
		TypeCategory:    "concept",
		Name:            "energy trading",
		Properties:      map[string]interface{}{"category": "business operation"},
		Embedding:       embedding6,
		ConfidenceScore: 0.85,
	})
	require.NoError(t, err, "Failed to create entity: energy trading concept")

	embedding7 := generateMockEmbedding(t, "California market state energy")
	_, err = repo.CreateDiscoveredEntity(ctx, &graph.EntityInput{
		UniqueID:        "concept-california-market",
		TypeCategory:    "concept",
		Name:            "California market",
		Properties:      map[string]interface{}{"category": "geographic market"},
		Embedding:       embedding7,
		ConfidenceScore: 0.82,
	})
	require.NoError(t, err, "Failed to create entity: California market concept")

	// Create relationships
	// SENT relationships (person -> email)
	jeff, _ := client.DiscoveredEntity.Query().Where(discoveredentity.UniqueIDEQ("jeff.skilling@enron.com")).First(ctx)
	ken, _ := client.DiscoveredEntity.Query().Where(discoveredentity.UniqueIDEQ("ken.lay@enron.com")).First(ctx)
	andrew, _ := client.DiscoveredEntity.Query().Where(discoveredentity.UniqueIDEQ("andrew.fastow@enron.com")).First(ctx)

	_, err = repo.CreateRelationship(ctx, &graph.RelationshipInput{
		Type:            "SENT",
		FromType:        "discovered_entity",
		FromID:          jeff.ID,
		ToType:          "email",
		ToID:            email1.ID,
		ConfidenceScore: 1.0,
		Timestamp:       date1,
	})
	require.NoError(t, err, "Failed to create SENT relationship")

	_, err = repo.CreateRelationship(ctx, &graph.RelationshipInput{
		Type:            "SENT",
		FromType:        "discovered_entity",
		FromID:          ken.ID,
		ToType:          "email",
		ToID:            email2.ID,
		ConfidenceScore: 1.0,
		Timestamp:       date2,
	})
	require.NoError(t, err, "Failed to create SENT relationship")

	_, err = repo.CreateRelationship(ctx, &graph.RelationshipInput{
		Type:            "SENT",
		FromType:        "discovered_entity",
		FromID:          andrew.ID,
		ToType:          "email",
		ToID:            email3.ID,
		ConfidenceScore: 1.0,
		Timestamp:       date3,
	})
	require.NoError(t, err, "Failed to create SENT relationship")

	// RECEIVED relationships (email -> person)
	_, err = repo.CreateRelationship(ctx, &graph.RelationshipInput{
		Type:            "RECEIVED",
		FromType:        "email",
		FromID:          email1.ID,
		ToType:          "discovered_entity",
		ToID:            ken.ID,
		ConfidenceScore: 1.0,
		Timestamp:       date1,
	})
	require.NoError(t, err, "Failed to create RECEIVED relationship")

	// COMMUNICATES_WITH relationships (person <-> person)
	_, err = repo.CreateRelationship(ctx, &graph.RelationshipInput{
		Type:            "COMMUNICATES_WITH",
		FromType:        "discovered_entity",
		FromID:          jeff.ID,
		ToType:          "discovered_entity",
		ToID:            ken.ID,
		ConfidenceScore: 0.95,
		Timestamp:       date1,
		Properties:      map[string]interface{}{"frequency": 2},
	})
	require.NoError(t, err, "Failed to create COMMUNICATES_WITH relationship")

	// MENTIONS relationships (email -> organization/concept)
	_, err = repo.CreateRelationship(ctx, &graph.RelationshipInput{
		Type:            "MENTIONS",
		FromType:        "email",
		FromID:          email1.ID,
		ToType:          "discovered_entity",
		ToID:            enronEntity.ID,
		ConfidenceScore: 0.90,
		Timestamp:       date1,
	})
	require.NoError(t, err, "Failed to create MENTIONS relationship")

	t.Log("Test entities and relationships created successfully")
}

// setupAPIRouter creates and configures the API router for testing
func setupAPIRouter(handler *api.Handler) *chi.Mux {
	r := chi.NewRouter()

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/entities/{id}", handler.GetEntity)
		r.Get("/entities", handler.SearchEntities)
		r.Get("/entities/{id}/relationships", handler.GetEntityRelationships)
		r.Get("/entities/{id}/neighbors", handler.GetEntityNeighbors)
		r.Post("/entities/path", handler.FindPath)
		r.Post("/entities/search", handler.SemanticSearch)
	})

	return r
}

// generateMockEmbedding creates a simple mock embedding vector for testing
// In production, this would come from the LLM
func generateMockEmbedding(t *testing.T, text string) []float32 {
	t.Helper()

	// Create a simple deterministic embedding based on text
	// For testing purposes, just create a vector of consistent length
	embedding := make([]float32, 1024) // mxbai-embed-large dimension
	for i := range embedding {
		// Use a simple hash-like function for deterministic values
		embedding[i] = float32((len(text) + i) % 100) / 100.0
	}
	return embedding
}

// Additional helper for reading HTTP response body
func readBody(t *testing.T, r io.Reader) string {
	t.Helper()
	body, err := io.ReadAll(r)
	require.NoError(t, err, "Failed to read response body")
	return string(body)
}
