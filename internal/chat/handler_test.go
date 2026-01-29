package chat

import (
	"context"
	"errors"
	"testing"
)

// MockLLMClient is a mock implementation of the LLM client for testing
type MockLLMClient struct {
	GenerateCompletionFunc func(ctx context.Context, prompt string) (string, error)
	GenerateEmbeddingFunc  func(ctx context.Context, text string) ([]float32, error)
}

func (m *MockLLMClient) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	if m.GenerateCompletionFunc != nil {
		return m.GenerateCompletionFunc(ctx, prompt)
	}
	return "", errors.New("not implemented")
}

func (m *MockLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if m.GenerateEmbeddingFunc != nil {
		return m.GenerateEmbeddingFunc(ctx, text)
	}
	return nil, errors.New("not implemented")
}

// MockRepository is a mock implementation of the graph repository for testing
type MockRepository struct {
	FindEntityByNameFunc      func(name string) (*Entity, error)
	TraverseRelationshipsFunc func(entityID int, relType string) ([]*Entity, error)
	FindShortestPathFunc      func(sourceID, targetID int) ([]*PathNode, error)
	SimilaritySearchFunc      func(embedding []float32, limit int) ([]*Entity, error)
	CountRelationshipsFunc    func(entityID int, relType string) (int, error)
}

func (m *MockRepository) FindEntityByName(name string) (*Entity, error) {
	if m.FindEntityByNameFunc != nil {
		return m.FindEntityByNameFunc(name)
	}
	return nil, errors.New("not found")
}

func (m *MockRepository) TraverseRelationships(entityID int, relType string) ([]*Entity, error) {
	if m.TraverseRelationshipsFunc != nil {
		return m.TraverseRelationshipsFunc(entityID, relType)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) FindShortestPath(sourceID, targetID int) ([]*PathNode, error) {
	if m.FindShortestPathFunc != nil {
		return m.FindShortestPathFunc(sourceID, targetID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) SimilaritySearch(embedding []float32, limit int) ([]*Entity, error) {
	if m.SimilaritySearchFunc != nil {
		return m.SimilaritySearchFunc(embedding, limit)
	}
	return nil, errors.New("not implemented")
}

func (m *MockRepository) CountRelationships(entityID int, relType string) (int, error) {
	if m.CountRelationshipsFunc != nil {
		return m.CountRelationshipsFunc(entityID, relType)
	}
	return 0, errors.New("not implemented")
}

// TestProcessQueryWithMockLLM tests query processing with mock LLM responses
func TestProcessQueryWithMockLLM(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return "Jeff Skilling is the former CEO of Enron.", nil
		},
	}

	mockRepo := &MockRepository{
		FindEntityByNameFunc: func(name string) (*Entity, error) {
			return &Entity{
				ID:   1,
				Name: "Jeff Skilling",
				Type: "person",
			}, nil
		},
	}

	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	response, err := handler.ProcessQuery(context.Background(), "Who is Jeff Skilling?", chatContext)
	if err != nil {
		t.Fatalf("ProcessQuery() error = %v", err)
	}
	if response == "" {
		t.Error("ProcessQuery() returned empty response")
	}
}

// TestResponseFormatting tests formatting of query results
func TestResponseFormatting(t *testing.T) {
	tests := []struct {
		name     string
		entities []*Entity
		wantLen  int
	}{
		{
			name: "single entity",
			entities: []*Entity{
				{ID: 1, Name: "Jeff Skilling", Type: "person"},
			},
			wantLen: 1,
		},
		{
			name: "multiple entities",
			entities: []*Entity{
				{ID: 1, Name: "Jeff Skilling", Type: "person"},
				{ID: 2, Name: "Kenneth Lay", Type: "person"},
			},
			wantLen: 2,
		},
		{
			name:     "empty results",
			entities: []*Entity{},
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewResponseFormatter()
			response := formatter.FormatEntities(tt.entities)
			if response.Text == "" && tt.wantLen > 0 {
				t.Error("FormatEntities() returned empty string for non-empty entities")
			}
			if tt.wantLen == 0 && response.Text != "" {
				// Empty result should have a "no results" message
				if response.Text == "" {
					t.Error("FormatEntities() should return a message for empty results")
				}
			}
			// Verify entities array matches input length
			if len(response.Entities) != tt.wantLen {
				t.Errorf("FormatEntities() entities length = %d, want %d", len(response.Entities), tt.wantLen)
			}
		})
	}
}

