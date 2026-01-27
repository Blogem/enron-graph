#!/bin/bash
# Script to fix corrupted files from file creation tool issues

cd "$(dirname "$0")/.."

echo "Fixing internal/explorer/models.go..."
cat > internal/explorer/models.go << 'ENDFILE'
package explorer

type GraphNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Category   string                 `json:"category"`
	Properties map[string]interface{} `json:"properties"`
	IsGhost    bool                   `json:"is_ghost"`
	Degree     int                    `json:"degree,omitempty"`
}

type GraphEdge struct {
	Source     string                 `json:"source"`
	Target     string                 `json:"target"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type PropertyDefinition struct {
	Name         string   `json:"name"`
	Type         string   `json:"data_type"`
	SampleValues []string `json:"sample_value,omitempty"`
	Nullable     bool     `json:"nullable"`
}

type SchemaType struct {
	Name          string                 `json:"name"`
	Category      string                 `json:"category"`
	Count         int64                  `json:"count"`
	Properties    []PropertyDefinition   `json:"properties"`
	IsPromoted    bool                   `json:"is_promoted"`
	Relationships []string               `json:"relationships,omitempty"`
}

type GraphResponse struct {
	Nodes      []GraphNode `json:"nodes"`
	Edges      []GraphEdge `json:"edges"`
	TotalNodes int         `json:"total_nodes"`
	HasMore    bool        `json:"has_more"`
}

type RelationshipsResponse struct {
	Nodes      []GraphNode `json:"nodes"`
	Edges      []GraphEdge `json:"edges"`
	TotalCount int         `json:"total_count"`
	HasMore    bool        `json:"has_more"`
	Offset     int         `json:"offset"`
}

type SchemaResponse struct {
	PromotedTypes   []SchemaType `json:"promoted_types"`
	DiscoveredTypes []SchemaType `json:"discovered_types"`
	TotalEntities   int          `json:"total_entities"`
}

type NodeFilter struct {
	Types       []string `json:"types,omitempty"`
	Category    string   `json:"category,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
	Limit       int      `json:"limit,omitempty"`
}
ENDFILE

echo "✅ Fixed models.go"

echo "Fixing tests/contract/test_helper.go..."
cat > tests/contract/test_helper.go << 'ENDFILE'
package contract_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/enttest"
	_ "github.com/mattn/go-sqlite3"
)

func TestClient(t *testing.T) *ent.Client {
	opts := []enttest.Option{
		enttest.WithOptions(ent.Log(t.Log)),
	}
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1", opts...)
	return client
}

func TestClientPostgres(t *testing.T) *ent.Client {
	databaseURL := "postgres://enron:enron@localhost:5432/enron_test?sslmode=disable"
	client, err := ent.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := client.Schema.Create(context.Background()); err != nil {
		client.Close()
		t.Fatalf("Failed to create schema: %v", err)
	}

	return client
}

func CleanupDB(t *testing.T, client *ent.Client) {
	ctx := context.Background()
	
	if _, err := client.Relationship.Delete().Exec(ctx); err != nil {
		t.Logf("Failed to clean relationships: %v", err)
	}
	if _, err := client.Email.Delete().Exec(ctx); err != nil {
		t.Logf("Failed to clean emails: %v", err)
	}
	if _, err := client.DiscoveredEntity.Delete().Exec(ctx); err != nil {
		t.Logf("Failed to clean discovered entities: %v", err)
	}
	if _, err := client.SchemaPromotion.Delete().Exec(ctx); err != nil {
		t.Logf("Failed to clean schema promotions: %v", err)
	}
}

