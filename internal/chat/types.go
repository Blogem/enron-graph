package chat

import (
	"context"
	"time"
)

// QueryType represents the type of query pattern matched
type QueryType string

const (
	QueryTypeEntityLookup  QueryType = "entity_lookup"
	QueryTypeRelationship  QueryType = "relationship"
	QueryTypePathFinding   QueryType = "path_finding"
	QueryTypeConceptSearch QueryType = "concept_search"
	QueryTypeAggregation   QueryType = "aggregation"
	QueryTypeAmbiguous     QueryType = "ambiguous"
	QueryTypeUnknown       QueryType = "unknown"
)

// MatchResult represents the result of pattern matching
type MatchResult struct {
	Type      QueryType
	Args      map[string]string
	Ambiguous bool
	Options   []string
}

// PatternMatcher interface for query pattern matching
type PatternMatcher interface {
	Match(query string) (*MatchResult, error)
}

// HistoryEntry represents a conversation history entry
type HistoryEntry struct {
	Query     string
	Response  string
	Timestamp time.Time
}

// TrackedEntity represents an entity mentioned in conversation
type TrackedEntity struct {
	Name      string
	Type      string
	ID        int
	Timestamp time.Time
}

// Entity represents a graph entity
type Entity struct {
	ID         int
	Name       string
	Type       string
	Properties map[string]interface{}
}

// PathNode represents a node in a path
type PathNode struct {
	Entity       *Entity
	Relationship string
}

// Context interface for conversation context management
type Context interface {
	AddQuery(query, response string)
	GetHistory() []HistoryEntry
	TrackEntity(name, entityType string, id int)
	GetTrackedEntities() map[string]TrackedEntity
	ResolvePronoun(query string) (*TrackedEntity, bool)
	BuildPromptContext(query string) string
	Clear()
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	GetLastMentionedEntity() (*TrackedEntity, bool)
}

// LLMClient interface for LLM operations
type LLMClient interface {
	GenerateCompletion(ctx context.Context, prompt string) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// Repository interface for graph operations
type Repository interface {
	FindEntityByName(name string) (*Entity, error)
	TraverseRelationships(entityID int, relType string) ([]*Entity, error)
	FindShortestPath(sourceID, targetID int) ([]*PathNode, error)
	SimilaritySearch(embedding []float32, limit int) ([]*Entity, error)
	CountRelationships(entityID int, relType string) (int, error)
}

// Handler interface for chat query processing
type Handler interface {
	ProcessQuery(ctx context.Context, query string, chatContext Context) (string, error)
}

// ResponseFormatter interface for formatting responses
type ResponseFormatter interface {
	FormatEntities(entities []*Entity) string
	FormatPath(path []*PathNode) string
	FormatCount(count int, description string) string
}
