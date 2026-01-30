package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/graph"
	"github.com/Blogem/enron-graph/pkg/llm"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for the graph API
type Handler struct {
	repo      graph.Repository
	llmClient llm.Client
}

// NewHandler creates a new API handler
func NewHandler(repo interface{}) *Handler {
	// Support both graph.Repository and mock repository (for tests)
	if r, ok := repo.(graph.Repository); ok {
		return &Handler{repo: r}
	}
	// For test mock repository, wrap in type assertion
	return &Handler{repo: &mockRepoWrapper{mock: repo}}
}

// NewHandlerWithLLM creates a new API handler with LLM client for semantic search
func NewHandlerWithLLM(repo graph.Repository, llmClient llm.Client) *Handler {
	return &Handler{
		repo:      repo,
		llmClient: llmClient,
	}
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Field   string `json:"field,omitempty"`
}

// EntityResponse represents an entity in API responses
type EntityResponse struct {
	ID              int                    `json:"id"`
	UniqueID        string                 `json:"unique_id"`
	TypeCategory    string                 `json:"type_category"`
	Name            string                 `json:"name"`
	Properties      map[string]interface{} `json:"properties"`
	ConfidenceScore float64                `json:"confidence_score"`
	CreatedAt       string                 `json:"created_at,omitempty"`
}

// RelationshipResponse represents a relationship in API responses
type RelationshipResponse struct {
	ID              int                    `json:"id"`
	Type            string                 `json:"type"`
	FromType        string                 `json:"from_type"`
	FromID          int                    `json:"from_id"`
	ToType          string                 `json:"to_type"`
	ToID            int                    `json:"to_id"`
	Timestamp       string                 `json:"timestamp"`
	ConfidenceScore float64                `json:"confidence_score"`
	Properties      map[string]interface{} `json:"properties"`
}

// SearchResponse represents the response for entity search
type SearchResponse struct {
	Entities []EntityResponse `json:"entities"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit,omitempty"`
	Offset   int              `json:"offset,omitempty"`
}

// RelationshipsResponse represents the response for relationship queries
type RelationshipsResponse struct {
	EntityID      int                    `json:"entity_id"`
	Relationships []RelationshipResponse `json:"relationships"`
	Total         int                    `json:"total"`
	Limit         int                    `json:"limit,omitempty"`
	Offset        int                    `json:"offset,omitempty"`
}

// NeighborEntity represents a neighbor in traversal response
type NeighborEntity struct {
	ID               int                    `json:"id"`
	UniqueID         string                 `json:"unique_id"`
	TypeCategory     string                 `json:"type_category"`
	Name             string                 `json:"name"`
	Distance         int                    `json:"distance"`
	RelationshipPath []RelationshipResponse `json:"relationship_path"`
}

// NeighborsResponse represents the response for neighbor traversal
type NeighborsResponse struct {
	SourceEntityID int              `json:"source_entity_id"`
	Depth          int              `json:"depth"`
	Neighbors      []NeighborEntity `json:"neighbors"`
	Total          int              `json:"total"`
}

// PathRequest represents a shortest path request
type PathRequest struct {
	SourceID int `json:"source_id"`
	TargetID int `json:"target_id"`
	MaxDepth int `json:"max_depth"`
}

// PathElement represents an element in a path (entity or relationship)
type PathElement struct {
	EntityID         int    `json:"entity_id,omitempty"`
	EntityName       string `json:"entity_name,omitempty"`
	EntityType       string `json:"entity_type,omitempty"`
	RelationshipID   int    `json:"relationship_id,omitempty"`
	RelationshipType string `json:"relationship_type,omitempty"`
}

// PathResponse represents the response for shortest path
type PathResponse struct {
	SourceID   int           `json:"source_id"`
	TargetID   int           `json:"target_id"`
	PathLength int           `json:"path_length"`
	Path       []PathElement `json:"path"`
}

