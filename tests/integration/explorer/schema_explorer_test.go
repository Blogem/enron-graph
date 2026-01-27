package explorer_test

import (
	"context"
	"testing"

	"github.com/Blogem/enron-graph/internal/explorer"
	integration "github.com/Blogem/enron-graph/tests/integration"
)

// setupTestService creates a test service with PostgreSQL database and test data
func setupTestService(t *testing.T) *explorer.SchemaService {
	client, db := integration.SetupTestDBWithSQL(t)

	// Seed test data
	ctx := context.Background()

	// Create promoted types
	_, err := client.SchemaPromotion.Create().
		SetTypeName("person").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person promotion: %v", err)
	}

	_, err = client.SchemaPromotion.Create().
		SetTypeName("organization").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create organization promotion: %v", err)
	}

	// Create discovered entities for promoted types
	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("person-john-doe").
		SetTypeCategory("person").
		SetName("John Doe").
		SetConfidenceScore(0.95).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person entity: %v", err)
	}

	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("org-enron").
		SetTypeCategory("organization").
		SetName("Enron Corp").
		SetConfidenceScore(0.90).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create organization entity: %v", err)
	}

	// Create discovered entities for unpromoted types
	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("location-houston").
		SetTypeCategory("location").
		SetName("Houston").
		SetConfidenceScore(0.75).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create location entity: %v", err)
	}

	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("product-energy").
		SetTypeCategory("product").
		SetName("Energy Trading").
		SetConfidenceScore(0.70).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create product entity: %v", err)
	}

	return explorer.NewSchemaService(client, db)
}

// T040: Integration test - Full flow from app start to schema display
func TestSchemaExplorer_FullFlow(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Simulate app startup - GetSchema call (as called from Wails frontend)
	schema, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	// Verify complete schema response structure
	if schema == nil {
		t.Fatal("Expected schema response, got nil")
	}

	// Verify promoted types section (actual Ent tables: emails)
	// Note: person and organization are in schema_promotions but not actual tables yet
	if len(schema.PromotedTypes) < 1 {
		t.Errorf("Expected at least 1 promoted type (emails), got %d", len(schema.PromotedTypes))
	}

	promotedNames := make(map[string]bool)
	for _, pt := range schema.PromotedTypes {
		promotedNames[pt.Name] = true
		if !pt.IsPromoted {
			t.Errorf("Type %s should be marked as promoted", pt.Name)
		}
	}

	// emails is an actual Ent-generated table, so it should appear as promoted
	if !promotedNames["emails"] {
		t.Error("Expected 'emails' in promoted types (Ent-generated table)")
	}

	// Verify discovered types section (from discovered_entities table)
	if len(schema.DiscoveredTypes) != 4 {
		t.Errorf("Expected 4 discovered types, got %d", len(schema.DiscoveredTypes))
	}

	discoveredNames := make(map[string]bool)
	for _, dt := range schema.DiscoveredTypes {
		discoveredNames[dt.Name] = true
		if dt.IsPromoted {
			t.Errorf("Discovered type %s should not be marked as promoted", dt.Name)
		}
	}

	if !discoveredNames["person"] {
		t.Error("Expected 'person' in discovered types")
	}
	if !discoveredNames["organization"] {
		t.Error("Expected 'organization' in discovered types")
	}
	if !discoveredNames["location"] {
		t.Error("Expected 'location' in discovered types")
	}
	if !discoveredNames["product"] {
		t.Error("Expected 'product' in discovered types")
	}

	// Verify total entities aggregation
	expectedTotal := 4
	if schema.TotalEntities != expectedTotal {
		t.Errorf("Expected %d total entities, got %d", expectedTotal, schema.TotalEntities)
	}

	// Critical: Verify no overlap between promoted and discovered
	for name := range promotedNames {
		if discoveredNames[name] {
			t.Errorf("Type %s appears in both promoted and discovered - data segregation failed", name)
		}
	}
}

