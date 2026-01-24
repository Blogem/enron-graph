package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepository implements a subset of graph.Repository for testing
type mockRepository struct {
	entities      map[int]*ent.DiscoveredEntity
	relationships map[int]*ent.Relationship
	nextID        int
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		entities:      make(map[int]*ent.DiscoveredEntity),
		relationships: make(map[int]*ent.Relationship),
		nextID:        1,
	}
}

func (m *mockRepository) FindEntityByID(ctx context.Context, id int) (*ent.DiscoveredEntity, error) {
	if entity, ok := m.entities[id]; ok {
		return entity, nil
	}
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) FindEntitiesByType(ctx context.Context, typeCategory string) ([]*ent.DiscoveredEntity, error) {
	var results []*ent.DiscoveredEntity
	for _, entity := range m.entities {
		if typeCategory == "" || entity.TypeCategory == typeCategory {
			results = append(results, entity)
		}
	}
	return results, nil
}

func (m *mockRepository) FindRelationshipsByEntity(ctx context.Context, entityType string, entityID int) ([]*ent.Relationship, error) {
	var results []*ent.Relationship
	for _, rel := range m.relationships {
		if rel.FromID == entityID || rel.ToID == entityID {
			results = append(results, rel)
		}
	}
	return results, nil
}

func (m *mockRepository) TraverseRelationships(ctx context.Context, fromID int, relType string, depth int) ([]*ent.DiscoveredEntity, error) {
	visited := make(map[int]bool)
	var results []*ent.DiscoveredEntity

	if _, ok := m.entities[fromID]; ok {
		visited[fromID] = true

		for _, rel := range m.relationships {
			var neighborID int
			if rel.FromID == fromID {
				neighborID = rel.ToID
			} else if rel.ToID == fromID {
				neighborID = rel.FromID
			}

			if neighborID > 0 && !visited[neighborID] {
				if neighbor, ok := m.entities[neighborID]; ok {
					results = append(results, neighbor)
					visited[neighborID] = true
				}
			}
		}
	}

	return results, nil
}

func (m *mockRepository) FindShortestPath(ctx context.Context, fromID, toID int) ([]*ent.Relationship, error) {
	if fromID == toID {
		return []*ent.Relationship{}, nil
	}

	queue := []int{fromID}
	visited := make(map[int]bool)
	parent := make(map[int]*ent.Relationship)
	visited[fromID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, rel := range m.relationships {
			var nextID int
			var pathRel *ent.Relationship

			if rel.FromID == current {
				nextID = rel.ToID
				pathRel = rel
			} else if rel.ToID == current {
				nextID = rel.FromID
				pathRel = rel
			}

			if nextID > 0 && !visited[nextID] {
				visited[nextID] = true
				parent[nextID] = pathRel
				queue = append(queue, nextID)

				if nextID == toID {
					var path []*ent.Relationship
					node := toID
					for node != fromID {
						rel := parent[node]
						path = append([]*ent.Relationship{rel}, path...)
						if rel.FromID == node {
							node = rel.ToID
						} else {
							node = rel.FromID
						}
					}
					return path, nil
				}
			}
		}
	}

	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) SimilaritySearch(ctx context.Context, embedding []float32, topK int, threshold float64) ([]*ent.DiscoveredEntity, error) {
	var results []*ent.DiscoveredEntity
	count := 0
	for _, entity := range m.entities {
		if count >= topK {
			break
		}
		results = append(results, entity)
		count++
	}
	return results, nil
}

func (m *mockRepository) Close() error {
	return nil
}

// Helper type for handler (will be implemented in handlers.go)
type Handler struct {
	repo interface{}
}

func NewHandler(repo interface{}) *Handler {
	return &Handler{repo: repo}
}

// Placeholder methods (to be implemented in handlers.go)
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in handlers.go
}

func (h *Handler) SearchEntities(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in handlers.go
}

func (h *Handler) GetEntityRelationships(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in handlers.go
}

func (h *Handler) GetEntityNeighbors(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in handlers.go
}

func (h *Handler) FindPath(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in handlers.go
}

func (h *Handler) SemanticSearch(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in handlers.go
}

// T055: Contract tests for entity endpoints
func TestGetEntityByID_Success(t *testing.T) {
	repo := newMockRepository()
	entity := &ent.DiscoveredEntity{
		ID:              1,
		UniqueID:        "jeff.skilling@enron.com",
		TypeCategory:    "person",
		Name:            "Jeff Skilling",
		Properties:      map[string]interface{}{"title": "CEO"},
		Embedding:       []float32{0.1, 0.2, 0.3},
		ConfidenceScore: 0.95,
	}
	repo.entities[1] = entity

	handler := NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/entities/1", nil)
	w := httptest.NewRecorder()

	handler.GetEntity(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, float64(1), response["id"])
	assert.Equal(t, "person", response["type_category"])
	assert.Equal(t, "Jeff Skilling", response["name"])
	assert.Equal(t, 0.95, response["confidence_score"])
}

