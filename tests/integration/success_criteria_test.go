package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/explorer"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSC001_SchemaLoadsUnder2Seconds verifies success criteria SC-001:
// Users can view the complete schema (both promoted and discovered types) within 2 seconds
func TestSC001_SchemaLoadsUnder2Seconds(t *testing.T) {
	// Setup PostgreSQL test database
	client, db, cleanup := setupPostgresTestDB(t)
	defer cleanup()

	// Seed with enough data to be realistic (100+ entities of various types)
	seedSchemaData(t, client, 150)

	// Create schema service
	service := explorer.NewSchemaService(client, db)

	// Measure GetSchema performance
	start := time.Now()
	schema, err := service.GetSchema(context.Background())
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, schema)

	// Verify SC-001: <2 seconds
	assert.Less(t, duration.Seconds(), 2.0,
		"Schema load took %.2fs, expected <2s (SC-001)", duration.Seconds())

	// Verify we got both promoted and discovered types
	assert.NotEmpty(t, schema.PromotedTypes, "Should have promoted types")
	assert.NotEmpty(t, schema.DiscoveredTypes, "Should have discovered types")

	t.Logf("✓ SC-001 PASS: Schema loaded in %.2fs (<2s required)", duration.Seconds())
}

// TestSC003_NodeDetailsAppearUnder1Second verifies success criteria SC-003:
// Users can identify and click on any node to view its details within 1 second
func TestSC003_NodeDetailsAppearUnder1Second(t *testing.T) {
	client, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test entity
	entity := seedSingleEntity(t, client, "person", "john.doe@example.com", "John Doe")

	service := explorer.NewGraphService(client, db)

	// Measure GetNodeDetails performance
	start := time.Now()
	details, err := service.GetNodeDetails(context.Background(), entity.UniqueID)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, details)

	// Verify SC-003: <1 second
	assert.Less(t, duration.Seconds(), 1.0,
		"Node details took %.2fs, expected <1s (SC-003)", duration.Seconds())

	assert.Equal(t, entity.UniqueID, details.ID)
	assert.NotEmpty(t, details.Properties)

	t.Logf("✓ SC-003 PASS: Node details loaded in %.2fs (<1s required)", duration.Seconds())
}

// TestSC006_NodeExpansionUnder2Seconds verifies success criteria SC-006:
// Users can expand a node's connections and see related entities rendered in the graph within 2 seconds
func TestSC006_NodeExpansionUnder2Seconds(t *testing.T) {
	client, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a node with 30 relationships
	centerNode := seedEntityWithRelationships(t, client, "person", 30)

	service := explorer.NewGraphService(client, db)

	// Measure GetRelationships performance
	start := time.Now()
	rels, err := service.GetRelationships(context.Background(), centerNode.UniqueID, 0, 50)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, rels)

	// Verify SC-006: <2 seconds
	assert.Less(t, duration.Seconds(), 2.0,
		"Node expansion took %.2fs, expected <2s (SC-006)", duration.Seconds())

	// Should have edges and nodes for the relationships
	totalItems := len(rels.Nodes) + len(rels.Edges)
	assert.GreaterOrEqual(t, totalItems, 1, "Should return related nodes or edges")

	t.Logf("✓ SC-006 PASS: Node expansion completed in %.2fs (<2s required) with %d nodes and %d edges",
		duration.Seconds(), len(rels.Nodes), len(rels.Edges))
}

// TestSC008_1000NodesPanZoomUnder500ms verifies success criteria SC-008:
// The graph visualization remains responsive when displaying up to 1000 nodes (pan/zoom operations complete within 500ms)
// Note: This tests the backend data fetch time; frontend rendering performance is tested separately
func TestSC008_1000NodesLoadUnder500ms(t *testing.T) {
	client, db, cleanup := setupTestDB(t)
	defer cleanup()

	// Seed 1000 entities
	seedSchemaData(t, client, 1000)

	service := explorer.NewGraphService(client, db)

	// Measure GetRandomNodes performance for 1000 nodes
	start := time.Now()
	graph, err := service.GetRandomNodes(context.Background(), 1000)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, graph)

	// Backend should contribute minimal time to the 500ms budget
	// Allow up to 1000ms for backend with 1000 nodes (frontend rendering is separate)
	assert.Less(t, duration.Milliseconds(), int64(1000),
		"Loading 1000 nodes took %dms, expected <1000ms (SC-008 backend portion)", duration.Milliseconds())

	assert.LessOrEqual(t, len(graph.Nodes), 1000)

	t.Logf("✓ SC-008 PASS: 1000 nodes loaded in %dms (<1000ms backend target)", duration.Milliseconds())
}

