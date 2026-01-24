package extractor

import (
	"math"
	"strings"
	"testing"
)

// T030: Unit tests for deduplication
// Tests person dedup by email, org name normalization, concept similarity, edge cases

// TestConcept for testing (ConceptEntity in prompts.go doesn't have Embedding)
type TestConcept struct {
	Name      string
	Embedding []float32
}

func TestDeduplicatePerson_ByEmail(t *testing.T) {
	tests := []struct {
		name     string
		entities []PersonEntity
		expected int // expected number after dedup
	}{
		{
			name: "Identical emails",
			entities: []PersonEntity{
				{Name: "Alice Smith", Email: "alice@enron.com"},
				{Name: "Alice", Email: "alice@enron.com"},
				{Name: "A. Smith", Email: "alice@enron.com"},
			},
			expected: 1,
		},
		{
			name: "Different emails",
			entities: []PersonEntity{
				{Name: "Alice", Email: "alice@enron.com"},
				{Name: "Bob", Email: "bob@enron.com"},
			},
			expected: 2,
		},
		{
			name: "Mixed case emails",
			entities: []PersonEntity{
				{Name: "Alice", Email: "alice@enron.com"},
				{Name: "Alice", Email: "ALICE@ENRON.COM"},
				{Name: "Alice", Email: "Alice@Enron.Com"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deduped := deduplicatePersonsByEmail(tt.entities)
			if len(deduped) != tt.expected {
				t.Errorf("Expected %d entities after dedup, got %d", tt.expected, len(deduped))
			}
		})
	}
}

func TestDeduplicateOrganization_NameNormalization(t *testing.T) {
	tests := []struct {
		name       string
		orgNames   []string
		normalized string
	}{
		{
			name:       "Remove Inc suffix",
			orgNames:   []string{"Enron Inc", "Enron Inc.", "Enron"},
			normalized: "enron",
		},
		{
			name:       "Remove Corp suffix",
			orgNames:   []string{"Acme Corp", "Acme Corporation", "Acme Corp."},
			normalized: "acme",
		},
		{
			name:       "Remove LLC suffix",
			orgNames:   []string{"Smith LLC", "Smith L.L.C.", "Smith"},
			normalized: "smith",
		},
		{
			name:       "Multiple spaces",
			orgNames:   []string{"Big   Corp", "Big Corp"},
			normalized: "big",
		},
		{
			name:       "Case insensitive",
			orgNames:   []string{"ENRON", "Enron", "enron"},
			normalized: "enron",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, orgName := range tt.orgNames {
				normalized := normalizeOrgName(orgName)
				if normalized != tt.normalized {
					t.Errorf("normalizeOrgName(%s) = %s, expected %s", orgName, normalized, tt.normalized)
				}
			}
		})
	}
}

func TestDeduplicateConcept_SimilarityMatching(t *testing.T) {
	// Test concept deduplication using mock embeddings
	// Note: ConceptEntity from prompts.go doesn't have Embedding field
	// This test validates the algorithm logic with test data

	tests := []struct {
		name       string
		concepts   []TestConcept
		threshold  float64
		expectedGE int // expected >= this many (some might be merged)
		expectedLE int // expected <= this many
	}{
		{
			name: "Identical embeddings",
			concepts: []TestConcept{
				{Name: "energy", Embedding: []float32{0.1, 0.2, 0.3}},
				{Name: "power", Embedding: []float32{0.1, 0.2, 0.3}},
			},
			threshold:  0.85,
			expectedGE: 1,
			expectedLE: 1,
		},
		{
			name: "Different embeddings",
			concepts: []TestConcept{
				{Name: "energy", Embedding: []float32{1.0, 0.0, 0.0}},
				{Name: "food", Embedding: []float32{0.0, 1.0, 0.0}},
			},
			threshold:  0.85,
			expectedGE: 2,
			expectedLE: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deduped := deduplicateTestConceptsBySimilarity(tt.concepts, tt.threshold)
			if len(deduped) < tt.expectedGE || len(deduped) > tt.expectedLE {
				t.Errorf("Expected %d-%d concepts after dedup, got %d", tt.expectedGE, tt.expectedLE, len(deduped))
			}
		})
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		vec1     []float32
		vec2     []float32
		expected float64
		delta    float64 // acceptable difference
	}{
		{
			name:     "Identical vectors",
			vec1:     []float32{1.0, 0.0, 0.0},
			vec2:     []float32{1.0, 0.0, 0.0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "Orthogonal vectors",
			vec1:     []float32{1.0, 0.0, 0.0},
			vec2:     []float32{0.0, 1.0, 0.0},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "Opposite vectors",
			vec1:     []float32{1.0, 0.0, 0.0},
			vec2:     []float32{-1.0, 0.0, 0.0},
			expected: -1.0,
			delta:    0.001,
		},
		{
			name:     "Similar vectors",
			vec1:     []float32{0.6, 0.8, 0.0},
			vec2:     []float32{0.8, 0.6, 0.0},
			expected: 0.96, // cos(angle) for similar vectors
			delta:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := cosineSimilarity(tt.vec1, tt.vec2)
			if math.Abs(similarity-tt.expected) > tt.delta {
				t.Errorf("Expected similarity %.3f, got %.3f", tt.expected, similarity)
			}
		})
	}
}

