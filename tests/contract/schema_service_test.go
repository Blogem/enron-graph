package contract_test

import (
	"context"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/enttest"
	"github.com/Blogem/enron-graph/internal/explorer"
	_ "github.com/mattn/go-sqlite3"
)

func NewTestClient(t *testing.T) *ent.Client {
	opts := []enttest.Option{
		enttest.WithOptions(ent.Log(t.Log)),
	}
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1", opts...)
	return client
}

func NewTestClientPostgres(t *testing.T) *ent.Client {
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
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person schema promotion: %v", err)
	}

	_, err = client.SchemaPromotion.Create().
		SetTypeName("organization").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create organization schema promotion: %v", err)
	}

	_, err = client.Email.Create().
		SetMessageID("test1@enron.com").
		SetSubject("Test Email 1").
		SetFrom("sender1@enron.com").
		SetTo([]string{"receiver1@enron.com"}).
		SetBody("Test body 1").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test email 1: %v", err)
	}

	email2, err := client.Email.Create().
		SetMessageID("test2@enron.com").
		SetSubject("Test Email 2").
		SetFrom("sender2@enron.com").
		SetTo([]string{"receiver2@enron.com"}).
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

	// Create some unpromoted discovered types
	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("location-houston").
		SetTypeCategory("location").
		SetName("Houston").
		SetConfidenceScore(0.75).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create discovered entity location1: %v", err)
	}

	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("product-electricity").
		SetTypeCategory("product").
		SetName("Electricity Trading").
		SetConfidenceScore(0.70).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create discovered entity product1: %v", err)
	}

	_, err = client.Relationship.Create().
		SetType("works_for").
		SetFromType("person").
		SetFromID(person1.ID).
		SetToType("organization").
		SetToID(org1.ID).
		SetConfidenceScore(0.85).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test relationship: %v", err)
	}

	_, err = client.Relationship.Create().
		SetType("mentioned_in").
		SetFromType("person").
		SetFromID(person1.ID).
		SetToType("email").
		SetToID(email2.ID).
		SetConfidenceScore(0.80).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test relationship 2: %v", err)
	}
}

// T016: Test GetSchema returns promoted types
func TestSchemaService_GetSchema_ReturnsPromotedTypes(t *testing.T) {
	client := NewTestClient(t)
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
	client := NewTestClient(t)
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
	client := NewTestClient(t)
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
	client := NewTestClient(t)
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
	client := NewTestClient(t)
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
	client := NewTestClient(t)
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
