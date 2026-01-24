package analyst

import (
	"testing"
)

// T077: Unit tests for candidate ranking
// Tests scoring formula (0.4*freq + 0.3*density + 0.3*consistency), threshold application, sorting

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name          string
		frequency     float64
		density       float64
		consistency   float64
		expectedScore float64
	}{
		{
			name:          "high quality candidate",
			frequency:     100.0,
			density:       50.0,
			consistency:   0.95,
			expectedScore: 100.0*0.4 + 50.0*0.3 + 0.95*0.3, // 40 + 15 + 0.285 = 55.285
		},
		{
			name:          "medium quality candidate",
			frequency:     50.0,
			density:       25.0,
			consistency:   0.70,
			expectedScore: 50.0*0.4 + 25.0*0.3 + 0.70*0.3, // 20 + 7.5 + 0.21 = 27.71
		},
		{
			name:          "low quality candidate",
			frequency:     10.0,
			density:       5.0,
			consistency:   0.50,
			expectedScore: 10.0*0.4 + 5.0*0.3 + 0.50*0.3, // 4 + 1.5 + 0.15 = 5.65
		},
		{
			name:          "perfect candidate",
			frequency:     200.0,
			density:       100.0,
			consistency:   1.0,
			expectedScore: 200.0*0.4 + 100.0*0.3 + 1.0*0.3, // 80 + 30 + 0.3 = 110.3
		},
		{
			name:          "zero values",
			frequency:     0.0,
			density:       0.0,
			consistency:   0.0,
			expectedScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateScore(tt.frequency, tt.density, tt.consistency)

			// Use epsilon for float comparison
			if abs(score-tt.expectedScore) > 0.01 {
				t.Errorf("Expected score %.3f, got %.3f", tt.expectedScore, score)
			}
		})
	}
}

func TestApplyThresholds(t *testing.T) {
	tests := []struct {
		name               string
		candidates         []TypeCandidate
		minOccurrences     int
		minConsistency     float64
		expectedCandidates []string
	}{
		{
			name: "filter by minimum occurrences",
			candidates: []TypeCandidate{
				{Type: "person", Frequency: 100, Consistency: 0.85},
				{Type: "organization", Frequency: 75, Consistency: 0.80},
				{Type: "concept", Frequency: 30, Consistency: 0.75}, // below threshold
			},
			minOccurrences:     50,
			minConsistency:     0.70,
			expectedCandidates: []string{"person", "organization"},
		},
		{
			name: "filter by minimum consistency",
			candidates: []TypeCandidate{
				{Type: "person", Frequency: 100, Consistency: 0.85},
				{Type: "organization", Frequency: 75, Consistency: 0.65}, // below threshold
				{Type: "concept", Frequency: 60, Consistency: 0.75},
			},
			minOccurrences:     50,
			minConsistency:     0.70,
			expectedCandidates: []string{"person", "concept"},
		},
		{
			name: "filter by both thresholds",
			candidates: []TypeCandidate{
				{Type: "person", Frequency: 100, Consistency: 0.85},
				{Type: "organization", Frequency: 40, Consistency: 0.90}, // low frequency
				{Type: "concept", Frequency: 60, Consistency: 0.65},      // low consistency
			},
			minOccurrences:     50,
			minConsistency:     0.70,
			expectedCandidates: []string{"person"},
		},
		{
			name: "all candidates pass",
			candidates: []TypeCandidate{
				{Type: "person", Frequency: 100, Consistency: 0.85},
				{Type: "organization", Frequency: 75, Consistency: 0.80},
				{Type: "concept", Frequency: 60, Consistency: 0.75},
			},
			minOccurrences:     50,
			minConsistency:     0.70,
			expectedCandidates: []string{"person", "organization", "concept"},
		},
		{
			name: "no candidates pass",
			candidates: []TypeCandidate{
				{Type: "concept", Frequency: 30, Consistency: 0.60},
				{Type: "event", Frequency: 20, Consistency: 0.55},
			},
			minOccurrences:     50,
			minConsistency:     0.70,
			expectedCandidates: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := ApplyThresholds(tt.candidates, tt.minOccurrences, tt.minConsistency)

			if len(filtered) != len(tt.expectedCandidates) {
				t.Errorf("Expected %d candidates, got %d", len(tt.expectedCandidates), len(filtered))
			}

			candidateMap := make(map[string]bool)
			for _, candidate := range filtered {
				candidateMap[candidate.Type] = true

				// Verify thresholds
				if candidate.Frequency < tt.minOccurrences {
					t.Errorf("Candidate '%s' has frequency %d < minimum %d", candidate.Type, candidate.Frequency, tt.minOccurrences)
				}
				if candidate.Consistency < tt.minConsistency {
					t.Errorf("Candidate '%s' has consistency %.2f < minimum %.2f", candidate.Type, candidate.Consistency, tt.minConsistency)
				}
			}

			for _, expected := range tt.expectedCandidates {
				if !candidateMap[expected] {
					t.Errorf("Expected candidate '%s' not found in filtered results", expected)
				}
			}
		})
	}
}

