package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/analyst"
)

// TestAnalystPatternDetection tests the analyst pattern detection functionality (T088)
func TestAnalystPatternDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup: Connect to test database with clean schema
	ctx := context.Background()
	client := SetupTestDB(t)

	// Pre-populate test database with diverse discovered entities
	err := populateTestEntities(ctx, client)
	if err != nil {
		t.Fatalf("Failed to populate test entities: %v", err)
	}

	// Test: Run pattern detection
	patterns, err := analyst.DetectPatterns(ctx, client)
	if err != nil {
		t.Fatalf("DetectPatterns failed: %v", err)
	}

	// Verify: Candidates identified with correct scores
	if len(patterns) == 0 {
		t.Fatal("Expected pattern detection to identify candidates, got 0")
	}

	t.Logf("Detected %d type patterns", len(patterns))

	// Verify each pattern has required statistics
	for typeName, stats := range patterns {
		if stats.Type != typeName {
			t.Errorf("Pattern type mismatch: expected %s, got %s", typeName, stats.Type)
		}

		if stats.Frequency <= 0 {
			t.Errorf("Pattern %s has invalid frequency: %d", typeName, stats.Frequency)
		}

		if stats.PropertyConsistency == nil {
			t.Errorf("Pattern %s has nil PropertyConsistency", typeName)
		}

		t.Logf("Pattern %s: Frequency=%d, AvgDensity=%.2f, Properties=%d",
			typeName, stats.Frequency, stats.AvgDensity, len(stats.Properties))
	}

	// Test: Rank candidates using the ranking algorithm
	minOccurrences := 10  // Lower threshold for test data
	minConsistency := 0.5 // Lower threshold for test data
	topN := 10

	candidates, err := analyst.AnalyzeAndRankCandidates(ctx, client, minOccurrences, minConsistency, topN)
	if err != nil {
		t.Fatalf("AnalyzeAndRankCandidates failed: %v", err)
	}

	// Verify: Ranking by frequency/density/consistency
	if len(candidates) == 0 {
		t.Fatal("Expected candidates to be ranked, got 0")
	}

	t.Logf("Ranked %d candidates", len(candidates))

	// Verify candidates are sorted by score (descending)
	for i := 0; i < len(candidates)-1; i++ {
		if candidates[i].Score < candidates[i+1].Score {
			t.Errorf("Candidates not properly sorted: candidate[%d].Score (%.2f) < candidate[%d].Score (%.2f)",
				i, candidates[i].Score, i+1, candidates[i+1].Score)
		}
	}

	// Verify each candidate has valid metrics
	for i, candidate := range candidates {
		t.Logf("Candidate %d: Type=%s, Frequency=%d, Density=%.2f, Consistency=%.2f, Score=%.2f",
			i+1, candidate.Type, candidate.Frequency, candidate.Density, candidate.Consistency, candidate.Score)

		if candidate.Frequency < minOccurrences {
			t.Errorf("Candidate %s has frequency %d below threshold %d",
				candidate.Type, candidate.Frequency, minOccurrences)
		}

		if candidate.Consistency < minConsistency {
			t.Errorf("Candidate %s has consistency %.2f below threshold %.2f",
				candidate.Type, candidate.Consistency, minConsistency)
		}

		// Verify score calculation
		expectedScore := 0.4*float64(candidate.Frequency) + 0.3*candidate.Density + 0.3*candidate.Consistency
		if candidate.Score != expectedScore {
			t.Errorf("Candidate %s has incorrect score: expected %.2f, got %.2f",
				candidate.Type, expectedScore, candidate.Score)
		}
	}

	// Verify: Check specific expected patterns from test data
	// We should have multiple types: person, organization, project
	typesSeen := make(map[string]bool)
	for _, candidate := range candidates {
		typesSeen[candidate.Type] = true
	}

	expectedTypes := []string{"person", "organization", "project"}
	for _, expectedType := range expectedTypes {
		if !typesSeen[expectedType] {
			t.Logf("Warning: Expected type '%s' not found in candidates (may be below threshold)", expectedType)
		}
	}

	// Teardown: Cleanup is handled by SetupTestDB's t.Cleanup
}

