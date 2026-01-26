package contract

import (
	"context"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/enttest"
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
		t.Fatalf("Failed to create person promotion: %v", err)
	}

	_, err = client.SchemaPromotion.Create().
		SetTypeName("organization").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create organization promotion: %v", err)
	}

	properties := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("person-john-doe").
		SetTypeCategory("person").
		SetName("John Doe").
		SetConfidenceScore(0.95).
		SetProperties(properties).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person entity: %v", err)
	}

	orgProperties := map[string]interface{}{
		"name":     "Enron Corp",
		"industry": "Energy",
		"founded":  1985,
	}

	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("org-enron").
		SetTypeCategory("organization").
		SetName("Enron Corp").
		SetConfidenceScore(0.90).
		SetProperties(orgProperties).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create organization entity: %v", err)
	}

	locationProperties := map[string]interface{}{
		"name":    "Houston",
		"state":   "Texas",
		"country": "USA",
	}

	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("location-houston").
		SetTypeCategory("location").
		SetName("Houston").
		SetConfidenceScore(0.75).
		SetProperties(locationProperties).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create location entity: %v", err)
	}

	productProperties := map[string]interface{}{
		"name":     "Energy Trading",
		"category": "Financial Product",
	}

	_, err = client.DiscoveredEntity.Create().
		SetUniqueID("product-energy").
		SetTypeCategory("product").
		SetName("Energy Trading").
		SetConfidenceScore(0.70).
		SetProperties(productProperties).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create product entity: %v", err)
	}

	t.Logf("Test data seeded: 2 promoted types, 4 entities (2 promoted, 2 discovered)")
}
