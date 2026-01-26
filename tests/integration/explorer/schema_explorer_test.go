package explorer_test

import (
	"context"
	"testing"

	"github.com/Blogem/enron-graph/ent/enttest"
	"github.com/Blogem/enron-graph/internal/explorer"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestService creates a test service with in-memory database and test data
func setupTestService(t *testing.T) *explorer.SchemaService {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

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

	return explorer.NewSchemaService(client)
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

	// Verify promoted types section
	if len(schema.PromotedTypes) != 2 {
		t.Errorf("Expected 2 promoted types, got %d", len(schema.PromotedTypes))
	}

	promotedNames := make(map[string]bool)
	for _, pt := range schema.PromotedTypes {
		promotedNames[pt.Name] = true
		if !pt.IsPromoted {
			t.Errorf("Type %s should be marked as promoted", pt.Name)
		}
		if pt.Count <= 0 {
			t.Errorf("Type %s should have count > 0, got %d", pt.Name, pt.Count)
		}
	}

	if !promotedNames["person"] {
		t.Error("Expected 'person' in promoted types")
	}
	if !promotedNames["organization"] {
		t.Error("Expected 'organization' in promoted types")
	}

	// Verify discovered types section
	if len(schema.DiscoveredTypes) != 2 {
		t.Errorf("Expected 2 discovered types, got %d", len(schema.DiscoveredTypes))
	}

	discoveredNames := make(map[string]bool)
	for _, dt := range schema.DiscoveredTypes {
		discoveredNames[dt.Name] = true
		if dt.IsPromoted {
			t.Errorf("Discovered type %s should not be marked as promoted", dt.Name)
		}
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

	// Simulate user clicking on "person" type in UI
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

	if !details.IsPromoted {
		t.Error("Expected person to be marked as promoted in details view")
	}

	// Verify properties metadata is included
	if len(details.Properties) == 0 {
		t.Error("Expected properties array for person type")
	}

	// Test clicking on a discovered type
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
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	ctx := context.Background()

	// Initial setup with 2 promoted types
	client.SchemaPromotion.Create().SetTypeName("person").SaveX(ctx)
	client.SchemaPromotion.Create().SetTypeName("organization").SaveX(ctx)
	client.DiscoveredEntity.Create().
		SetUniqueID("p1").SetTypeCategory("person").SetName("John").SetConfidenceScore(0.9).SaveX(ctx)
	client.DiscoveredEntity.Create().
		SetUniqueID("o1").SetTypeCategory("organization").SetName("Enron").SetConfidenceScore(0.9).SaveX(ctx)
	client.DiscoveredEntity.Create().
		SetUniqueID("l1").SetTypeCategory("location").SetName("Houston").SetConfidenceScore(0.8).SaveX(ctx)

	service := explorer.NewSchemaService(client)

	// Get initial schema
	initialSchema, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("Initial GetSchema failed: %v", err)
	}

	initialPromotedCount := len(initialSchema.PromotedTypes)
	if initialPromotedCount != 2 {
		t.Fatalf("Expected 2 initial promoted types, got %d", initialPromotedCount)
	}

	// Simulate external promotion of location type
	_, err = client.SchemaPromotion.Create().
		SetTypeName("location").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to add location promotion: %v", err)
	}

	// Without refresh, cached schema should still show old data
	schemaBeforeRefresh, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema before refresh failed: %v", err)
	}

	// Cache test: should still show old count
	if len(schemaBeforeRefresh.PromotedTypes) != initialPromotedCount {
		t.Logf("Cache behavior: Expected %d promoted types (cached), got %d",
			initialPromotedCount, len(schemaBeforeRefresh.PromotedTypes))
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

	// Verify the promoted count increased
	if len(refreshedSchema.PromotedTypes) != initialPromotedCount+1 {
		t.Errorf("Expected %d promoted types after refresh, got %d",
			initialPromotedCount+1, len(refreshedSchema.PromotedTypes))
	}

	// Verify location is now in promoted types
	foundInPromoted := false
	for _, pt := range refreshedSchema.PromotedTypes {
		if pt.Name == "location" {
			foundInPromoted = true
			if !pt.IsPromoted {
				t.Error("Location should be marked as promoted after refresh")
			}
			break
		}
	}

	if !foundInPromoted {
		t.Error("Expected 'location' in promoted types after refresh")
	}

	// Verify location is no longer in discovered types
	for _, dt := range refreshedSchema.DiscoveredTypes {
		if dt.Name == "location" {
			t.Error("Location should not appear in discovered types after being promoted")
		}
	}
}