// populateTestEntities creates diverse discovered entities for testing pattern detection
func populateTestEntities(ctx context.Context, client interface{}) error {
	// Type assertion to get the ent client
	type entClient interface {
		DiscoveredEntity() interface{}
	}

	// We'll use a simpler approach - directly use the client's DiscoveredEntity method
	// This creates diverse entities with different types and properties

	// Create person entities (high frequency, high consistency)
	personProperties := map[string]interface{}{
		"email":    "user@example.com",
		"role":     "employee",
		"location": "Houston",
	}

	for i := 1; i <= 50; i++ {
		_, err := client.(*ent.Client).DiscoveredEntity.
			Create().
			SetUniqueID(fmt.Sprintf("person_%d", i)).
			SetTypeCategory("person").
			SetName(fmt.Sprintf("Person %d", i)).
			SetProperties(personProperties).
			SetConfidenceScore(0.9).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	// Create organization entities (medium frequency, high consistency)
	orgProperties := map[string]interface{}{
		"domain":   "example.com",
		"industry": "energy",
		"size":     "large",
	}

	for i := 1; i <= 30; i++ {
		_, err := client.(*ent.Client).DiscoveredEntity.
			Create().
			SetUniqueID(fmt.Sprintf("org_%d", i)).
			SetTypeCategory("organization").
			SetName(fmt.Sprintf("Organization %d", i)).
			SetProperties(orgProperties).
			SetConfidenceScore(0.85).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	// Create project entities (medium frequency, medium consistency)
	for i := 1; i <= 25; i++ {
		projectProps := map[string]interface{}{
			"status": "active",
			"budget": 1000000,
		}

		// Some projects have extra properties (inconsistent)
		if i%3 == 0 {
			projectProps["deadline"] = "2026-12-31"
		}
		if i%5 == 0 {
			projectProps["manager"] = "person_1"
		}

		_, err := client.(*ent.Client).DiscoveredEntity.
			Create().
			SetUniqueID(fmt.Sprintf("project_%d", i)).
			SetTypeCategory("project").
			SetName(fmt.Sprintf("Project %d", i)).
			SetProperties(projectProps).
			SetConfidenceScore(0.75).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	// Create concept entities (low frequency, low consistency - should be filtered out)
	for i := 1; i <= 5; i++ {
		conceptProps := map[string]interface{}{
			"category": "business",
		}

		// Very inconsistent properties
		if i == 1 {
			conceptProps["sentiment"] = "positive"
		} else if i == 2 {
			conceptProps["importance"] = "high"
		}

		_, err := client.(*ent.Client).DiscoveredEntity.
			Create().
			SetUniqueID(fmt.Sprintf("concept_%d", i)).
			SetTypeCategory("concept").
			SetName(fmt.Sprintf("Concept %d", i)).
			SetProperties(conceptProps).
			SetConfidenceScore(0.6).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	// Create relationships to test density calculation
	// Person-to-person relationships (high density)
	for i := 1; i <= 20; i++ {
		fromID := i
		toID := (i % 50) + 1

		_, err := client.(*ent.Client).Relationship.
			Create().
			SetType("COMMUNICATES_WITH").
			SetFromType("discovered_entity").
			SetFromID(fromID).
			SetToType("discovered_entity").
			SetToID(toID).
			SetConfidenceScore(0.9).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	// Organization-to-person relationships (medium density)
	for i := 1; i <= 15; i++ {
		fromID := 50 + i // Organization IDs start after persons
		toID := (i % 50) + 1

		_, err := client.(*ent.Client).Relationship.
			Create().
			SetType("EMPLOYS").
			SetFromType("discovered_entity").
			SetFromID(fromID).
			SetToType("discovered_entity").
			SetToID(toID).
			SetConfidenceScore(0.85).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
