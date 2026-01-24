package analyst

import (
	"context"
	"math"

	"github.com/Blogem/enron-graph/ent"
)

// T081: Embedding clustering implementation
// Group entities by vector similarity (cosine >0.85)

// EntityWithEmbedding represents an entity with its embedding vector
type EntityWithEmbedding struct {
	ID        int
	Type      string
	Name      string
	Embedding []float32
}

// Cluster represents a group of similar entities
type Cluster struct {
	Type    string
	Members []EntityWithEmbedding
}

// ClusterInfo contains summary information about a cluster
type ClusterInfo struct {
	Type string
	Size int
}

// CosineSimilarity calculates the cosine similarity between two vectors
func CosineSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) || len(vec1) == 0 {
		return 0.0
	}

	var dotProduct, mag1, mag2 float64

	for i := 0; i < len(vec1); i++ {
		dotProduct += float64(vec1[i]) * float64(vec2[i])
		mag1 += float64(vec1[i]) * float64(vec1[i])
		mag2 += float64(vec2[i]) * float64(vec2[i])
	}

	mag1 = math.Sqrt(mag1)
	mag2 = math.Sqrt(mag2)

	if mag1 == 0 || mag2 == 0 {
		return 0.0
	}

	return dotProduct / (mag1 * mag2)
}

// GroupBySimilarity groups entities into clusters based on embedding similarity
func GroupBySimilarity(entities []EntityWithEmbedding, threshold float64) []Cluster {
	if len(entities) == 0 {
		return []Cluster{}
	}

	clusters := []Cluster{}
	assigned := make(map[int]bool)

	for i, entity := range entities {
		if assigned[i] {
			continue
		}

		// Start new cluster
		cluster := Cluster{
			Type:    entity.Type,
			Members: []EntityWithEmbedding{entity},
		}
		assigned[i] = true

		// Find similar entities
		for j := i + 1; j < len(entities); j++ {
			if assigned[j] {
				continue
			}

			similarity := CosineSimilarity(entity.Embedding, entities[j].Embedding)
			if similarity >= threshold {
				cluster.Members = append(cluster.Members, entities[j])
				assigned[j] = true
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

// IdentifyClusters groups entities by type and similarity, returning cluster info
func IdentifyClusters(entities []EntityWithEmbedding, threshold float64) []ClusterInfo {
	// Group by type first
	typeGroups := make(map[string][]EntityWithEmbedding)
	for _, entity := range entities {
		typeGroups[entity.Type] = append(typeGroups[entity.Type], entity)
	}

	// Cluster within each type
	result := []ClusterInfo{}
	for entityType, group := range typeGroups {
		clusters := GroupBySimilarity(group, threshold)

		// Aggregate cluster sizes
		totalMembers := 0
		for _, cluster := range clusters {
			totalMembers += len(cluster.Members)
		}

		if totalMembers > 0 {
			result = append(result, ClusterInfo{
				Type: entityType,
				Size: totalMembers,
			})
		}
	}

	return result
}

// ExtractTypeCandidates extracts type names from clusters that meet minimum size threshold
func ExtractTypeCandidates(clusters []Cluster, minSize int) []string {
	candidates := []string{}

	for _, cluster := range clusters {
		if len(cluster.Members) >= minSize {
			candidates = append(candidates, cluster.Type)
		}
	}

	return candidates
}

// ClusterEntities performs clustering on discovered entities
func ClusterEntities(ctx context.Context, client *ent.Client, threshold float64) ([]Cluster, error) {
	// Query entities with embeddings
	entities, err := client.DiscoveredEntity.
		Query().
		Where().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to EntityWithEmbedding
	ewes := make([]EntityWithEmbedding, 0, len(entities))
	for _, entity := range entities {
		if len(entity.Embedding) > 0 {
			ewes = append(ewes, EntityWithEmbedding{
				ID:        entity.ID,
				Type:      entity.TypeCategory,
				Name:      entity.Name,
				Embedding: entity.Embedding,
			})
		}
	}

	return GroupBySimilarity(ewes, threshold), nil
}