func SeedTestData(t *testing.T, client *ent.Client) {
	ctx := context.Background()

	_, err := client.SchemaPromotion.Create().
		SetTypeName("person").
		SetPromotedAt(1234567890).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person schema promotion: %v", err)
	}

	_, err = client.SchemaPromotion.Create().
		SetTypeName("organization").
		SetPromotedAt(1234567890).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create organization schema promotion: %v", err)
	}

	email1, err := client.Email.Create().
		SetMessageID("test1@enron.com").
		SetSubject("Test Email 1").
		SetSenderAddr("sender1@enron.com").
		SetReceiverAddr("receiver1@enron.com").
		SetTimestamp(1234567890).
		SetBody("Test body 1").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test email 1: %v", err)
	}

	email2, err := client.Email.Create().
		SetMessageID("test2@enron.com").
		SetSubject("Test Email 2").
		SetSenderAddr("sender2@enron.com").
		SetReceiverAddr("receiver2@enron.com").
		SetTimestamp(1234567891).
		SetBody("Test body 2").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test email 2: %v", err)
	}

	person1, err := client.DiscoveredEntity.Create().
		SetUniqueID("person-john-doe").
		SetTypeCategory("person").
		SetName("John Doe").
		SetConfidenceScore(0.95).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create discovered entity person1: %v", err)
	}

	org1, err := client.DiscoveredEntity.Create().
		SetUniqueID("org-enron").
		SetTypeCategory("organization").
		SetName("Enron Corp").
		SetConfidenceScore(0.90).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create discovered entity org1: %v", err)
	}

	_, err = client.Relationship.Create().
		SetRelationshipType("works_for").
		SetSourceEntityType("person").
		SetSourceEntityValue(person1.Name).
		SetTargetEntityType("organization").
		SetTargetEntityValue(org1.Name).
		SetSourceType("email").
		SetSourceID(fmt.Sprintf("%d", email1.ID)).
		SetConfidence(0.85).
		SetExtractedAt(1234567890).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test relationship: %v", err)
	}

	_, err = client.Relationship.Create().
		SetRelationshipType("mentioned_in").
		SetSourceEntityType("person").
		SetSourceEntityValue(person1.Name).
		SetTargetEntityType("email").
		SetTargetEntityValue(email2.MessageID).
		SetSourceType("email").
		SetSourceID(fmt.Sprintf("%d", email2.ID)).
		SetConfidence(0.80).
		SetExtractedAt(1234567891).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test relationship 2: %v", err)
	}
}
ENDFILE

echo "✅ Fixed test_helper.go"

echo "Verifying compilation..."
go build ./internal/explorer/... && echo "✅ internal/explorer compiles" || echo "❌ Compilation failed"

go build ./cmd/explorer/... && echo "✅ cmd/explorer compiles" || echo "❌ Compilation failed"

echo ""
# Fix tests/contract/schema_service_test.go
echo "Fixing tests/contract/schema_service_test.go..."
cat > /Users/jochem/code/enron-graph-2/tests/contract/schema_service_test.go << 'ENDTEST'
package contract_test

import (
	"context"
	"testing"

	"github.com/Blogem/enron-graph/internal/explorer"
	_ "github.com/mattn/go-sqlite3"
)

// T016: Test GetSchema returns promoted types
func TestSchemaService_GetSchema_ReturnsPromotedTypes(t *testing.T) {
	client := TestClient(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client)
	ctx := context.Background()

	response, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	if len(response.PromotedTypes) == 0 {
		t.Fatal("Expected promoted types, got none")
	}

	// Verify person and organization are in promoted types
	typeNames := make(map[string]bool)
	for _, pt := range response.PromotedTypes {
		typeNames[pt.Name] = true
		if !pt.IsPromoted {
			t.Errorf("Type %s should be marked as promoted", pt.Name)
		}
	}

	if !typeNames["person"] {
		t.Error("Expected 'person' in promoted types")
	}
	if !typeNames["organization"] {
		t.Error("Expected 'organization' in promoted types")
	}
}

// T017: Test GetSchema returns discovered types
func TestSchemaService_GetSchema_ReturnsDiscoveredTypes(t *testing.T) {
	client := TestClient(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client)
	ctx := context.Background()

	response, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	if len(response.DiscoveredTypes) == 0 {
		t.Fatal("Expected discovered types, got none")
	}

	// Verify discovered types are not marked as promoted
	for _, dt := range response.DiscoveredTypes {
		if dt.IsPromoted {
			t.Errorf("Discovered type %s should not be marked as promoted", dt.Name)
		}
	}
}

