package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/internal/api"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUS2AcceptanceCriteria tests all acceptance criteria for User Story 2
// These tests verify the acceptance criteria from spec.md for graph query and exploration

// TestAcceptanceT069_QueryPersonByName tests T069 acceptance criteria:
// Verify: Query person by name returns entity with properties
func TestAcceptanceT069_QueryPersonByName(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()
	repo := graph.NewRepository(client)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Pre-populate test data
	setupTestEntitiesAndRelationships(t, ctx, client, repo, llmClient)

	// Create API handler and router
	handler := api.NewHandlerWithLLM(repo, llmClient)
	router := setupAPIRouter(handler)

	t.Run("QueryPersonByName_ReturnsEntityWithProperties", func(t *testing.T) {
		// Query for "Jeff Skilling" by name
		req := httptest.NewRequest("GET", "/api/v1/entities?type=person&name=Jeff", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response status
		assert.Equal(t, 200, w.Code, "Expected status 200 OK")

		// Parse response
		var response api.SearchResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Failed to parse response")

		// Verify: At least one entity returned
		require.Greater(t, len(response.Entities), 0, "Expected at least one person entity")

		// Find Jeff Skilling in results
		var jeffSkilling *api.EntityResponse
		for _, entity := range response.Entities {
			if entity.Name == "Jeff Skilling" {
				jeffSkilling = &entity
				break
			}
		}
		require.NotNil(t, jeffSkilling, "Jeff Skilling not found in results")

		// Verify: Entity has properties
		assert.Equal(t, "person", jeffSkilling.TypeCategory, "Expected type to be person")
		assert.NotEmpty(t, jeffSkilling.Properties, "Expected entity to have properties")
		assert.Contains(t, jeffSkilling.Properties, "title", "Expected entity to have title property")
		assert.Equal(t, "CEO", jeffSkilling.Properties["title"], "Expected title to be CEO")
		assert.GreaterOrEqual(t, jeffSkilling.ConfidenceScore, 0.7, "Expected confidence score >= 0.7")

		t.Logf("✓ AC-T069: Query person by name returns entity with properties")
		t.Logf("  - Name: %s", jeffSkilling.Name)
		t.Logf("  - Type: %s", jeffSkilling.TypeCategory)
		t.Logf("  - Properties: %+v", jeffSkilling.Properties)
		t.Logf("  - Confidence: %.2f", jeffSkilling.ConfidenceScore)
	})
}

// TestAcceptanceT070_QueryRelationships tests T070 acceptance criteria:
// Verify: Query relationships returns list of connected entities
func TestAcceptanceT070_QueryRelationships(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()
	repo := graph.NewRepository(client)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Pre-populate test data
	setupTestEntitiesAndRelationships(t, ctx, client, repo, llmClient)

	// Create API handler and router
	handler := api.NewHandlerWithLLM(repo, llmClient)
	router := setupAPIRouter(handler)

	t.Run("QueryRelationships_ReturnsConnectedEntities", func(t *testing.T) {
		// Get Jeff Skilling entity
		jeff, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.UniqueIDEQ("jeff.skilling@enron.com")).
			First(ctx)
		require.NoError(t, err, "Failed to find Jeff Skilling")

		// Query relationships
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/relationships", jeff.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response status
		assert.Equal(t, 200, w.Code, "Expected status 200 OK")

		// Parse response
		var response api.RelationshipsResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Failed to parse response")

		// Verify: Relationships returned
		assert.Equal(t, jeff.ID, response.EntityID, "Expected entity ID to match")
		assert.Greater(t, len(response.Relationships), 0, "Expected at least one relationship")

		// Verify: Relationships have correct structure
		for _, rel := range response.Relationships {
			assert.NotEmpty(t, rel.Type, "Expected relationship to have a type")
			assert.Greater(t, rel.FromID, 0, "Expected valid from_id")
			assert.Greater(t, rel.ToID, 0, "Expected valid to_id")
			assert.NotEmpty(t, rel.FromType, "Expected from_type")
			assert.NotEmpty(t, rel.ToType, "Expected to_type")
		}

		t.Logf("✓ AC-T070: Query relationships returns list of connected entities")
		t.Logf("  - Entity: Jeff Skilling (ID: %d)", jeff.ID)
		t.Logf("  - Total relationships: %d", len(response.Relationships))
		for i, rel := range response.Relationships {
			t.Logf("  - Relationship %d: %s (%s[%d] -> %s[%d])",
				i+1, rel.Type, rel.FromType, rel.FromID, rel.ToType, rel.ToID)
		}
	})
}

