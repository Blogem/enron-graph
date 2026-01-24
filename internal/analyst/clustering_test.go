package analyst

import (
	"math"
	"testing"
)

// T076: Unit tests for embedding clustering
// Tests similarity grouping (cosine >0.85), cluster identification, type candidate extraction

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name               string
		vector1            []float32
		vector2            []float32
		expectedSimilarity float64
		threshold          float64
	}{
		{
			name:               "identical vectors",
			vector1:            []float32{1.0, 0.0, 0.0},
			vector2:            []float32{1.0, 0.0, 0.0},
			expectedSimilarity: 1.0,
			threshold:          0.85,
		},
		{
			name:               "opposite vectors",
			vector1:            []float32{1.0, 0.0, 0.0},
			vector2:            []float32{-1.0, 0.0, 0.0},
			expectedSimilarity: -1.0,
			threshold:          0.85,
		},
		{
			name:               "orthogonal vectors",
			vector1:            []float32{1.0, 0.0, 0.0},
			vector2:            []float32{0.0, 1.0, 0.0},
			expectedSimilarity: 0.0,
			threshold:          0.85,
		},
		{
			name:               "high similarity (above threshold)",
			vector1:            []float32{0.9, 0.1, 0.0},
			vector2:            []float32{0.95, 0.05, 0.0},
			expectedSimilarity: 0.999, // approximately
			threshold:          0.85,
		},
		{
			name:               "low similarity (below threshold)",
			vector1:            []float32{1.0, 0.0, 0.0},
			vector2:            []float32{0.5, 0.866, 0.0},
			expectedSimilarity: 0.5,
			threshold:          0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := CosineSimilarity(tt.vector1, tt.vector2)

			// Use larger epsilon for approximate comparisons
			if math.Abs(similarity-tt.expectedSimilarity) > 0.01 {
				t.Errorf("Expected similarity %.3f, got %.3f", tt.expectedSimilarity, similarity)
			}

			// Check threshold
			if tt.expectedSimilarity >= tt.threshold && similarity < tt.threshold {
				t.Errorf("Similarity %.3f should be >= threshold %.3f", similarity, tt.threshold)
			}
		})
	}
}

func TestGroupBySimilarity(t *testing.T) {
	tests := []struct {
		name             string
		entities         []EntityWithEmbedding
		threshold        float64
		expectedClusters int
		minClusterSize   int
	}{
		{
			name: "single cluster with high similarity",
			entities: []EntityWithEmbedding{
				{ID: 1, Type: "person", Embedding: []float32{1.0, 0.0, 0.0}},
				{ID: 2, Type: "person", Embedding: []float32{0.99, 0.01, 0.0}},
				{ID: 3, Type: "person", Embedding: []float32{0.98, 0.02, 0.0}},
			},
			threshold:        0.85,
			expectedClusters: 1,
			minClusterSize:   3,
		},
		{
			name: "two distinct clusters",
			entities: []EntityWithEmbedding{
				{ID: 1, Type: "person", Embedding: []float32{1.0, 0.0, 0.0}},
				{ID: 2, Type: "person", Embedding: []float32{0.99, 0.01, 0.0}},
				{ID: 3, Type: "organization", Embedding: []float32{0.0, 1.0, 0.0}},
				{ID: 4, Type: "organization", Embedding: []float32{0.01, 0.99, 0.0}},
			},
			threshold:        0.85,
			expectedClusters: 2,
			minClusterSize:   2,
		},
		{
			name: "outliers below threshold",
			entities: []EntityWithEmbedding{
				{ID: 1, Type: "concept", Embedding: []float32{1.0, 0.0, 0.0}},
				{ID: 2, Type: "concept", Embedding: []float32{0.0, 1.0, 0.0}},
				{ID: 3, Type: "concept", Embedding: []float32{0.0, 0.0, 1.0}},
			},
			threshold:        0.85,
			expectedClusters: 0, // No clusters if all entities are too different
			minClusterSize:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusters := GroupBySimilarity(tt.entities, tt.threshold)

			// Filter clusters by minimum size
			validClusters := []Cluster{}
			for _, cluster := range clusters {
				if len(cluster.Members) >= tt.minClusterSize {
					validClusters = append(validClusters, cluster)
				}
			}

			if len(validClusters) != tt.expectedClusters {
				t.Errorf("Expected %d clusters (min size %d), got %d", tt.expectedClusters, tt.minClusterSize, len(validClusters))
			}

			// Verify cluster similarity
			for i, cluster := range validClusters {
				for j := 0; j < len(cluster.Members)-1; j++ {
					for k := j + 1; k < len(cluster.Members); k++ {
						member1 := cluster.Members[j]
						member2 := cluster.Members[k]
						similarity := CosineSimilarity(member1.Embedding, member2.Embedding)
						if similarity < tt.threshold {
							t.Errorf("Cluster %d: members %d and %d have similarity %.3f < threshold %.3f",
								i, member1.ID, member2.ID, similarity, tt.threshold)
						}
					}
				}
			}
		})
	}
}

