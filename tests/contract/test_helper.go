package contract

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/ent/enttest"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	testDBHost     = "localhost"
	testDBPort     = "5432"
	testDBUser     = "enron"
	testDBPassword = "enron123"
	testDBName     = "enron_contract_test"
)

// NewTestClientWithDB creates both an ent client and the underlying SQL DB for testing with PostgreSQL
func NewTestClientWithDB(t *testing.T) (*ent.Client, *sql.DB) {
	t.Helper()

	// Connect to postgres database to create test database
	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword)

	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer adminDB.Close()

	// Drop test database if it exists (cleanup from previous failed tests)
	_, _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))

	// Create fresh test database
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database
	testDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)

	testDB, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create ent client with schema migration
	client := enttest.Open(t, "postgres", testDSN)

	// Register cleanup function
	t.Cleanup(func() {
		client.Close()
		testDB.Close()

		// Drop test database
		_, _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	})

	return client, testDB
}

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

// SeedGraphTestData creates test data for graph service tests
func SeedGraphTestData(t *testing.T, client *ent.Client) {
	ctx := context.Background()

	// Create diverse set of entities for graph testing
	entities := []struct {
		uniqueID     string
		typeCategory string
		name         string
		confidence   float64
		properties   map[string]interface{}
	}{
		{
			uniqueID:     "person-jeff-skilling",
			typeCategory: "person",
			name:         "Jeff Skilling",
			confidence:   0.98,
			properties: map[string]interface{}{
				"email": "jeff.skilling@enron.com",
				"title": "CEO",
			},
		},
		{
			uniqueID:     "person-ken-lay",
			typeCategory: "person",
			name:         "Ken Lay",
			confidence:   0.97,
			properties: map[string]interface{}{
				"email": "ken.lay@enron.com",
				"title": "Chairman",
			},
		},
		{
			uniqueID:     "person-andy-fastow",
			typeCategory: "person",
			name:         "Andrew Fastow",
			confidence:   0.96,
			properties: map[string]interface{}{
				"email": "andrew.fastow@enron.com",
				"title": "CFO",
			},
		},
		{
			uniqueID:     "org-enron",
			typeCategory: "organization",
			name:         "Enron Corporation",
			confidence:   0.95,
			properties: map[string]interface{}{
				"industry": "Energy",
				"location": "Houston",
			},
		},
		{
			uniqueID:     "org-arthur-andersen",
			typeCategory: "organization",
			name:         "Arthur Andersen",
			confidence:   0.94,
			properties: map[string]interface{}{
				"industry": "Accounting",
			},
		},
		{
			uniqueID:     "concept-energy-trading",
			typeCategory: "concept",
			name:         "Energy Trading",
			confidence:   0.85,
			properties: map[string]interface{}{
				"category": "Business Activity",
			},
		},
		{
			uniqueID:     "location-houston",
			typeCategory: "location",
			name:         "Houston",
			confidence:   0.80,
			properties: map[string]interface{}{
				"state":   "Texas",
				"country": "USA",
			},
		},
	}

	createdEntities := make([]*ent.DiscoveredEntity, 0, len(entities))
	for _, e := range entities {
		entity, err := client.DiscoveredEntity.Create().
			SetUniqueID(e.uniqueID).
			SetTypeCategory(e.typeCategory).
			SetName(e.name).
			SetConfidenceScore(e.confidence).
			SetProperties(e.properties).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create entity %s: %v", e.name, err)
		}
		createdEntities = append(createdEntities, entity)
	}

	// Create relationships between entities
	relationships := []struct {
		fromUniqueID string
		toUniqueID   string
		relType      string
	}{
		{"person-jeff-skilling", "org-enron", "WORKED_AT"},
		{"person-ken-lay", "org-enron", "WORKED_AT"},
		{"person-andy-fastow", "org-enron", "WORKED_AT"},
		{"person-jeff-skilling", "person-ken-lay", "COMMUNICATED_WITH"},
		{"person-jeff-skilling", "concept-energy-trading", "MENTIONED"},
		{"org-enron", "location-houston", "LOCATED_IN"},
		{"org-enron", "org-arthur-andersen", "AUDITED_BY"},
	}

	// Build a map of unique_id to entity for lookups
	entityMap := make(map[string]*ent.DiscoveredEntity)
	for _, e := range createdEntities {
		entityMap[e.UniqueID] = e
	}

	for _, r := range relationships {
		fromEntity, ok1 := entityMap[r.fromUniqueID]
		toEntity, ok2 := entityMap[r.toUniqueID]

		if !ok1 || !ok2 {
			t.Logf("Warning: Skipping relationship %s -> %s (entity not found)", r.fromUniqueID, r.toUniqueID)
			continue
		}

		_, err := client.Relationship.Create().
			SetType(r.relType).
			SetFromType("discovered_entity").
			SetFromID(fromEntity.ID).
			SetToType("discovered_entity").
			SetToID(toEntity.ID).
			SetConfidenceScore(1.0).
			Save(ctx)
		if err != nil {
			t.Logf("Warning: Failed to create relationship %s -> %s: %v", r.fromUniqueID, r.toUniqueID, err)
		}
	}

	t.Logf("Graph test data seeded: %d entities, %d relationships", len(entities), len(relationships))
}