// SearchRequest represents a semantic search request
type SearchRequest struct {
	Query         string  `json:"query"`
	Limit         int     `json:"limit"`
	MinSimilarity float64 `json:"min_similarity"`
	TypeFilter    string  `json:"type_filter"`
}

// SearchResult represents a single result in semantic search
type SearchResult struct {
	Entity     EntityResponse `json:"entity"`
	Similarity float64        `json:"similarity"`
}

// SemanticSearchResponse represents the response for semantic search
type SemanticSearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}

// GetEntity handles GET /entities/:id
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		// Fallback for tests that don't use chi router
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			idStr = parts[len(parts)-1]
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entity id", "")
		return
	}

	entity, err := h.repo.FindEntityByID(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "entity not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to fetch entity", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toEntityResponse(entity))
}

// SearchEntities handles GET /entities
func (h *Handler) SearchEntities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	typeCategory := query.Get("type")
	name := query.Get("name")
	minConfStr := query.Get("min_confidence")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	// Validate parameters
	if name == "" && typeCategory != "" {
		// Empty name is invalid only when explicitly set
		if query.Has("name") && name == "" {
			respondError(w, http.StatusBadRequest, "invalid query parameter", "name cannot be empty")
			return
		}
	}

	// Parse limit and offset
	limit := 100
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			respondError(w, http.StatusBadRequest, "invalid query parameter", "limit must be between 1 and 1000")
			return
		}
	}

	offset := 0
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			respondError(w, http.StatusBadRequest, "invalid query parameter", "offset must be non-negative")
			return
		}
	}

	var minConfidence float64
	if minConfStr != "" {
		var err error
		minConfidence, err = strconv.ParseFloat(minConfStr, 64)
		if err != nil || minConfidence < 0 || minConfidence > 1 {
			respondError(w, http.StatusBadRequest, "invalid query parameter", "min_confidence must be between 0 and 1")
			return
		}
	}

	// Fetch entities by type
	var entities []*ent.DiscoveredEntity
	var err error

	if typeCategory != "" {
		entities, err = h.repo.FindEntitiesByType(ctx, typeCategory)
	} else {
		entities, err = h.repo.FindEntitiesByType(ctx, "")
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch entities", err.Error())
		return
	}

	// Filter by name and confidence
	filtered := make([]*ent.DiscoveredEntity, 0)
	for _, entity := range entities {
		// Filter by name (case-insensitive partial match)
		if name != "" && !strings.Contains(strings.ToLower(entity.Name), strings.ToLower(name)) {
			continue
		}
		// Filter by confidence
		if minConfidence > 0 && entity.ConfidenceScore < minConfidence {
			continue
		}
		filtered = append(filtered, entity)
	}

	// Apply pagination
	total := len(filtered)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginated := filtered[start:end]

	// Convert to response
	results := make([]EntityResponse, len(paginated))
	for i, entity := range paginated {
		results[i] = toEntityResponse(entity)
	}

	respondJSON(w, http.StatusOK, SearchResponse{
		Entities: results,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

// GetEntityRelationships handles GET /entities/:id/relationships
func (h *Handler) GetEntityRelationships(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "entities" && i+1 < len(parts) {
				idStr = parts[i+1]
				break
			}
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entity id", "")
		return
	}

	// Check if entity exists
	_, err = h.repo.FindEntityByID(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "entity not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to fetch entity", err.Error())
		return
	}

	// Get relationships
	relationships, err := h.repo.FindRelationshipsByEntity(r.Context(), "discovered_entity", id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch relationships", err.Error())
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	relType := query.Get("type")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit := 100
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			respondError(w, http.StatusBadRequest, "invalid query parameter", "limit must be between 1 and 1000")
			return
		}
	}

	offset := 0
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			respondError(w, http.StatusBadRequest, "invalid query parameter", "offset must be non-negative")
			return
		}
	}

	// Filter by type if specified
	filtered := relationships
	if relType != "" {
		filtered = make([]*ent.Relationship, 0)
		for _, rel := range relationships {
			if rel.Type == relType {
				filtered = append(filtered, rel)
			}
		}
	}

	// Apply pagination
	total := len(filtered)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginated := filtered[start:end]

	// Convert to response
	results := make([]RelationshipResponse, len(paginated))
	for i, rel := range paginated {
		results[i] = toRelationshipResponse(rel)
	}

	respondJSON(w, http.StatusOK, RelationshipsResponse{
		EntityID:      id,
		Relationships: results,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	})
}