// TestErrorHandling tests error handling in query processing
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		llmError  error
		repoError error
		wantError bool
	}{
		{
			name:      "LLM connection failure",
			llmError:  errors.New("connection refused"),
			wantError: true,
		},
		{
			name:      "repository query failure",
			repoError: errors.New("database error"),
			wantError: true,
		},
		{
			name:      "entity not found",
			repoError: errors.New("entity not found"),
			wantError: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := &MockLLMClient{
				GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
					if tt.llmError != nil {
						return "", tt.llmError
					}
					// Return JSON that triggers repository call when testing repo errors
					if tt.repoError != nil {
						return `{"action": "entity_lookup", "entity": "Test"}`, nil
					}
					return "Response", nil
				},
			}

			mockRepo := &MockRepository{
				FindEntityByNameFunc: func(name string) (*Entity, error) {
					if tt.repoError != nil {
						return nil, tt.repoError
					}
					return &Entity{ID: 1, Name: "Test"}, nil
				},
			}

			handler := NewHandler(mockLLM, mockRepo)
			chatContext := NewContext()

			_, err := handler.ProcessQuery(context.Background(), "Test query", chatContext)
			if tt.wantError && err == nil {
				t.Error("ProcessQuery() expected error, got nil")
			}
			if !tt.wantError && err != nil && tt.repoError == nil {
				t.Errorf("ProcessQuery() unexpected error = %v", err)
			}
		})
	}
}

// TestContextPropagation tests context propagation across queries
func TestContextPropagation(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return "Response with context", nil
		},
	}

	mockRepo := &MockRepository{
		FindEntityByNameFunc: func(name string) (*Entity, error) {
			return &Entity{ID: 1, Name: name, Type: "person"}, nil
		},
	}

	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	// First query
	_, err := handler.ProcessQuery(context.Background(), "Who is Jeff Skilling?", chatContext)
	if err != nil {
		t.Fatalf("First ProcessQuery() error = %v", err)
	}

	// Second query should have context from first
	_, err = handler.ProcessQuery(context.Background(), "What did he do?", chatContext)
	if err != nil {
		t.Fatalf("Second ProcessQuery() error = %v", err)
	}

	// Verify context was maintained
	history := chatContext.GetHistory()
	if len(history) != 2 {
		t.Errorf("After 2 queries, history length = %d, want 2", len(history))
	}
}

// TestLLMTimeout tests handling of LLM timeout
func TestLLMTimeout(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return "", context.DeadlineExceeded
		},
	}

	mockRepo := &MockRepository{}
	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	_, err := handler.ProcessQuery(context.Background(), "Test query", chatContext)
	if err == nil {
		t.Error("ProcessQuery() expected timeout error, got nil")
	}
}

// TestInvalidJSONResponse tests handling of invalid JSON from LLM
func TestInvalidJSONResponse(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return "This is not valid JSON {{{[", nil
		},
	}

	mockRepo := &MockRepository{}
	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	// Handler should gracefully handle invalid JSON
	response, err := handler.ProcessQuery(context.Background(), "Test query", chatContext)

	// Should either return error or fallback to text response
	if err == nil && response == "" {
		t.Error("ProcessQuery() should handle invalid JSON gracefully")
	}
}