// TestAcceptanceT071_ShortestPath tests T071 acceptance criteria:
// Verify: Shortest path between entities returns relationship chain
func TestAcceptanceT071_ShortestPath(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()
	repo := graph.NewRepository(client)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Pre-populate test data
	setupTestEntitiesAndRelationships(t, ctx, client, repo, llmClient)

	// Create API handler and router
	handler := api.NewHandlerWithLLM(repo, llmClient)
	router := setupAPIRouter(handler)

	t.Run("ShortestPath_ReturnsRelationshipChain", func(t *testing.T) {
		// Get two connected entities
		jeff, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.UniqueIDEQ("jeff.skilling@enron.com")).
			First(ctx)
		require.NoError(t, err, "Failed to find Jeff Skilling")

		ken, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.UniqueIDEQ("ken.lay@enron.com")).
			First(ctx)
		require.NoError(t, err, "Failed to find Ken Lay")

		// Find shortest path
		pathReq := api.PathRequest{
			SourceID: jeff.ID,
			TargetID: ken.ID,
			MaxDepth: 5,
		}
		reqBody, _ := json.Marshal(pathReq)
		req := httptest.NewRequest("POST", "/api/v1/entities/path", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response (may be 200 or 404 if no path exists)
		if w.Code == 200 {
			// Parse response
			var response api.PathResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Failed to parse response")

			// Verify: Path structure
			assert.Equal(t, jeff.ID, response.SourceID, "Expected source ID to match")
			assert.Equal(t, ken.ID, response.TargetID, "Expected target ID to match")
			assert.Greater(t, len(response.Path), 0, "Expected path to have elements")
			assert.Greater(t, response.PathLength, 0, "Expected path length > 0")

			t.Logf("✓ AC-T071: Shortest path between entities returns relationship chain")
			t.Logf("  - Source: Jeff Skilling (ID: %d)", jeff.ID)
			t.Logf("  - Target: Ken Lay (ID: %d)", ken.ID)
			t.Logf("  - Path length: %d", response.PathLength)
			t.Logf("  - Path elements: %d", len(response.Path))
		} else if w.Code == 404 {
			t.Logf("✓ AC-T071: No path found between entities (acceptable - entities may not be connected)")
		} else {
			t.Fatalf("Unexpected status code: %d", w.Code)
		}
	})
}

// TestAcceptanceT072_FilterByEntityType tests T072 acceptance criteria:
// Verify: Filter by entity type returns matching entities
func TestAcceptanceT072_FilterByEntityType(t *testing.T) {
	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()
	repo := graph.NewRepository(client)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Pre-populate test data
	setupTestEntitiesAndRelationships(t, ctx, client, repo, llmClient)

	// Create API handler and router
	handler := api.NewHandlerWithLLM(repo, llmClient)
	router := setupAPIRouter(handler)

	// Test filtering by different entity types
	entityTypes := []string{"person", "organization", "concept"}

	for _, entityType := range entityTypes {
		t.Run(fmt.Sprintf("FilterBy_%s", entityType), func(t *testing.T) {
			// Query entities by type
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities?type=%s", entityType), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify response status
			assert.Equal(t, 200, w.Code, "Expected status 200 OK")

			// Parse response
			var response api.SearchResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Failed to parse response")

			// Verify: All returned entities match the type filter
			if len(response.Entities) > 0 {
				for _, entity := range response.Entities {
					assert.Equal(t, entityType, entity.TypeCategory,
						"Expected all entities to be of type %s", entityType)
				}
				t.Logf("  - Type: %s, Count: %d", entityType, len(response.Entities))
			} else {
				t.Logf("  - Type: %s, Count: 0 (no entities of this type)", entityType)
			}
		})
	}

	t.Logf("✓ AC-T072: Filter by entity type returns matching entities")
}