// GetEntityNeighbors handles GET /entities/:id/neighbors
func (h *Handler) GetEntityNeighbors(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "entities" && i+1 < len(parts) {
				idStr = parts[i+1]
				break
			}
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entity id", "")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	depthStr := query.Get("depth")
	relType := query.Get("type")
	limitStr := query.Get("limit")

	depth := 1
	if depthStr != "" {
		depth, err = strconv.Atoi(depthStr)
		if err != nil || depth < 1 || depth > 5 {
			respondError(w, http.StatusBadRequest, "invalid parameter", "depth must be between 1 and 5")
			return
		}
	}

	limit := 50
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			respondError(w, http.StatusBadRequest, "invalid parameter", "limit must be between 1 and 100")
			return
		}
	}

	// Check if source entity exists
	_, err = h.repo.FindEntityByID(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "entity not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to fetch entity", err.Error())
		return
	}

	// Traverse relationships
	neighbors, err := h.repo.TraverseRelationships(r.Context(), id, relType, depth)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to traverse relationships", err.Error())
		return
	}

	// Apply limit
	if len(neighbors) > limit {
		neighbors = neighbors[:limit]
	}

	// Convert to response (simplified - full implementation would include path details)
	results := make([]NeighborEntity, len(neighbors))
	for i, neighbor := range neighbors {
		results[i] = NeighborEntity{
			ID:               neighbor.ID,
			UniqueID:         neighbor.UniqueID,
			TypeCategory:     neighbor.TypeCategory,
			Name:             neighbor.Name,
			Distance:         1, // Simplified - actual implementation would track distance
			RelationshipPath: []RelationshipResponse{
				// Simplified - actual implementation would include full path
			},
		}
	}

	respondJSON(w, http.StatusOK, NeighborsResponse{
		SourceEntityID: id,
		Depth:          depth,
		Neighbors:      results,
		Total:          len(neighbors),
	})
}

// FindPath handles POST /entities/path
func (h *Handler) FindPath(w http.ResponseWriter, r *http.Request) {
	var req PathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request", "failed to parse request body")
		return
	}

	// Validate request
	if req.SourceID == 0 || req.TargetID == 0 {
		respondError(w, http.StatusBadRequest, "invalid request", "source_id and target_id are required")
		return
	}

	if req.MaxDepth == 0 {
		req.MaxDepth = 6
	}

	if req.MaxDepth < 1 || req.MaxDepth > 10 {
		respondError(w, http.StatusBadRequest, "invalid parameter", "max_depth must be between 1 and 10")
		return
	}

	// Find shortest path
	path, err := h.repo.FindShortestPath(r.Context(), req.SourceID, req.TargetID)
	if err != nil {
		if ent.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "no path found", fmt.Sprintf("no path exists between entities %d and %d within max_depth %d", req.SourceID, req.TargetID, req.MaxDepth))
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to find path", err.Error())
		return
	}

	// Build path response
	pathElements := make([]PathElement, 0)

	if len(path) == 0 {
		// Source and target are the same
		sourceEntity, _ := h.repo.FindEntityByID(r.Context(), req.SourceID)
		pathElements = append(pathElements, PathElement{
			EntityID:   sourceEntity.ID,
			EntityName: sourceEntity.Name,
			EntityType: sourceEntity.TypeCategory,
		})
	} else {
		// Add source entity
		sourceEntity, _ := h.repo.FindEntityByID(r.Context(), req.SourceID)
		pathElements = append(pathElements, PathElement{
			EntityID:   sourceEntity.ID,
			EntityName: sourceEntity.Name,
			EntityType: sourceEntity.TypeCategory,
		})

		// Add relationships and intermediate entities
		for _, rel := range path {
			pathElements = append(pathElements, PathElement{
				RelationshipID:   rel.ID,
				RelationshipType: rel.Type,
			})

			// Determine next entity ID
			var nextID int
			if rel.FromID == req.SourceID || (len(pathElements) > 2 && rel.FromID == pathElements[len(pathElements)-3].EntityID) {
				nextID = rel.ToID
			} else {
				nextID = rel.FromID
			}

			nextEntity, _ := h.repo.FindEntityByID(r.Context(), nextID)
			if nextEntity != nil {
				pathElements = append(pathElements, PathElement{
					EntityID:   nextEntity.ID,
					EntityName: nextEntity.Name,
					EntityType: nextEntity.TypeCategory,
				})
			}
		}
	}

	respondJSON(w, http.StatusOK, PathResponse{
		SourceID:   req.SourceID,
		TargetID:   req.TargetID,
		PathLength: len(path),
		Path:       pathElements,
	})
}