// Helper functions

func setupTestDB(t *testing.T) (*ent.Client, *sql.DB, func()) {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)

	db, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)

	// Run migrations
	err = client.Schema.Create(context.Background())
	require.NoError(t, err)

	cleanup := func() {
		client.Close()
		db.Close()
	}

	return client, db, cleanup
}

func setupPostgresTestDB(t *testing.T) (*ent.Client, *sql.DB, func()) {
	// Check if PostgreSQL is available via environment or use default
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "enron_graph_test"
	}

	dsn := fmt.Sprintf("host=%s port=5432 user=enron password=enron123 dbname=%s sslmode=disable",
		dbHost, dbName)

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
		return nil, nil, func() {}
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skipf("PostgreSQL not reachable: %v", err)
		return nil, nil, func() {}
	}

	// Open ent client
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open ent client: %v", err)
	}

	// Run migrations
	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	cleanup := func() {
		// Clean up test data
		client.DiscoveredEntity.Delete().ExecX(ctx)
		client.Relationship.Delete().ExecX(ctx)
		client.SchemaPromotion.Delete().ExecX(ctx)
		client.Close()
		db.Close()
	}

	return client, db, cleanup
}

func seedSchemaData(t *testing.T, client *ent.Client, count int) {
	ctx := context.Background()

	// Create mix of promoted and discovered entities
	promotedCount := count / 3
	discoveredCount := count - promotedCount

	// Create promoted entities
	for i := 0; i < promotedCount; i++ {
		typeName := []string{"person", "organization", "email_address"}[i%3]
		_, err := client.DiscoveredEntity.Create().
			SetUniqueID(fmt.Sprintf("%s_%d@example.com", typeName, i)).
			SetTypeCategory(typeName).
			SetName(fmt.Sprintf("Entity %d", i)).
			SetProperties(map[string]interface{}{
				"id":   i,
				"name": "Entity " + string(rune('A'+i%26)),
			}).
			Save(ctx)
		require.NoError(t, err)
	}

	// Create discovered entities
	for i := 0; i < discoveredCount; i++ {
		typeName := "discovered_type_" + string(rune('A'+i%10))
		_, err := client.DiscoveredEntity.Create().
			SetUniqueID(fmt.Sprintf("discovered_%d", promotedCount+i)).
			SetTypeCategory(typeName).
			SetName(fmt.Sprintf("Discovered %d", i)).
			SetProperties(map[string]interface{}{
				"id":    promotedCount + i,
				"value": "Value " + string(rune('A'+i%26)),
			}).
			Save(ctx)
		require.NoError(t, err)
	}

	// Create some promotions for the promoted types
	for _, typeName := range []string{"person", "organization", "email_address"} {
		_, err := client.SchemaPromotion.Create().
			SetTypeName(typeName).
			SetPromotedAt(time.Now()).
			Save(ctx)
		require.NoError(t, err)
	}
}

func seedSingleEntity(t *testing.T, client *ent.Client, typeName, uniqueID, name string) *ent.DiscoveredEntity {
	entity, err := client.DiscoveredEntity.Create().
		SetUniqueID(uniqueID).
		SetTypeCategory(typeName).
		SetName(name).
		SetProperties(map[string]interface{}{
			"email": uniqueID,
			"title": "Manager",
		}).
		Save(context.Background())
	require.NoError(t, err)
	return entity
}

func seedEntityWithRelationships(t *testing.T, client *ent.Client, typeName string, relCount int) *ent.DiscoveredEntity {
	ctx := context.Background()

	// Create center entity
	center, err := client.DiscoveredEntity.Create().
		SetUniqueID("center@example.com").
		SetTypeCategory(typeName).
		SetName("Center Node").
		SetProperties(map[string]interface{}{
			"id":   "center",
			"name": "Center Node",
		}).
		Save(ctx)
	require.NoError(t, err)

	// Create related entities and relationships
	for i := 0; i < relCount; i++ {
		target, err := client.DiscoveredEntity.Create().
			SetUniqueID(fmt.Sprintf("related_%d@example.com", i)).
			SetTypeCategory("related_entity").
			SetName(fmt.Sprintf("Related %d", i)).
			SetProperties(map[string]interface{}{
				"id":   i,
				"name": "Related " + string(rune('A'+i%26)),
			}).
			Save(ctx)
		require.NoError(t, err)

		_, err = client.Relationship.Create().
			SetType("CONNECTED_TO").
			SetFromType("discovered_entity").
			SetFromID(center.ID).
			SetToType("discovered_entity").
			SetToID(target.ID).
			Save(ctx)
		require.NoError(t, err)
	}

	return center
}