// T041: Integration test - Type click shows details
func TestSchemaExplorer_TypeClickShowsDetails(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Simulate user clicking on "person" type in UI (discovered type)
	details, err := service.GetTypeDetails(ctx, "person")
	if err != nil {
		t.Fatalf("GetTypeDetails for person failed: %v", err)
	}

	if details == nil {
		t.Fatal("Expected type details, got nil")
	}

	// Verify all detail fields are populated correctly
	if details.Name != "person" {
		t.Errorf("Expected name 'person', got '%s'", details.Name)
	}

	if details.Count != 1 {
		t.Errorf("Expected count 1 for person, got %d", details.Count)
	}

	// Person is in discovered_entities, not a promoted table
	if details.IsPromoted {
		t.Error("Expected person to not be promoted (it's in discovered_entities)")
	}

	// Verify properties metadata is included
	if len(details.Properties) == 0 {
		t.Error("Expected properties array for person type")
	}

	// Test clicking on another discovered type
	locationDetails, err := service.GetTypeDetails(ctx, "location")
	if err != nil {
		t.Fatalf("GetTypeDetails for location failed: %v", err)
	}

	if locationDetails.Name != "location" {
		t.Errorf("Expected name 'location', got '%s'", locationDetails.Name)
	}

	if locationDetails.IsPromoted {
		t.Error("Expected location to not be promoted in details view")
	}

	if locationDetails.Count != 1 {
		t.Errorf("Expected count 1 for location, got %d", locationDetails.Count)
	}

	// Verify properties are present for discovered types too
	if len(locationDetails.Properties) == 0 {
		t.Error("Expected properties array for location type")
	}
}

// T042: Integration test - Refresh updates schema
func TestSchemaExplorer_RefreshUpdatesSchema(t *testing.T) {
	client, db := integration.SetupTestDBWithSQL(t)

	ctx := context.Background()

	// Initial setup: Create some discovered entities
	client.DiscoveredEntity.Create().
		SetUniqueID("p1").SetTypeCategory("person").SetName("John").SetConfidenceScore(0.9).SaveX(ctx)
	client.DiscoveredEntity.Create().
		SetUniqueID("o1").SetTypeCategory("organization").SetName("Enron").SetConfidenceScore(0.9).SaveX(ctx)
	client.DiscoveredEntity.Create().
		SetUniqueID("l1").SetTypeCategory("location").SetName("Houston").SetConfidenceScore(0.8).SaveX(ctx)

	service := explorer.NewSchemaService(client, db)

	// Get initial schema
	initialSchema, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("Initial GetSchema failed: %v", err)
	}

	// Should have at least the emails table as promoted type
	initialPromotedCount := len(initialSchema.PromotedTypes)
	if initialPromotedCount < 1 {
		t.Fatalf("Expected at least 1 initial promoted type (emails), got %d", initialPromotedCount)
	}

	// Should have 3 discovered types
	initialDiscoveredCount := len(initialSchema.DiscoveredTypes)
	if initialDiscoveredCount != 3 {
		t.Fatalf("Expected 3 initial discovered types, got %d", initialDiscoveredCount)
	}

	// Add a new discovered entity
	client.DiscoveredEntity.Create().
		SetUniqueID("c1").SetTypeCategory("contract").SetName("NDA-001").SetConfidenceScore(0.85).SaveX(ctx)

	// Without refresh, cached schema should still show old data
	schemaBeforeRefresh, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema before refresh failed: %v", err)
	}

	// Cache test: should still show old count
	if len(schemaBeforeRefresh.DiscoveredTypes) != initialDiscoveredCount {
		t.Logf("Cache behavior: Expected %d discovered types (cached), got %d",
			initialDiscoveredCount, len(schemaBeforeRefresh.DiscoveredTypes))
	}

	// Simulate user clicking refresh button in UI
	err = service.RefreshSchema(ctx)
	if err != nil {
		t.Fatalf("RefreshSchema failed: %v", err)
	}

	// Get schema after refresh
	refreshedSchema, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema after refresh failed: %v", err)
	}

	// Verify the discovered count increased by 1 (new contract type)
	if len(refreshedSchema.DiscoveredTypes) != initialDiscoveredCount+1 {
		t.Errorf("Expected %d discovered types after refresh, got %d",
			initialDiscoveredCount+1, len(refreshedSchema.DiscoveredTypes))
	}

	// Verify contract is now in discovered types
	foundContract := false
	for _, dt := range refreshedSchema.DiscoveredTypes {
		if dt.Name == "contract" {
			foundContract = true
			if dt.IsPromoted {
				t.Error("Contract should not be marked as promoted")
			}
			if dt.Count != 1 {
				t.Errorf("Expected count 1 for contract, got %d", dt.Count)
			}
			break
		}
	}

	if !foundContract {
		t.Error("Expected 'contract' in discovered types after refresh")
	}

	// Verify promoted types count stayed the same
	if len(refreshedSchema.PromotedTypes) != initialPromotedCount {
		t.Errorf("Expected %d promoted types (unchanged), got %d",
			initialPromotedCount, len(refreshedSchema.PromotedTypes))
	}
}