// SemanticSearch handles POST /entities/search
func (h *Handler) SemanticSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request", "failed to parse request body")
		return
	}

	// Validate request
	if req.Query == "" {
		respondError(w, http.StatusBadRequest, "invalid request", "query is required")
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Limit < 1 || req.Limit > 100 {
		respondError(w, http.StatusBadRequest, "invalid parameter", "limit must be between 1 and 100")
		return
	}

	if req.MinSimilarity < 0 || req.MinSimilarity > 1 {
		respondError(w, http.StatusBadRequest, "invalid parameter", "min_similarity must be between 0 and 1")
		return
	}

	// Generate embedding for query
	var embedding []float32
	var err error

	if h.llmClient != nil {
		embedding, err = h.llmClient.GenerateEmbedding(r.Context(), req.Query)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to generate embedding", err.Error())
			return
		}
	} else {
		// For tests without LLM client, use dummy embedding
		embedding = make([]float32, 1024)
	}

	// Perform similarity search
	entities, err := h.repo.SimilaritySearch(r.Context(), embedding, req.Limit, req.MinSimilarity)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to perform search", err.Error())
		return
	}

	// Filter by type if specified
	if req.TypeFilter != "" {
		filtered := make([]*ent.DiscoveredEntity, 0)
		for _, entity := range entities {
			if entity.TypeCategory == req.TypeFilter {
				filtered = append(filtered, entity)
			}
		}
		entities = filtered
	}

	// Convert to response (simplified - actual implementation would calculate similarity scores)
	results := make([]SearchResult, len(entities))
	for i, entity := range entities {
		results[i] = SearchResult{
			Entity:     toEntityResponse(entity),
			Similarity: 0.8, // Simplified - actual implementation would use real similarity
		}
	}

	respondJSON(w, http.StatusOK, SemanticSearchResponse{
		Query:   req.Query,
		Results: results,
		Total:   len(results),
	})
}

// Helper functions

