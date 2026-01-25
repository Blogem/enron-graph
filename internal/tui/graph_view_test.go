package tui

import (
	"fmt"
	"strings"
	"testing"
)

// TestNodeFormatting tests that nodes are formatted correctly as [Type: Name]
func TestNodeFormatting(t *testing.T) {
	tests := []struct {
		name         string
		entityType   string
		entityName   string
		expectedType string
		expectedName string
	}{
		{
			name:         "person entity",
			entityType:   "person",
			entityName:   "Jeff Skilling",
			expectedType: "person",
			expectedName: "Jeff Skilling",
		},
		{
			name:         "organization entity",
			entityType:   "organization",
			entityName:   "Enron Corp",
			expectedType: "organization",
			expectedName: "Enron Corp",
		},
		{
			name:         "concept entity",
			entityType:   "concept",
			entityName:   "energy trading",
			expectedType: "concept",
			expectedName: "energy trading",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNode(tt.entityType, tt.entityName)
			if !strings.Contains(result, tt.expectedType) {
				t.Errorf("formatNode() result doesn't contain type %v", tt.expectedType)
			}
			if !strings.Contains(result, tt.expectedName) {
				t.Errorf("formatNode() result doesn't contain name %v", tt.expectedName)
			}
		})
	}
}

// TestEdgeFormatting tests that edges are formatted correctly as ---[REL_TYPE]-->
func TestEdgeFormatting(t *testing.T) {
	tests := []struct {
		name     string
		relType  string
		expected string
	}{
		{
			name:     "SENT relationship",
			relType:  "SENT",
			expected: "---[SENT]-->",
		},
		{
			name:     "RECEIVED relationship",
			relType:  "RECEIVED",
			expected: "---[RECEIVED]-->",
		},
		{
			name:     "MENTIONS relationship",
			relType:  "MENTIONS",
			expected: "---[MENTIONS]-->",
		},
		{
			name:     "COMMUNICATES_WITH relationship",
			relType:  "COMMUNICATES_WITH",
			expected: "---[COMMUNICATES_WITH]-->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEdge(tt.relType)
			if result != tt.expected {
				t.Errorf("formatEdge() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestLayoutCalculation tests that tree layout calculations work correctly
func TestLayoutCalculation(t *testing.T) {
	tests := []struct {
		name        string
		nodeCount   int
		wantRows    int
		wantColumns int
	}{
		{
			name:        "single node",
			nodeCount:   1,
			wantRows:    1,
			wantColumns: 1,
		},
		{
			name:        "three nodes",
			nodeCount:   3,
			wantRows:    2,
			wantColumns: 2,
		},
		{
			name:        "ten nodes",
			nodeCount:   10,
			wantRows:    3,
			wantColumns: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, cols := calculateTreeLayout(tt.nodeCount)
			if rows != tt.wantRows || cols != tt.wantColumns {
				t.Errorf("calculateTreeLayout(%d) = (%d, %d), want (%d, %d)",
					tt.nodeCount, rows, cols, tt.wantRows, tt.wantColumns)
			}
		})
	}
}

// TestNodeLimitEnforcement tests that rendering enforces max 50 nodes
func TestNodeLimitEnforcement(t *testing.T) {
	// Create mock graph with 100 nodes
	nodes := make([]GraphNode, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = GraphNode{
			ID:   i,
			Type: "person",
			Name: fmt.Sprintf("Person%d", i),
		}
	}

	graph := &Graph{
		Nodes: nodes,
		Edges: []GraphEdge{},
	}

	rendered := renderGraph(graph)

	// Count nodes in rendered output
	nodeCount := strings.Count(rendered, "[person:")
	if nodeCount > 50 {
		t.Errorf("renderGraph() rendered %d nodes, max should be 50", nodeCount)
	}
}

// TestColorCodingByType tests that different entity types get different colors
func TestColorCodingByType(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		wantColor  string
	}{
		{
			name:       "person should be blue",
			entityType: "person",
			wantColor:  "blue",
		},
		{
			name:       "organization should be green",
			entityType: "organization",
			wantColor:  "green",
		},
		{
			name:       "concept should be yellow",
			entityType: "concept",
			wantColor:  "yellow",
		},
		{
			name:       "email should be gray",
			entityType: "email",
			wantColor:  "gray",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := getColorForEntityType(tt.entityType)
			if color != tt.wantColor {
				t.Errorf("getColorForEntityType(%s) = %s, want %s",
					tt.entityType, color, tt.wantColor)
			}
		})
	}
}
