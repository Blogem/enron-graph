package tui

import (
	"testing"
)

// BenchmarkGraphRendering500Nodes verifies T114: SC-007 - <3s render time for 500 nodes
func BenchmarkGraphRendering500Nodes(b *testing.B) {
	// Create graph with 500 nodes
	nodes := make([]GraphNode, 500)
	for i := 0; i < 500; i++ {
		nodes[i] = GraphNode{
			ID:   i,
			Type: "person",
			Name: "Person " + string(rune('A'+(i%26))),
		}
	}

	// Create edges (each node connects to next 3 nodes)
	edges := []GraphEdge{}
	for i := 0; i < 497; i++ {
		for j := 1; j <= 3; j++ {
			edges = append(edges, GraphEdge{
				FromID:  i,
				ToID:    i + j,
				RelType: "CONNECTS_TO",
			})
		}
	}

	graph := &Graph{
		Nodes: nodes,
		Edges: edges,
	}

	// Benchmark rendering
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderGraph(graph)
	}
}

// BenchmarkGraphRenderingWithLimit tests that 50-node limit keeps performance fast
func BenchmarkGraphRenderingWithLimit(b *testing.B) {
	// Create graph with 1000 nodes (but only 50 will be rendered)
	nodes := make([]GraphNode, 1000)
	for i := 0; i < 1000; i++ {
		nodes[i] = GraphNode{
			ID:   i,
			Type: "person",
			Name: "Person " + string(rune('A'+(i%26))),
		}
	}

	graph := &Graph{
		Nodes: nodes,
		Edges: []GraphEdge{},
	}

	// Benchmark rendering (should be limited to 50 nodes internally)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderGraph(graph)
	}
}

// BenchmarkEntityListRendering benchmarks entity list table rendering
func BenchmarkEntityListRendering(b *testing.B) {
	// Create 1000 test entities
	entities := make([]Entity, 1000)
	for i := 0; i < 1000; i++ {
		entities[i] = Entity{
			ID:         i,
			Type:       "person",
			Name:       "Person " + string(rune('A'+(i%26))),
			Confidence: 0.95,
		}
	}

	entityList := NewEntityListModel()
	entityList.entities = entities

	// Benchmark rendering
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = entityList.View(120, 40)
	}
}

// BenchmarkEntityFiltering benchmarks filter performance
func BenchmarkEntityFiltering(b *testing.B) {
	// Create 10000 test entities with mixed types
	entities := make([]Entity, 10000)
	types := []string{"person", "organization", "concept", "email"}
	for i := 0; i < 10000; i++ {
		entities[i] = Entity{
			ID:         i,
			Type:       types[i%4],
			Name:       "Entity " + string(rune('A'+(i%26))),
			Confidence: 0.95,
		}
	}

	// Benchmark filtering
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filterEntitiesByType(entities, "person")
	}
}

// BenchmarkEntitySearch benchmarks search performance
func BenchmarkEntitySearch(b *testing.B) {
	// Create 10000 test entities
	entities := make([]Entity, 10000)
	for i := 0; i < 10000; i++ {
		entities[i] = Entity{
			ID:         i,
			Type:       "person",
			Name:       "Person " + string(rune('A'+(i%26))),
			Confidence: 0.95,
		}
	}

	// Benchmark search
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = searchEntitiesByName(entities, "Person A")
	}
}

// BenchmarkFullTUIRender benchmarks complete TUI rendering
func BenchmarkFullTUIRender(b *testing.B) {
	// Create model with realistic data
	model := NewModel()

	entities := make([]Entity, 1000)
	for i := 0; i < 1000; i++ {
		entities[i] = Entity{
			ID:         i,
			Type:       "person",
			Name:       "Person " + string(rune('A'+(i%26))),
			Confidence: 0.95,
		}
	}
	model.LoadEntities(entities)

	// Set window size
	model.width = 120
	model.height = 40

	// Benchmark full rendering
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}