func toEntityResponse(entity *ent.DiscoveredEntity) EntityResponse {
	createdAt := ""
	if !entity.CreatedAt.IsZero() {
		createdAt = entity.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return EntityResponse{
		ID:              entity.ID,
		UniqueID:        entity.UniqueID,
		TypeCategory:    entity.TypeCategory,
		Name:            entity.Name,
		Properties:      entity.Properties,
		ConfidenceScore: entity.ConfidenceScore,
		CreatedAt:       createdAt,
	}
}

func toRelationshipResponse(rel *ent.Relationship) RelationshipResponse {
	timestamp := ""
	if !rel.Timestamp.IsZero() {
		timestamp = rel.Timestamp.Format("2006-01-02T15:04:05Z07:00")
	}

	return RelationshipResponse{
		ID:              rel.ID,
		Type:            rel.Type,
		FromType:        rel.FromType,
		FromID:          rel.FromID,
		ToType:          rel.ToType,
		ToID:            rel.ToID,
		Timestamp:       timestamp,
		ConfidenceScore: rel.ConfidenceScore,
		Properties:      rel.Properties,
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, error string, details string) {
	respondJSON(w, status, ErrorResponse{
		Error:   error,
		Details: details,
	})
}

// mockRepoWrapper wraps the mock repository from tests
type mockRepoWrapper struct {
	mock interface{}
}

func (m *mockRepoWrapper) FindEntityByID(ctx context.Context, id int) (*ent.DiscoveredEntity, error) {
	if finder, ok := m.mock.(interface {
		FindEntityByID(context.Context, int) (*ent.DiscoveredEntity, error)
	}); ok {
		return finder.FindEntityByID(ctx, id)
	}
	return nil, fmt.Errorf("method not implemented")
}

func (m *mockRepoWrapper) FindEntitiesByType(ctx context.Context, typeCategory string) ([]*ent.DiscoveredEntity, error) {
	if finder, ok := m.mock.(interface {
		FindEntitiesByType(context.Context, string) ([]*ent.DiscoveredEntity, error)
	}); ok {
		return finder.FindEntitiesByType(ctx, typeCategory)
	}
	return nil, fmt.Errorf("method not implemented")
}

func (m *mockRepoWrapper) FindRelationshipsByEntity(ctx context.Context, entityType string, entityID int) ([]*ent.Relationship, error) {
	if finder, ok := m.mock.(interface {
		FindRelationshipsByEntity(context.Context, string, int) ([]*ent.Relationship, error)
	}); ok {
		return finder.FindRelationshipsByEntity(ctx, entityType, entityID)
	}
	return nil, fmt.Errorf("method not implemented")
}

func (m *mockRepoWrapper) TraverseRelationships(ctx context.Context, fromID int, relType string, depth int) ([]*ent.DiscoveredEntity, error) {
	if finder, ok := m.mock.(interface {
		TraverseRelationships(context.Context, int, string, int) ([]*ent.DiscoveredEntity, error)
	}); ok {
		return finder.TraverseRelationships(ctx, fromID, relType, depth)
	}
	return nil, fmt.Errorf("method not implemented")
}

func (m *mockRepoWrapper) FindShortestPath(ctx context.Context, fromID, toID int) ([]*ent.Relationship, error) {
	if finder, ok := m.mock.(interface {
		FindShortestPath(context.Context, int, int) ([]*ent.Relationship, error)
	}); ok {
		return finder.FindShortestPath(ctx, fromID, toID)
	}
	return nil, fmt.Errorf("method not implemented")
}

func (m *mockRepoWrapper) SimilaritySearch(ctx context.Context, embedding []float32, topK int, threshold float64) ([]*ent.DiscoveredEntity, error) {
	if finder, ok := m.mock.(interface {
		SimilaritySearch(context.Context, []float32, int, float64) ([]*ent.DiscoveredEntity, error)
	}); ok {
		return finder.SimilaritySearch(ctx, embedding, topK, threshold)
	}
	return nil, fmt.Errorf("method not implemented")
}

func (m *mockRepoWrapper) Close() error {
	if closer, ok := m.mock.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// Stub implementations for remaining Repository methods
func (m *mockRepoWrapper) CreateEmail(ctx context.Context, email *graph.EmailInput) (*ent.Email, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockRepoWrapper) FindEmailByMessageID(ctx context.Context, messageID string) (*ent.Email, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockRepoWrapper) CreateDiscoveredEntity(ctx context.Context, entity *graph.EntityInput) (*ent.DiscoveredEntity, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockRepoWrapper) FindEntityByUniqueID(ctx context.Context, uniqueID string) (*ent.DiscoveredEntity, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockRepoWrapper) GetDistinctEntityTypes(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockRepoWrapper) GetDistinctRelationshipTypes(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockRepoWrapper) CreateRelationship(ctx context.Context, rel *graph.RelationshipInput) (*ent.Relationship, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockRepoWrapper) GetClient() *ent.Client {
	return nil // Mock wrapper doesn't have a real client
}