// SeedNodeWithManyRelationships creates a node with a specific number of relationships
func SeedNodeWithManyRelationships(t *testing.T, client *ent.Client, count int) string {
	ctx := context.Background()

	// Create the central node
	centralNode, err := client.DiscoveredEntity.Create().
		SetUniqueID("central-node").
		SetTypeCategory("person").
		SetName("Central Person").
		SetConfidenceScore(0.95).
		SetProperties(map[string]interface{}{
			"email": "central@enron.com",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create central node: %v", err)
	}

	// Create target nodes and relationships
	for i := 0; i < count; i++ {
		targetID := fmt.Sprintf("target-%d", i)

		targetEntity, err := client.DiscoveredEntity.Create().
			SetUniqueID(targetID).
			SetTypeCategory("person").
			SetName(fmt.Sprintf("Person %d", i)).
			SetConfidenceScore(0.80).
			SetProperties(map[string]interface{}{
				"email": fmt.Sprintf("person%d@enron.com", i),
			}).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create target node %d: %v", i, err)
		}

		_, err = client.Relationship.Create().
			SetType("COMMUNICATED_WITH").
			SetFromType("discovered_entity").
			SetFromID(centralNode.ID).
			SetToType("discovered_entity").
			SetToID(targetEntity.ID).
			SetConfidenceScore(1.0).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create relationship %d: %v", i, err)
		}
	}

	t.Logf("Created node with %d relationships", count)
	return centralNode.UniqueID
}

// SeedMixedTypeGraphData creates test data with multiple types and categories for filtering tests
func SeedMixedTypeGraphData(t *testing.T, client *ent.Client) {
	ctx := context.Background()

	// Create diverse entities for filtering tests
	entities := []struct {
		uniqueID     string
		typeCategory string
		name         string
		confidence   float64
		properties   map[string]interface{}
	}{
		{
			uniqueID:     "person-jeff-skilling",
			typeCategory: "person",
			name:         "Jeff Skilling",
			confidence:   0.98,
			properties: map[string]interface{}{
				"email": "jeff.skilling@enron.com",
				"title": "CEO",
			},
		},
		{
			uniqueID:     "person-ken-lay",
			typeCategory: "person",
			name:         "Ken Lay",
			confidence:   0.97,
			properties: map[string]interface{}{
				"email": "ken.lay@enron.com",
				"title": "Chairman",
			},
		},
		{
			uniqueID:     "person-andy-fastow",
			typeCategory: "person",
			name:         "Andrew Fastow",
			confidence:   0.96,
			properties: map[string]interface{}{
				"email": "andrew.fastow@enron.com",
				"title": "CFO",
			},
		},
		{
			uniqueID:     "org-enron",
			typeCategory: "organization",
			name:         "Enron Corporation",
			confidence:   0.95,
			properties: map[string]interface{}{
				"industry": "Energy",
				"location": "Houston",
			},
		},
		{
			uniqueID:     "org-arthur-andersen",
			typeCategory: "organization",
			name:         "Arthur Andersen",
			confidence:   0.94,
			properties: map[string]interface{}{
				"industry": "Accounting",
				"website":  "www.arthurandersen.com",
			},
		},
		{
			uniqueID:     "location-houston",
			typeCategory: "location",
			name:         "Houston",
			confidence:   0.80,
			properties: map[string]interface{}{
				"state":   "Texas",
				"country": "USA",
			},
		},
		{
			uniqueID:     "concept-energy-trading",
			typeCategory: "concept",
			name:         "Energy Trading",
			confidence:   0.85,
			properties: map[string]interface{}{
				"category": "Business Activity",
			},
		},
	}

	for _, e := range entities {
		_, err := client.DiscoveredEntity.Create().
			SetUniqueID(e.uniqueID).
			SetTypeCategory(e.typeCategory).
			SetName(e.name).
			SetConfidenceScore(e.confidence).
			SetProperties(e.properties).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create entity %s: %v", e.name, err)
		}
	}

	t.Logf("Mixed type graph data seeded: %d entities (3 person, 2 org, 1 location, 1 concept)", len(entities))
}