func TestGetEntityByID_NotFound(t *testing.T) {
	repo := newMockRepository()
	handler := NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/entities/999", nil)
	w := httptest.NewRecorder()

	handler.GetEntity(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "not found")
}

func TestSearchEntities_FilterByType(t *testing.T) {
	repo := newMockRepository()
	repo.entities[1] = &ent.DiscoveredEntity{
		ID:           1,
		UniqueID:     "person1",
		TypeCategory: "person",
		Name:         "John Doe",
	}
	repo.entities[2] = &ent.DiscoveredEntity{
		ID:           2,
		UniqueID:     "org1",
		TypeCategory: "organization",
		Name:         "Enron",
	}
	repo.entities[3] = &ent.DiscoveredEntity{
		ID:           3,
		UniqueID:     "person2",
		TypeCategory: "person",
		Name:         "Jane Smith",
	}

	handler := NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/entities?type=person", nil)
	w := httptest.NewRecorder()

	handler.SearchEntities(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	entities := response["entities"].([]interface{})
	assert.Equal(t, 2, len(entities))

	for _, e := range entities {
		entity := e.(map[string]interface{})
		assert.Equal(t, "person", entity["type_category"])
	}
}

func TestSearchEntities_FilterByName(t *testing.T) {
	repo := newMockRepository()
	repo.entities[1] = &ent.DiscoveredEntity{
		ID:           1,
		UniqueID:     "person1",
		TypeCategory: "person",
		Name:         "John Doe",
	}
	repo.entities[2] = &ent.DiscoveredEntity{
		ID:           2,
		UniqueID:     "person2",
		TypeCategory: "person",
		Name:         "Jane Smith",
	}

	handler := NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/entities?name=John Doe", nil)
	w := httptest.NewRecorder()

	handler.SearchEntities(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	entities := response["entities"].([]interface{})
	assert.Equal(t, 1, len(entities))

	entity := entities[0].(map[string]interface{})
	assert.Equal(t, "John Doe", entity["name"])
}

func TestSearchEntities_InvalidParameters(t *testing.T) {
	repo := newMockRepository()
	handler := NewHandler(repo)

	testCases := []struct {
		name          string
		queryString   string
		expectedError string
	}{
		{
			name:          "invalid type",
			queryString:   "type=invalid_type_with_special_chars!@#",
			expectedError: "invalid",
		},
		{
			name:          "empty name",
			queryString:   "name=",
			expectedError: "invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/entities?"+tc.queryString, nil)
			w := httptest.NewRecorder()

			handler.SearchEntities(w, req)

			if w.Code == http.StatusBadRequest {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tc.expectedError)
			}
		})
	}
}

func TestSearchEntities_ValidResponseSchema(t *testing.T) {
	repo := newMockRepository()
	repo.entities[1] = &ent.DiscoveredEntity{
		ID:              1,
		UniqueID:        "test",
		TypeCategory:    "person",
		Name:            "Test User",
		Properties:      map[string]interface{}{"key": "value"},
		Embedding:       []float32{0.1},
		ConfidenceScore: 0.8,
	}

	handler := NewHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/entities", nil)
	w := httptest.NewRecorder()

	handler.SearchEntities(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "entities")
	assert.Contains(t, response, "total")

	entities := response["entities"].([]interface{})
	if len(entities) > 0 {
		entity := entities[0].(map[string]interface{})
		assert.Contains(t, entity, "id")
		assert.Contains(t, entity, "unique_id")
		assert.Contains(t, entity, "type_category")
		assert.Contains(t, entity, "name")
		assert.Contains(t, entity, "properties")
		assert.Contains(t, entity, "confidence_score")
	}
}

