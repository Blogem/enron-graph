package analyst

import (
	"context"
	"sort"

	"github.com/Blogem/enron-graph/ent"
)

// T082: Candidate ranking implementation
// Score = 0.4*frequency + 0.3*density + 0.3*consistency

// TypeCandidate represents a type candidate with scoring metrics
type TypeCandidate struct {
	Type        string
	Frequency   int
	Density     float64
	Consistency float64
	Score       float64
}

// CalculateScore calculates the ranking score for a candidate
// Formula: 0.4*frequency + 0.3*density + 0.3*consistency
func CalculateScore(frequency, density, consistency float64) float64 {
	return 0.4*frequency + 0.3*density + 0.3*consistency
}

// ApplyThresholds filters candidates based on minimum occurrence and consistency thresholds
func ApplyThresholds(candidates []TypeCandidate, minOccurrences int, minConsistency float64) []TypeCandidate {
	filtered := []TypeCandidate{}

	for _, candidate := range candidates {
		if candidate.Frequency >= minOccurrences && candidate.Consistency >= minConsistency {
			filtered = append(filtered, candidate)
		}
	}

	return filtered
}

// SortByScore sorts candidates by score in descending order
func SortByScore(candidates []TypeCandidate) []TypeCandidate {
	sorted := make([]TypeCandidate, len(candidates))
	copy(sorted, candidates)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	return sorted
}

// RankCandidates ranks candidates by score and returns top N
func RankCandidates(candidates []TypeCandidate, topN int) []TypeCandidate {
	sorted := SortByScore(candidates)

	if len(sorted) > topN {
		return sorted[:topN]
	}

	return sorted
}

// AnalyzeAndRankCandidates performs complete analysis and ranking
func AnalyzeAndRankCandidates(ctx context.Context, client *ent.Client, minOccurrences int, minConsistency float64, topN int) ([]TypeCandidate, error) {
	// Detect patterns
	patterns, err := DetectPatterns(ctx, client)
	if err != nil {
		return nil, err
	}

	// Convert to candidates
	candidates := make([]TypeCandidate, 0, len(patterns))
	for _, stats := range patterns {
		// Calculate average consistency across all properties
		var totalConsistency float64
		propCount := 0
		for _, consistency := range stats.PropertyConsistency {
			totalConsistency += consistency
			propCount++
		}

		avgConsistency := 0.0
		if propCount > 0 {
			avgConsistency = totalConsistency / float64(propCount)
		}

		candidate := TypeCandidate{
			Type:        stats.Type,
			Frequency:   stats.Frequency,
			Density:     stats.AvgDensity,
			Consistency: avgConsistency,
		}

		// Calculate score
		candidate.Score = CalculateScore(
			float64(candidate.Frequency),
			candidate.Density,
			candidate.Consistency,
		)

		candidates = append(candidates, candidate)
	}

	// Apply thresholds
	filtered := ApplyThresholds(candidates, minOccurrences, minConsistency)

	// Rank and return top N
	return RankCandidates(filtered, topN), nil
}
