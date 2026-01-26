package contract

import (
	"context"
	"testing"

	"github.com/Blogem/enron-graph/internal/explorer"
	_ "github.com/mattn/go-sqlite3"
)

// T016: Test GetSchema returns promoted types
func TestSchemaService_GetSchema_ReturnsPromotedTypes(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client, db)
	ctx := context.Background()

	response, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	if len(response.PromotedTypes) == 0 {
		t.Fatal("Expected promoted types, got none")
	}

	// Verify emails table is in promoted types (actual Ent schema table)
	typeNames := make(map[string]bool)
	for _, pt := range response.PromotedTypes {
		typeNames[pt.Name] = true
		if !pt.IsPromoted {
			t.Errorf("Type %s should be marked as promoted", pt.Name)
		}
	}

	if !typeNames["emails"] {
		t.Error("Expected 'emails' in promoted types (actual Ent table)")
	}
}

// T017: Test GetSchema returns discovered types
func TestSchemaService_GetSchema_ReturnsDiscoveredTypes(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client, db)
	ctx := context.Background()

	response, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	if len(response.DiscoveredTypes) == 0 {
		t.Fatal("Expected discovered types, got none")
	}

	// Verify location and product are in discovered types
	typeNames := make(map[string]bool)
	for _, dt := range response.DiscoveredTypes {
		typeNames[dt.Name] = true
		if dt.IsPromoted {
			t.Errorf("Type %s should not be marked as promoted", dt.Name)
		}
	}

	if !typeNames["location"] {
		t.Error("Expected 'location' in discovered types")
	}
	if !typeNames["product"] {
		t.Error("Expected 'product' in discovered types")
	}
}

// T018: Test GetSchema has no overlap between promoted and discovered
func TestSchemaService_GetSchema_NoOverlapBetweenCategories(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client, db)
	ctx := context.Background()

	response, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	promotedNames := make(map[string]bool)
	for _, pt := range response.PromotedTypes {
		promotedNames[pt.Name] = true
	}

	discoveredNames := make(map[string]bool)
	for _, dt := range response.DiscoveredTypes {
		discoveredNames[dt.Name] = true
	}

	// Check for overlap
	for name := range promotedNames {
		if discoveredNames[name] {
			t.Errorf("Type %s appears in both promoted and discovered", name)
		}
	}
}

// T019: Test GetSchema includes property metadata
func TestSchemaService_GetSchema_IncludesPropertyMetadata(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client, db)
	ctx := context.Background()

	response, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	// Check that at least one type has properties
	foundProperties := false
	for _, pt := range response.PromotedTypes {
		if len(pt.Properties) > 0 {
			foundProperties = true
			break
		}
	}

	if !foundProperties {
		for _, dt := range response.DiscoveredTypes {
			if len(dt.Properties) > 0 {
				foundProperties = true
				break
			}
		}
	}

	if !foundProperties {
		t.Error("Expected at least one type to have properties")
	}
}

// T020: Test GetTypeDetails returns type details
func TestSchemaService_GetTypeDetails_ReturnsTypeDetails(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client, db)
	ctx := context.Background()

	// Query for a discovered type (person is in discovered_entities, not an actual table)
	details, err := service.GetTypeDetails(ctx, "person")
	if err != nil {
		t.Fatalf("GetTypeDetails failed: %v", err)
	}

	if details == nil {
		t.Fatal("Expected type details, got nil")
	}

	if details.Name != "person" {
		t.Errorf("Expected name 'person', got '%s'", details.Name)
	}

	if details.Count <= 0 {
		t.Errorf("Expected count > 0, got %d", details.Count)
	}

	// person is a discovered type, not promoted (no actual table)
	if details.IsPromoted {
		t.Error("Expected person to NOT be promoted (it's a discovered type)")
	}

	if len(details.Properties) == 0 {
		t.Error("Expected properties for person type")
	}
}

// T021: Test RefreshSchema updates schema
func TestSchemaService_RefreshSchema_UpdatesSchema(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client, db)
	ctx := context.Background()

	// Get initial schema
	initialResponse, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("Initial GetSchema failed: %v", err)
	}

	initialDiscoveredCount := len(initialResponse.DiscoveredTypes)

	// Add a new entity with a new type to discovered_entities
	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("event-meeting-123").
		SetTypeCategory("event").
		SetName("Board Meeting").
		SetConfidenceScore(0.80).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to add new entity: %v", err)
	}

	// Refresh schema
	err = service.RefreshSchema(ctx)
	if err != nil {
		t.Fatalf("RefreshSchema failed: %v", err)
	}

	// Get updated schema
	updatedResponse, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("Updated GetSchema failed: %v", err)
	}

	if len(updatedResponse.DiscoveredTypes) != initialDiscoveredCount+1 {
		t.Errorf("Expected %d discovered types after refresh, got %d",
			initialDiscoveredCount+1, len(updatedResponse.DiscoveredTypes))
	}
}