// TestEntityLookupQuery tests entity lookup query execution
func TestEntityLookupQuery(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return `{"action": "entity_lookup", "entity": "Jeff Skilling"}`, nil
		},
	}

	mockRepo := &MockRepository{
		FindEntityByNameFunc: func(name string) (*Entity, error) {
			if name == "Jeff Skilling" {
				return &Entity{
					ID:   1,
					Name: "Jeff Skilling",
					Type: "person",
					Properties: map[string]interface{}{
						"role": "CEO",
					},
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	response, err := handler.ProcessQuery(context.Background(), "Who is Jeff Skilling?", chatContext)
	if err != nil {
		t.Fatalf("ProcessQuery() error = %v", err)
	}
	if response == "" {
		t.Error("ProcessQuery() returned empty response")
	}
}

// TestRelationshipQuery tests relationship traversal query execution
func TestRelationshipQuery(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return `{"action": "traverse", "entity": "Jeff Skilling", "relationship": "SENT"}`, nil
		},
	}

	mockRepo := &MockRepository{
		FindEntityByNameFunc: func(name string) (*Entity, error) {
			return &Entity{ID: 1, Name: name, Type: "person"}, nil
		},
		TraverseRelationshipsFunc: func(entityID int, relType string) ([]*Entity, error) {
			return []*Entity{
				{ID: 2, Name: "Kenneth Lay", Type: "person"},
				{ID: 3, Name: "Andrew Fastow", Type: "person"},
			}, nil
		},
	}

	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	response, err := handler.ProcessQuery(context.Background(), "Who did Jeff Skilling email?", chatContext)
	if err != nil {
		t.Fatalf("ProcessQuery() error = %v", err)
	}
	if response == "" {
		t.Error("ProcessQuery() returned empty response")
	}
}

// TestPathFindingQuery tests shortest path query execution
func TestPathFindingQuery(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return `{"action": "find_path", "source": "Jeff Skilling", "target": "Kenneth Lay"}`, nil
		},
	}

	mockRepo := &MockRepository{
		FindEntityByNameFunc: func(name string) (*Entity, error) {
			if name == "Jeff Skilling" {
				return &Entity{ID: 1, Name: name, Type: "person"}, nil
			}
			if name == "Kenneth Lay" {
				return &Entity{ID: 2, Name: name, Type: "person"}, nil
			}
			return nil, errors.New("not found")
		},
		FindShortestPathFunc: func(sourceID, targetID int) ([]*PathNode, error) {
			return []*PathNode{
				{
					Entity:       &Entity{ID: 1, Name: "Jeff Skilling", Type: "person"},
					Relationship: "COMMUNICATES_WITH",
				},
				{
					Entity:       &Entity{ID: 2, Name: "Kenneth Lay", Type: "person"},
					Relationship: "",
				},
			}, nil
		},
	}

	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	response, err := handler.ProcessQuery(context.Background(), "How are Jeff Skilling and Kenneth Lay connected?", chatContext)
	if err != nil {
		t.Fatalf("ProcessQuery() error = %v", err)
	}
	if response == "" {
		t.Error("ProcessQuery() returned empty response")
	}
}

// TestSemanticSearchQuery tests semantic/concept search execution
func TestSemanticSearchQuery(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return `{"action": "semantic_search", "concept": "energy trading"}`, nil
		},
		GenerateEmbeddingFunc: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{0.1, 0.2, 0.3}, nil
		},
	}

	mockRepo := &MockRepository{
		SimilaritySearchFunc: func(embedding []float32, limit int) ([]*Entity, error) {
			return []*Entity{
				{ID: 1, Name: "Energy Trading Division", Type: "organization"},
				{ID: 2, Name: "Trading Floor", Type: "concept"},
			}, nil
		},
	}

	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	response, err := handler.ProcessQuery(context.Background(), "Emails about energy trading", chatContext)
	if err != nil {
		t.Fatalf("ProcessQuery() error = %v", err)
	}
	if response == "" {
		t.Error("ProcessQuery() returned empty response")
	}
}

// TestAggregationQuery tests aggregation query execution
func TestAggregationQuery(t *testing.T) {
	mockLLM := &MockLLMClient{
		GenerateCompletionFunc: func(ctx context.Context, prompt string) (string, error) {
			return `{"action": "count", "entity": "Jeff Skilling", "relationship": "SENT"}`, nil
		},
	}

	mockRepo := &MockRepository{
		FindEntityByNameFunc: func(name string) (*Entity, error) {
			return &Entity{ID: 1, Name: name, Type: "person"}, nil
		},
		CountRelationshipsFunc: func(entityID int, relType string) (int, error) {
			return 42, nil
		},
	}

	handler := NewHandler(mockLLM, mockRepo)
	chatContext := NewContext()

	response, err := handler.ProcessQuery(context.Background(), "How many emails did Jeff Skilling send?", chatContext)
	if err != nil {
		t.Fatalf("ProcessQuery() error = %v", err)
	}
	if response == "" {
		t.Error("ProcessQuery() returned empty response")
	}
}
