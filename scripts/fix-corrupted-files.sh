#!/bin/bash
# Script to fix corrupted files from file creation tool issues

cd "$(dirname "$0")/.."

echo "Fixing internal/explorer/models.go..."
cat > internal/explorer/models.go << 'ENDFILE'
package explorer

type GraphNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Category   string                 `json:"category"`
	Properties map[string]interface{} `json:"properties"`
	IsGhost    bool                   `json:"is_ghost"`
	Degree     int                    `json:"degree,omitempty"`
}

type GraphEdge struct {
	Source     string                 `json:"source"`
	Target     string                 `json:"target"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type PropertyDefinition struct {
	Name         string   `json:"name"`
	Type         string   `json:"data_type"`
	SampleValues []string `json:"sample_value,omitempty"`
	Nullable     bool     `json:"nullable"`
}

type SchemaType struct {
	Name          string                 `json:"name"`
	Category      string                 `json:"category"`
	Count         int64                  `json:"count"`
	Properties    []PropertyDefinition   `json:"properties"`
	IsPromoted    bool                   `json:"is_promoted"`
	Relationships []string               `json:"relationships,omitempty"`
}

type GraphResponse struct {
	Nodes      []GraphNode `json:"nodes"`
	Edges      []GraphEdge `json:"edges"`
	TotalNodes int         `json:"total_nodes"`
	HasMore    bool        `json:"has_more"`
}

type RelationshipsResponse struct {
	Nodes      []GraphNode `json:"nodes"`
	Edges      []GraphEdge `json:"edges"`
	TotalCount int         `json:"total_count"`
	HasMore    bool        `json:"has_more"`
	Offset     int         `json:"offset"`
}

type SchemaResponse struct {
	PromotedTypes   []SchemaType `json:"promoted_types"`
	DiscoveredTypes []SchemaType `json:"discovered_types"`
	TotalEntities   int          `json:"total_entities"`
}

type NodeFilter struct {
	Types       []string `json:"types,omitempty"`
	Category    string   `json:"category,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
	Limit       int      `json:"limit,omitempty"`
}
ENDFILE

echo "✅ Fixed models.go"

echo "Verifying compilation..."
go build ./internal/explorer/... && echo "✅ internal/explorer compiles" || echo "❌ Compilation failed"

go build ./cmd/explorer/... && echo "✅ cmd/explorer compiles" || echo "❌ Compilation failed"

echo ""
echo "Done! Run this script anytime to restore corrupted files."