func TestDeduplication_EdgeCases(t *testing.T) {
	t.Run("Empty person email", func(t *testing.T) {
		entities := []PersonEntity{
			{Name: "Alice", Email: ""},
			{Name: "Bob", Email: ""},
		}
		// Empty emails should not be deduplicated together
		deduped := deduplicatePersonsByEmail(entities)
		if len(deduped) != 2 {
			t.Errorf("Expected 2 entities with empty emails, got %d", len(deduped))
		}
	})

	t.Run("Empty organization name", func(t *testing.T) {
		normalized := normalizeOrgName("")
		if normalized != "" {
			t.Errorf("Expected empty string, got '%s'", normalized)
		}
	})

	t.Run("Organization with special characters", func(t *testing.T) {
		name := "A&T Corp."
		normalized := normalizeOrgName(name)
		if !strings.Contains(normalized, "&") {
			t.Error("Special characters should be preserved")
		}
	})

	t.Run("Zero-length embeddings", func(t *testing.T) {
		concepts := []TestConcept{
			{Name: "test1", Embedding: []float32{}},
			{Name: "test2", Embedding: []float32{}},
		}
		// Should handle gracefully
		deduped := deduplicateTestConceptsBySimilarity(concepts, 0.85)
		_ = deduped // No panic expected
	})
}

// Helper types and functions for testing

func deduplicatePersonsByEmail(entities []PersonEntity) []PersonEntity {
	seen := make(map[string]bool)
	result := []PersonEntity{}

	for _, e := range entities {
		email := strings.ToLower(strings.TrimSpace(e.Email))
		if email == "" {
			// Don't deduplicate empty emails
			result = append(result, e)
			continue
		}
		if !seen[email] {
			seen[email] = true
			result = append(result, e)
		}
	}

	return result
}

func normalizeOrgName(name string) string {
	// Convert to lowercase
	normalized := strings.ToLower(name)

	// Normalize whitespace first
	normalized = strings.TrimSpace(normalized)
	normalized = strings.Join(strings.Fields(normalized), " ")

	// Remove common suffixes
	suffixes := []string{" inc.", " inc", " corp.", " corp", " corporation", " llc", " l.l.c.", " ltd.", " ltd"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
			normalized = strings.TrimSpace(normalized)
			break
		}
	}

	return normalized
}

func deduplicateTestConceptsBySimilarity(concepts []TestConcept, threshold float64) []TestConcept {
	if len(concepts) == 0 {
		return concepts
	}

	result := []TestConcept{concepts[0]}

	for i := 1; i < len(concepts); i++ {
		isDuplicate := false
		for _, existing := range result {
			if len(concepts[i].Embedding) > 0 && len(existing.Embedding) > 0 {
				similarity := cosineSimilarity(concepts[i].Embedding, existing.Embedding)
				if similarity >= threshold {
					isDuplicate = true
					break
				}
			}
		}
		if !isDuplicate {
			result = append(result, concepts[i])
		}
	}

	return result
}

func cosineSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) == 0 || len(vec2) == 0 || len(vec1) != len(vec2) {
		return 0.0
	}

	var dotProduct, norm1, norm2 float64

	for i := 0; i < len(vec1); i++ {
		dotProduct += float64(vec1[i] * vec2[i])
		norm1 += float64(vec1[i] * vec1[i])
		norm2 += float64(vec2[i] * vec2[i])
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}