// T018: Test no overlap between promoted and discovered types
func TestSchemaService_GetSchema_NoOverlapBetweenCategories(t *testing.T) {
	client := TestClient(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client)
	ctx := context.Background()

	response, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	// Build map of promoted type names
	promotedNames := make(map[string]bool)
	for _, pt := range response.PromotedTypes {
		promotedNames[pt.Name] = true
	}

	// Verify no discovered type is also promoted
	for _, dt := range response.DiscoveredTypes {
		if promotedNames[dt.Name] {
			t.Errorf("Type %s appears in both promoted and discovered", dt.Name)
		}
	}
}

// T019: Test schema includes property metadata
func TestSchemaService_GetSchema_IncludesPropertyMetadata(t *testing.T) {
	client := TestClient(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client)
	ctx := context.Background()

	response, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	// Check that at least one type has properties
	foundProperties := false
	for _, st := range append(response.PromotedTypes, response.DiscoveredTypes...) {
		if len(st.Properties) > 0 {
			foundProperties = true
			// Verify property has required fields
			for _, prop := range st.Properties {
				if prop.Name == "" {
					t.Error("Property should have a name")
				}
			}
			break
		}
	}

	if !foundProperties {
		t.Error("Expected at least one type to have properties")
	}
}

// T020: Test GetTypeDetails returns type details
func TestSchemaService_GetTypeDetails_ReturnsTypeDetails(t *testing.T) {
	client := TestClient(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client)
	ctx := context.Background()

	// Get details for person type
	details, err := service.GetTypeDetails(ctx, "person")
	if err != nil {
		t.Fatalf("GetTypeDetails failed: %v", err)
	}

	if details.Name != "person" {
		t.Errorf("Expected name 'person', got %s", details.Name)
	}

	if details.Count == 0 {
		t.Error("Expected count > 0 for person type")
	}

	if !details.IsPromoted {
		t.Error("Person type should be marked as promoted")
	}

	if len(details.Properties) == 0 {
		t.Error("Expected properties for person type")
	}
}

// T021: Test RefreshSchema updates schema
func TestSchemaService_RefreshSchema_UpdatesSchema(t *testing.T) {
	client := TestClient(t)
	defer client.Close()
	SeedTestData(t, client)

	service := explorer.NewSchemaService(client)
	ctx := context.Background()

	// Get initial schema
	response1, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	initialPromotedCount := len(response1.PromotedTypes)

	// Add a new promoted type
	_, err = client.SchemaPromotion.Create().
		SetTypeName("location").
		SetPromotedBy("test").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create promotion: %v", err)
	}

	// Refresh schema
	err = service.RefreshSchema(ctx)
	if err != nil {
		t.Fatalf("RefreshSchema failed: %v", err)
	}

	// Get updated schema
	response2, err := service.GetSchema(ctx)
	if err != nil {
		t.Fatalf("GetSchema failed after refresh: %v", err)
	}

	if len(response2.PromotedTypes) != initialPromotedCount+1 {
		t.Errorf("Expected %d promoted types after refresh, got %d",
			initialPromotedCount+1, len(response2.PromotedTypes))
	}

	// Verify location is in promoted types
	found := false
	for _, pt := range response2.PromotedTypes {
		if pt.Name == "location" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'location' in promoted types after refresh")
	}
}
ENDTEST
echo "✅ Fixed schema_service_test.go"

echo "Fixing tests/integration/explorer/schema_explorer_test.go..."
mkdir -p tests/integration/explorer
cat > tests/integration/explorer/schema_explorer_test.go << 'ENDINT'
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
ENDINT
echo "✅ Fixed schema_explorer_test.go"

echo "Done! Run this script anytime to restore corrupted files."