func TestIdentifyClusters(t *testing.T) {
	tests := []struct {
		name             string
		entities         []EntityWithEmbedding
		threshold        float64
		expectedClusters []ClusterInfo
	}{
		{
			name: "identify person cluster",
			entities: []EntityWithEmbedding{
				{ID: 1, Type: "person", Name: "Alice", Embedding: []float32{1.0, 0.0, 0.0}},
				{ID: 2, Type: "person", Name: "Bob", Embedding: []float32{0.99, 0.01, 0.0}},
				{ID: 3, Type: "person", Name: "Charlie", Embedding: []float32{0.98, 0.02, 0.0}},
			},
			threshold: 0.85,
			expectedClusters: []ClusterInfo{
				{Type: "person", Size: 3},
			},
		},
		{
			name: "identify multiple clusters by type",
			entities: []EntityWithEmbedding{
				{ID: 1, Type: "person", Name: "Alice", Embedding: []float32{1.0, 0.0, 0.0}},
				{ID: 2, Type: "person", Name: "Bob", Embedding: []float32{0.99, 0.01, 0.0}},
				{ID: 3, Type: "organization", Name: "Enron", Embedding: []float32{0.0, 1.0, 0.0}},
				{ID: 4, Type: "organization", Name: "Dynegy", Embedding: []float32{0.01, 0.99, 0.0}},
			},
			threshold: 0.85,
			expectedClusters: []ClusterInfo{
				{Type: "person", Size: 2},
				{Type: "organization", Size: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusters := IdentifyClusters(tt.entities, tt.threshold)

			if len(clusters) != len(tt.expectedClusters) {
				t.Errorf("Expected %d clusters, got %d", len(tt.expectedClusters), len(clusters))
			}

			// Create map for easier comparison
			clusterMap := make(map[string]int)
			for _, cluster := range clusters {
				clusterMap[cluster.Type] = cluster.Size
			}

			for _, expected := range tt.expectedClusters {
				size, ok := clusterMap[expected.Type]
				if !ok {
					t.Errorf("Expected cluster type '%s' not found", expected.Type)
					continue
				}
				if size != expected.Size {
					t.Errorf("Cluster '%s': expected size %d, got %d", expected.Type, expected.Size, size)
				}
			}
		})
	}
}

func TestExtractTypeCandidates(t *testing.T) {
	tests := []struct {
		name               string
		clusters           []Cluster
		minSize            int
		expectedCandidates []string
	}{
		{
			name: "extract candidates above minimum size",
			clusters: []Cluster{
				{Type: "person", Members: make([]EntityWithEmbedding, 100)},
				{Type: "organization", Members: make([]EntityWithEmbedding, 75)},
				{Type: "concept", Members: make([]EntityWithEmbedding, 30)},
			},
			minSize:            50,
			expectedCandidates: []string{"person", "organization"},
		},
		{
			name: "no candidates below threshold",
			clusters: []Cluster{
				{Type: "concept", Members: make([]EntityWithEmbedding, 20)},
				{Type: "event", Members: make([]EntityWithEmbedding, 15)},
			},
			minSize:            50,
			expectedCandidates: []string{},
		},
		{
			name: "all candidates above threshold",
			clusters: []Cluster{
				{Type: "person", Members: make([]EntityWithEmbedding, 150)},
				{Type: "organization", Members: make([]EntityWithEmbedding, 100)},
				{Type: "concept", Members: make([]EntityWithEmbedding, 75)},
			},
			minSize:            50,
			expectedCandidates: []string{"person", "organization", "concept"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize cluster members to match expected size
			for i := range tt.clusters {
				if len(tt.clusters[i].Members) == 0 {
					// Members were created with make(), so length is already set
					continue
				}
			}

			candidates := ExtractTypeCandidates(tt.clusters, tt.minSize)

			if len(candidates) != len(tt.expectedCandidates) {
				t.Errorf("Expected %d candidates, got %d", len(tt.expectedCandidates), len(candidates))
			}

			candidateMap := make(map[string]bool)
			for _, candidate := range candidates {
				candidateMap[candidate] = true
			}

			for _, expected := range tt.expectedCandidates {
				if !candidateMap[expected] {
					t.Errorf("Expected candidate '%s' not found", expected)
				}
			}
		})
	}
}