func TestSortByScore(t *testing.T) {
	tests := []struct {
		name          string
		candidates    []TypeCandidate
		expectedOrder []string
	}{
		{
			name: "sort by descending score",
			candidates: []TypeCandidate{
				{Type: "concept", Score: 25.5},
				{Type: "person", Score: 55.3},
				{Type: "organization", Score: 40.2},
			},
			expectedOrder: []string{"person", "organization", "concept"},
		},
		{
			name: "already sorted",
			candidates: []TypeCandidate{
				{Type: "person", Score: 100.0},
				{Type: "organization", Score: 75.0},
				{Type: "concept", Score: 50.0},
			},
			expectedOrder: []string{"person", "organization", "concept"},
		},
		{
			name: "reverse order",
			candidates: []TypeCandidate{
				{Type: "concept", Score: 10.0},
				{Type: "organization", Score: 50.0},
				{Type: "person", Score: 100.0},
			},
			expectedOrder: []string{"person", "organization", "concept"},
		},
		{
			name: "equal scores maintain stable sort",
			candidates: []TypeCandidate{
				{Type: "person", Score: 50.0},
				{Type: "organization", Score: 50.0},
			},
			expectedOrder: []string{"person", "organization"}, // or organization, person - stable sort
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted := SortByScore(tt.candidates)

			if len(sorted) != len(tt.expectedOrder) {
				t.Errorf("Expected %d candidates, got %d", len(tt.expectedOrder), len(sorted))
			}

			for i, candidate := range sorted {
				if i >= len(tt.expectedOrder) {
					break
				}
				expected := tt.expectedOrder[i]
				// For equal scores, accept either order (stable sort)
				if i > 0 && sorted[i-1].Score == candidate.Score {
					continue
				}
				if candidate.Type != expected {
					t.Errorf("Position %d: expected type '%s', got '%s'", i, expected, candidate.Type)
				}

				// Verify descending order
				if i > 0 && sorted[i].Score > sorted[i-1].Score {
					t.Errorf("Scores not in descending order: position %d (%.2f) > position %d (%.2f)",
						i, sorted[i].Score, i-1, sorted[i-1].Score)
				}
			}
		})
	}
}

func TestRankCandidates(t *testing.T) {
	tests := []struct {
		name               string
		candidates         []TypeCandidate
		topN               int
		expectedCandidates []string
	}{
		{
			name: "rank top 3 candidates",
			candidates: []TypeCandidate{
				{Type: "person", Frequency: 100, Density: 50, Consistency: 0.85},
				{Type: "organization", Frequency: 75, Density: 40, Consistency: 0.80},
				{Type: "concept", Frequency: 60, Density: 30, Consistency: 0.75},
				{Type: "event", Frequency: 40, Density: 20, Consistency: 0.70},
			},
			topN:               3,
			expectedCandidates: []string{"person", "organization", "concept"},
		},
		{
			name: "request more than available",
			candidates: []TypeCandidate{
				{Type: "person", Frequency: 100, Density: 50, Consistency: 0.85},
				{Type: "organization", Frequency: 75, Density: 40, Consistency: 0.80},
			},
			topN:               10,
			expectedCandidates: []string{"person", "organization"},
		},
		{
			name: "top 1 candidate",
			candidates: []TypeCandidate{
				{Type: "person", Frequency: 100, Density: 50, Consistency: 0.85},
				{Type: "organization", Frequency: 75, Density: 40, Consistency: 0.80},
				{Type: "concept", Frequency: 60, Density: 30, Consistency: 0.75},
			},
			topN:               1,
			expectedCandidates: []string{"person"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate scores for each candidate
			for i := range tt.candidates {
				tt.candidates[i].Score = CalculateScore(
					float64(tt.candidates[i].Frequency),
					tt.candidates[i].Density,
					tt.candidates[i].Consistency,
				)
			}

			ranked := RankCandidates(tt.candidates, tt.topN)

			if len(ranked) != len(tt.expectedCandidates) {
				t.Errorf("Expected %d candidates, got %d", len(tt.expectedCandidates), len(ranked))
			}

			for i, candidate := range ranked {
				if i >= len(tt.expectedCandidates) {
					break
				}
				expected := tt.expectedCandidates[i]
				if candidate.Type != expected {
					t.Errorf("Position %d: expected type '%s', got '%s'", i, expected, candidate.Type)
				}
			}
		})
	}
}
