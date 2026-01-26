package integration

import (
	"context"
	"testing"

	"github.com/Blogem/enron-graph/ent/discoveredentity"
	"github.com/Blogem/enron-graph/ent/relationship"
)

// TestNoDuplicateEntitiesByUniqueID tests SC-009 data consistency requirement (T148)
// Validates that no duplicate entities exist with the same unique_id
func TestNoDuplicateEntitiesByUniqueID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, db := SetupTestDBWithSQL(t)

	// Create diverse test dataset with various entity types
	testEntities := []struct {
		uniqueID     string
		typeCategory string
		name         string
	}{
		{"jeff.skilling@enron.com", "person", "Jeff Skilling"},
		{"kenneth.lay@enron.com", "person", "Kenneth Lay"},
		{"enron.com", "organization", "Enron Corporation"},
		{"energy-trading", "concept", "Energy Trading"},
		{"california-crisis", "concept", "California Energy Crisis"},
		{"andy.fastow@enron.com", "person", "Andy Fastow"},
		{"ect.enron.com", "organization", "Enron Capital & Trade"},
	}

	// Load test entities
	for _, te := range testEntities {
		_, err := client.DiscoveredEntity.Create().
			SetUniqueID(te.uniqueID).
			SetTypeCategory(te.typeCategory).
			SetName(te.name).
			SetConfidenceScore(0.85).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create test entity %s: %v", te.uniqueID, err)
		}
	}

	// Query for duplicates by unique_id
	// This query groups by unique_id and counts occurrences
	type duplicateResult struct {
		UniqueID string `json:"unique_id"`
		Count    int    `json:"count"`
	}

	var duplicates []duplicateResult

	// Use raw SQL to find duplicates (ent doesn't have good GROUP BY support)
	rows, err := db.QueryContext(ctx, `
		SELECT unique_id, COUNT(*) as count
		FROM discovered_entities
		GROUP BY unique_id
		HAVING COUNT(*) > 1
	`)
	if err != nil {
		t.Fatalf("Failed to query for duplicates: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dup duplicateResult
		if err := rows.Scan(&dup.UniqueID, &dup.Count); err != nil {
			t.Fatalf("Failed to scan duplicate row: %v", err)
		}
		duplicates = append(duplicates, dup)
	}

	// Verify count = 0
	if len(duplicates) > 0 {
		t.Errorf("Found %d duplicate unique_id values:", len(duplicates))
		for _, dup := range duplicates {
			t.Errorf("  - unique_id '%s' appears %d times", dup.UniqueID, dup.Count)
		}
	}

	// Additional verification: check that unique constraint is enforced
	// Try to insert a duplicate and verify it fails
	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("jeff.skilling@enron.com"). // Already exists
		SetTypeCategory("person").
		SetName("Jeffrey K. Skilling").
		SetConfidenceScore(0.90).
		Save(ctx)

	if err == nil {
		t.Error("Expected error when inserting duplicate unique_id, but got none")
	}

	t.Logf("✓ Data integrity verified: No duplicate entities with same unique_id")
}

// TestAllRelationshipsReferenceValidEntities tests SC-009 referential integrity (T149)
// Validates that all relationships reference valid entities in the discovered_entities table
func TestAllRelationshipsReferenceValidEntities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := SetupTestDB(t)

	// Create test entities
	entity1, err := client.DiscoveredEntity.Create().
		SetUniqueID("person1@enron.com").
		SetTypeCategory("person").
		SetName("Person One").
		SetConfidenceScore(0.90).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create entity1: %v", err)
	}

	entity2, err := client.DiscoveredEntity.Create().
		SetUniqueID("person2@enron.com").
		SetTypeCategory("person").
		SetName("Person Two").
		SetConfidenceScore(0.88).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create entity2: %v", err)
	}

	entity3, err := client.DiscoveredEntity.Create().
		SetUniqueID("org1@enron.com").
		SetTypeCategory("organization").
		SetName("Organization One").
		SetConfidenceScore(0.92).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create entity3: %v", err)
	}

	// Create valid relationships between entities
	validRelationships := []struct {
		relType  string
		fromType string
		fromID   int
		toType   string
		toID     int
	}{
		{"COMMUNICATES_WITH", "discovered_entity", entity1.ID, "discovered_entity", entity2.ID},
		{"COMMUNICATES_WITH", "discovered_entity", entity2.ID, "discovered_entity", entity1.ID},
		{"WORKS_FOR", "discovered_entity", entity1.ID, "discovered_entity", entity3.ID},
		{"WORKS_FOR", "discovered_entity", entity2.ID, "discovered_entity", entity3.ID},
	}

	for _, rel := range validRelationships {
		_, err := client.Relationship.Create().
			SetType(rel.relType).
			SetFromType(rel.fromType).
			SetFromID(rel.fromID).
			SetToType(rel.toType).
			SetToID(rel.toID).
			SetConfidenceScore(0.85).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create relationship: %v", err)
		}
	}

	// Query relationships with LEFT JOIN on entities to find orphaned relationships
	// Check "from" references
	orphanedFromRels, err := client.Relationship.Query().
		Where(relationship.FromTypeEQ("discovered_entity")).
		All(ctx)
	if err != nil {
		t.Fatalf("Failed to query relationships: %v", err)
	}

	var invalidFromRefs []int
	for _, rel := range orphanedFromRels {
		// Verify the from_id exists in discovered_entities
		exists, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.IDEQ(rel.FromID)).
			Exist(ctx)
		if err != nil {
			t.Fatalf("Failed to check entity existence: %v", err)
		}
		if !exists {
			invalidFromRefs = append(invalidFromRefs, rel.ID)
		}
	}

	if len(invalidFromRefs) > 0 {
		t.Errorf("Found %d relationships with invalid from_id references:", len(invalidFromRefs))
		for _, relID := range invalidFromRefs {
			t.Errorf("  - relationship ID %d", relID)
		}
	}

	// Check "to" references
	orphanedToRels, err := client.Relationship.Query().
		Where(relationship.ToTypeEQ("discovered_entity")).
		All(ctx)
	if err != nil {
		t.Fatalf("Failed to query relationships: %v", err)
	}

	var invalidToRefs []int
	for _, rel := range orphanedToRels {
		// Verify the to_id exists in discovered_entities
		exists, err := client.DiscoveredEntity.Query().
			Where(discoveredentity.IDEQ(rel.ToID)).
			Exist(ctx)
		if err != nil {
			t.Fatalf("Failed to check entity existence: %v", err)
		}
		if !exists {
			invalidToRefs = append(invalidToRefs, rel.ID)
		}
	}

	if len(invalidToRefs) > 0 {
		t.Errorf("Found %d relationships with invalid to_id references:", len(invalidToRefs))
		for _, relID := range invalidToRefs {
			t.Errorf("  - relationship ID %d", relID)
		}
	}

	// Verify all relationships have valid references
	totalInvalid := len(invalidFromRefs) + len(invalidToRefs)
	if totalInvalid == 0 {
		t.Logf("✓ Referential integrity verified: All %d relationships reference valid entities", len(orphanedFromRels))
	}
}