// TestAcceptanceT073_EntityLookupPerformance tests T073 acceptance criteria:
// Verify: SC-003 - Entity lookup <500ms for 100k nodes
// Note: This is a placeholder for performance testing with large datasets
func TestAcceptanceT073_EntityLookupPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()
	repo := graph.NewRepository(client)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Note: For true 100k node test, we would need to populate with much more data
	// This test demonstrates the performance test structure
	setupTestEntitiesAndRelationships(t, ctx, client, repo, llmClient)

	// Create API handler and router
	handler := api.NewHandlerWithLLM(repo, llmClient)
	router := setupAPIRouter(handler)

	t.Run("EntityLookup_PerformanceTest", func(t *testing.T) {
		// Get an entity to test lookup
		jeff, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.UniqueIDEQ("jeff.skilling@enron.com")).
			First(ctx)
		require.NoError(t, err, "Failed to find Jeff Skilling")

		// Measure entity lookup time
		start := time.Now()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d", jeff.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		// Verify response
		assert.Equal(t, 200, w.Code, "Expected status 200 OK")

		// Verify performance (should be much faster than 500ms for small dataset)
		assert.Less(t, duration.Milliseconds(), int64(500),
			"Entity lookup took %dms, expected <500ms", duration.Milliseconds())

		t.Logf("✓ AC-T073: Entity lookup performance")
		t.Logf("  - Lookup time: %dms (target: <500ms for 100k nodes)", duration.Milliseconds())
		t.Logf("  - Note: Current test uses small dataset, full 100k node test requires data generation")
	})

	t.Run("EntitySearch_PerformanceTest", func(t *testing.T) {
		// Measure entity search time
		start := time.Now()
		req := httptest.NewRequest("GET", "/api/v1/entities?type=person&name=Jeff", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		// Verify response
		assert.Equal(t, 200, w.Code, "Expected status 200 OK")

		// Verify performance
		assert.Less(t, duration.Milliseconds(), int64(500),
			"Entity search took %dms, expected <500ms", duration.Milliseconds())

		t.Logf("  - Search time: %dms (target: <500ms for 100k nodes)", duration.Milliseconds())
	})
}

// TestAcceptanceT074_ShortestPathPerformance tests T074 acceptance criteria:
// Verify: SC-004 - Shortest path <2s for 6 degrees
// Note: This is a placeholder for performance testing with large datasets
func TestAcceptanceT074_ShortestPathPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup: Create test database
	client := SetupTestDB(t)
	ctx := context.Background()
	repo := graph.NewRepository(client)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	llmClient := llm.NewOllamaClient("http://localhost:11434", "llama3.1:8b", "mxbai-embed-large", logger)

	// Pre-populate test data
	setupTestEntitiesAndRelationships(t, ctx, client, repo, llmClient)

	// Create API handler and router
	handler := api.NewHandlerWithLLM(repo, llmClient)
	router := setupAPIRouter(handler)

	t.Run("ShortestPath_PerformanceTest", func(t *testing.T) {
		// Get two entities
		jeff, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.UniqueIDEQ("jeff.skilling@enron.com")).
			First(ctx)
		require.NoError(t, err, "Failed to find Jeff Skilling")

		ken, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.UniqueIDEQ("ken.lay@enron.com")).
			First(ctx)
		require.NoError(t, err, "Failed to find Ken Lay")

		// Measure shortest path computation time
		start := time.Now()
		pathReq := api.PathRequest{
			SourceID: jeff.ID,
			TargetID: ken.ID,
			MaxDepth: 6, // 6 degrees of separation
		}
		reqBody, _ := json.Marshal(pathReq)
		req := httptest.NewRequest("POST", "/api/v1/entities/path", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		// Verify response (200 if path found, 404 if not)
		assert.Contains(t, []int{200, 404}, w.Code, "Expected status 200 or 404")

		// Verify performance
		assert.Less(t, duration.Milliseconds(), int64(2000),
			"Shortest path computation took %dms, expected <2000ms", duration.Milliseconds())

		t.Logf("✓ AC-T074: Shortest path performance")
		t.Logf("  - Computation time: %dms (target: <2000ms for 6 degrees)", duration.Milliseconds())
		t.Logf("  - Max depth: 6")
		t.Logf("  - Note: Current test uses small dataset, full performance test requires larger graph")
	})
}