// T056: Contract tests for relationship endpoints
func TestGetEntityRelationships_Success(t *testing.T) {
	repo := newMockRepository()
	repo.entities[1] = &ent.DiscoveredEntity{
		ID:           1,
		UniqueID:     "person1",
		TypeCategory: "person",
		Name:         "Jeff Skilling",
	}
	repo.entities[2] = &ent.DiscoveredEntity{
		ID:           2,
		UniqueID:     "person2",
		TypeCategory: "person",
		Name:         "Kenneth Lay",
	}
	repo.relationships[1] = &ent.Relationship{
		ID:              1,
		Type:            "COMMUNICATES_WITH",
		FromType:        "discovered_entity",
		FromID:          1,
		ToType:          "discovered_entity",
		ToID:            2,
		ConfidenceScore: 0.9,
	}

	handler := NewHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/entities/1/relationships", nil)
	w := httptest.NewRecorder()

	handler.GetEntityRelationships(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	relationships := response["relationships"].([]interface{})
	assert.GreaterOrEqual(t, len(relationships), 1)

	rel := relationships[0].(map[string]interface{})
	assert.Equal(t, "COMMUNICATES_WITH", rel["type"])
	assert.Equal(t, float64(1), rel["from_id"])
	assert.Equal(t, float64(2), rel["to_id"])
}

func TestGetEntityNeighbors_Depth1(t *testing.T) {
	repo := newMockRepository()

	repo.entities[1] = &ent.DiscoveredEntity{ID: 1, UniqueID: "e1", TypeCategory: "person", Name: "Entity 1"}
	repo.entities[2] = &ent.DiscoveredEntity{ID: 2, UniqueID: "e2", TypeCategory: "person", Name: "Entity 2"}
	repo.entities[3] = &ent.DiscoveredEntity{ID: 3, UniqueID: "e3", TypeCategory: "person", Name: "Entity 3"}

	repo.relationships[1] = &ent.Relationship{ID: 1, Type: "KNOWS", FromID: 1, ToID: 2}
	repo.relationships[2] = &ent.Relationship{ID: 2, Type: "KNOWS", FromID: 2, ToID: 3}

	handler := NewHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/entities/1/neighbors?depth=1", nil)
	w := httptest.NewRecorder()

	handler.GetEntityNeighbors(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	neighbors := response["neighbors"].([]interface{})
	assert.Equal(t, 1, len(neighbors))

	neighbor := neighbors[0].(map[string]interface{})
	assert.Equal(t, "Entity 2", neighbor["name"])
}

func TestGetEntityNeighbors_Depth3(t *testing.T) {
	repo := newMockRepository()

	repo.entities[1] = &ent.DiscoveredEntity{ID: 1, UniqueID: "e1", TypeCategory: "person", Name: "Entity 1"}
	repo.entities[2] = &ent.DiscoveredEntity{ID: 2, UniqueID: "e2", TypeCategory: "person", Name: "Entity 2"}
	repo.entities[3] = &ent.DiscoveredEntity{ID: 3, UniqueID: "e3", TypeCategory: "person", Name: "Entity 3"}
	repo.entities[4] = &ent.DiscoveredEntity{ID: 4, UniqueID: "e4", TypeCategory: "person", Name: "Entity 4"}

	repo.relationships[1] = &ent.Relationship{ID: 1, Type: "KNOWS", FromID: 1, ToID: 2}
	repo.relationships[2] = &ent.Relationship{ID: 2, Type: "KNOWS", FromID: 2, ToID: 3}
	repo.relationships[3] = &ent.Relationship{ID: 3, Type: "KNOWS", FromID: 3, ToID: 4}

	handler := NewHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/entities/1/neighbors?depth=3", nil)
	w := httptest.NewRecorder()

	handler.GetEntityNeighbors(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	neighbors := response["neighbors"].([]interface{})
	assert.GreaterOrEqual(t, len(neighbors), 1)
}

func TestGetEntityNeighbors_DepthValidation(t *testing.T) {
	repo := newMockRepository()
	handler := NewHandler(repo)

	testCases := []struct {
		name        string
		depth       string
		expectError bool
	}{
		{"valid depth 1", "1", false},
		{"valid depth 5", "5", false},
		{"invalid negative depth", "-1", true},
		{"invalid zero depth", "0", true},
		{"invalid too large depth", "100", true},
		{"invalid non-numeric", "abc", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/entities/1/neighbors?depth="+tc.depth, nil)
			w := httptest.NewRecorder()

			handler.GetEntityNeighbors(w, req)

			if tc.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			}
		})
	}
}

// T057: Contract tests for path finding
func TestFindPath_Success(t *testing.T) {
	repo := newMockRepository()

	repo.entities[1] = &ent.DiscoveredEntity{ID: 1, UniqueID: "e1", TypeCategory: "person", Name: "Entity 1"}
	repo.entities[2] = &ent.DiscoveredEntity{ID: 2, UniqueID: "e2", TypeCategory: "person", Name: "Entity 2"}
	repo.entities[3] = &ent.DiscoveredEntity{ID: 3, UniqueID: "e3", TypeCategory: "person", Name: "Entity 3"}

	repo.relationships[1] = &ent.Relationship{ID: 1, Type: "KNOWS", FromID: 1, ToID: 2}
	repo.relationships[2] = &ent.Relationship{ID: 2, Type: "KNOWS", FromID: 2, ToID: 3}

	handler := NewHandler(repo)

	requestBody := map[string]interface{}{
		"source_id": 1,
		"target_id": 3,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/entities/path", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.FindPath(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	path := response["path"].([]interface{})
	assert.GreaterOrEqual(t, len(path), 1)

	firstRel := path[0].(map[string]interface{})
	assert.Contains(t, firstRel, "type")
	assert.Contains(t, firstRel, "from_id")
	assert.Contains(t, firstRel, "to_id")
}

func TestFindPath_NoPathExists(t *testing.T) {
	repo := newMockRepository()

	repo.entities[1] = &ent.DiscoveredEntity{ID: 1, UniqueID: "e1", TypeCategory: "person", Name: "Entity 1"}
	repo.entities[2] = &ent.DiscoveredEntity{ID: 2, UniqueID: "e2", TypeCategory: "person", Name: "Entity 2"}

	handler := NewHandler(repo)

	requestBody := map[string]interface{}{
		"source_id": 1,
		"target_id": 2,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/entities/path", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.FindPath(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "path")
}

func TestFindPath_InvalidRequestBody(t *testing.T) {
	repo := newMockRepository()
	handler := NewHandler(repo)

	testCases := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid json", "{invalid}"},
		{"missing source_id", `{"target_id": 2}`},
		{"missing target_id", `{"source_id": 1}`},
		{"invalid source_id type", `{"source_id": "abc", "target_id": 2}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/entities/path", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.FindPath(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "error")
		})
	}
}

// T058: Contract tests for semantic search
func TestSemanticSearch_Success(t *testing.T) {
	repo := newMockRepository()

	repo.entities[1] = &ent.DiscoveredEntity{
		ID:           1,
		UniqueID:     "concept1",
		TypeCategory: "concept",
		Name:         "Energy Trading",
		Embedding:    []float32{0.1, 0.2, 0.3},
	}
	repo.entities[2] = &ent.DiscoveredEntity{
		ID:           2,
		UniqueID:     "concept2",
		TypeCategory: "concept",
		Name:         "Power Markets",
		Embedding:    []float32{0.15, 0.25, 0.35},
	}

	handler := NewHandler(repo)

	requestBody := map[string]interface{}{
		"text":      "energy trading strategies",
		"top_k":     5,
		"threshold": 0.7,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/entities/search", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SemanticSearch(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	results := response["results"].([]interface{})
	assert.GreaterOrEqual(t, len(results), 1)

	result := results[0].(map[string]interface{})
	assert.Contains(t, result, "entity")
	assert.Contains(t, result, "similarity")

	similarity := result["similarity"].(float64)
	assert.GreaterOrEqual(t, similarity, 0.0)
	assert.LessOrEqual(t, similarity, 1.0)
}

func TestSemanticSearch_ValidateRequestBody(t *testing.T) {
	repo := newMockRepository()
	handler := NewHandler(repo)

	testCases := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"invalid json", "{invalid}"},
		{"missing text", `{"top_k": 5}`},
		{"empty text", `{"text": "", "top_k": 5}`},
		{"invalid top_k", `{"text": "test", "top_k": -1}`},
		{"invalid threshold", `{"text": "test", "threshold": 1.5}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/entities/search", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SemanticSearch(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "error")
		})
	}
}

func TestSemanticSearch_RankedResults(t *testing.T) {
	repo := newMockRepository()

	for i := 1; i <= 5; i++ {
		repo.entities[i] = &ent.DiscoveredEntity{
			ID:           i,
			UniqueID:     "concept" + string(rune(i)),
			TypeCategory: "concept",
			Name:         "Concept " + string(rune(i)),
			Embedding:    []float32{float32(i) * 0.1, float32(i) * 0.2},
		}
	}

	handler := NewHandler(repo)

	requestBody := map[string]interface{}{
		"text":  "test query",
		"top_k": 3,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/entities/search", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SemanticSearch(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	results := response["results"].([]interface{})
	assert.LessOrEqual(t, len(results), 3)

	if len(results) > 1 {
		for i := 0; i < len(results)-1; i++ {
			curr := results[i].(map[string]interface{})
			next := results[i+1].(map[string]interface{})

			currSim := curr["similarity"].(float64)
			nextSim := next["similarity"].(float64)

			assert.GreaterOrEqual(t, currSim, nextSim, "Results should be sorted by similarity descending")
		}
	}
}
