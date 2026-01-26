package explorer_test

import (
	"context"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/explorer"
	integration "github.com/Blogem/enron-graph/tests/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDetailTestService(t *testing.T) *explorer.GraphService {
	client, db := integration.SetupTestDBWithSQL(t)
	seedDetailTestData(t, client)
	return explorer.NewGraphService(client, db)
}

func seedDetailTestData(t *testing.T, client *ent.Client) {
	ctx := context.Background()

	// Create a discovered entity with comprehensive properties including metadata
	person1, err := client.DiscoveredEntity.Create().
		SetUniqueID("person-alice-detailed").
		SetTypeCategory("person").
		SetName("Alice Smith").
		SetConfidenceScore(0.95).
		SetProperties(map[string]interface{}{
			// User properties (T102)
			"email":       "alice@example.com",
			"role":        "Senior Manager",
			"department":  "Engineering",
			"phone":       "+1-555-0123",
			"location":    "San Francisco, CA",
			"employee_id": "EMP-12345",
			// Metadata properties (T104)
			"confidence":    0.95,
			"discovered_at": time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			"created_at":    time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
			"last_seen":     time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"source":        "email_analyzer",
		}).
		Save(ctx)
	require.NoError(t, err, "Failed to create person-alice-detailed")

	// Create a person with minimal properties for comparison
	person2, err := client.DiscoveredEntity.Create().
		SetUniqueID("person-bob-minimal").
		SetTypeCategory("person").
		SetName("Bob Johnson").
		SetConfidenceScore(0.90).
		SetProperties(map[string]interface{}{
			"email": "bob@example.com",
			"role":  "Developer",
		}).
		Save(ctx)
	require.NoError(t, err, "Failed to create person-bob-minimal")

	// Create organization entities as related entities (T103)
	org1, err := client.DiscoveredEntity.Create().
		SetUniqueID("org-techcorp").
		SetTypeCategory("organization").
		SetName("TechCorp Inc").
		SetConfidenceScore(0.88).
		SetProperties(map[string]interface{}{
			"industry": "Technology",
			"founded":  "2010",
			"size":     "1000-5000",
		}).
		Save(ctx)
	require.NoError(t, err, "Failed to create org-techcorp")

	org2, err := client.DiscoveredEntity.Create().
		SetUniqueID("org-opensource").
		SetTypeCategory("project").
		SetName("OpenSource Initiative").
		SetConfidenceScore(0.85).
		SetProperties(map[string]interface{}{
			"type":        "open-source",
			"description": "Community-driven project",
		}).
		Save(ctx)
	require.NoError(t, err, "Failed to create org-opensource")

	// Create multiple relationships to test related entities list (T103)
	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person1.ID).
		SetToType("discovered_entity").
		SetToID(org1.ID).
		SetType("works_at").
		SetConfidenceScore(0.92).
		Save(ctx)
	require.NoError(t, err, "Failed to create relationship person1->org1")

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person1.ID).
		SetToType("discovered_entity").
		SetToID(org2.ID).
		SetType("contributes_to").
		SetConfidenceScore(0.87).
		Save(ctx)
	require.NoError(t, err, "Failed to create relationship person1->org2")

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person2.ID).
		SetToType("discovered_entity").
		SetToID(person1.ID).
		SetType("reports_to").
		SetConfidenceScore(0.90).
		Save(ctx)
	require.NoError(t, err, "Failed to create relationship person2->person1")

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person1.ID).
		SetToType("discovered_entity").
		SetToID(person2.ID).
		SetType("manages").
		SetConfidenceScore(0.90).
		Save(ctx)
	require.NoError(t, err, "Failed to create relationship person1->person2")
}

// T102: Integration test - Detail panel shows all properties for selected node
func TestDetailPanel_ShowsAllProperties(t *testing.T) {
	service := setupDetailTestService(t)
	ctx := context.Background()

	// Get node details for the node with comprehensive properties
	nodeDetails, err := service.GetNodeDetails(ctx, "person-alice-detailed")
	require.NoError(t, err, "GetNodeDetails should succeed")
	require.NotNil(t, nodeDetails, "Node details should not be nil")

	// Verify basic node information is present
	assert.Equal(t, "person-alice-detailed", nodeDetails.ID, "Node ID should match")
	assert.Equal(t, "person", nodeDetails.Type, "Node type should be 'person'")
	assert.Equal(t, "discovered", nodeDetails.Category, "Node category should be 'discovered'")
	assert.NotNil(t, nodeDetails.Properties, "Properties map should not be nil")

	// Verify all user properties are present (T102)
	expectedUserProperties := []string{
		"email", "role", "department", "phone", "location", "employee_id",
	}
	for _, prop := range expectedUserProperties {
		assert.Contains(t, nodeDetails.Properties, prop, "Property %s should be present", prop)
		assert.NotEmpty(t, nodeDetails.Properties[prop], "Property %s should have a value", prop)
	}

	// Verify specific property values
	assert.Equal(t, "alice@example.com", nodeDetails.Properties["email"])
	assert.Equal(t, "Senior Manager", nodeDetails.Properties["role"])
	assert.Equal(t, "Engineering", nodeDetails.Properties["department"])
	assert.Equal(t, "+1-555-0123", nodeDetails.Properties["phone"])
	assert.Equal(t, "San Francisco, CA", nodeDetails.Properties["location"])
	assert.Equal(t, "EMP-12345", nodeDetails.Properties["employee_id"])

	// Verify metadata properties are also present
	assert.Contains(t, nodeDetails.Properties, "confidence", "Metadata 'confidence' should be present")
	assert.Contains(t, nodeDetails.Properties, "discovered_at", "Metadata 'discovered_at' should be present")
	assert.Contains(t, nodeDetails.Properties, "source", "Metadata 'source' should be present")
}

// T103: Integration test - Detail panel shows related entities list
func TestDetailPanel_ShowsRelatedEntities(t *testing.T) {
	service := setupDetailTestService(t)
	ctx := context.Background()

	// Get node details for Alice who has multiple relationships
	nodeDetails, err := service.GetNodeDetails(ctx, "person-alice-detailed")
	require.NoError(t, err, "GetNodeDetails should succeed")
	require.NotNil(t, nodeDetails, "Node details should not be nil")

	// Get the relationships for this node
	relResponse, err := service.GetRelationships(ctx, "person-alice-detailed", 0, 50)
	require.NoError(t, err, "GetRelationships should succeed")
	require.NotNil(t, relResponse, "Relationships response should not be nil")

	// Verify we have related entities
	assert.Greater(t, len(relResponse.Nodes), 0, "Should have related entities")
	assert.Greater(t, len(relResponse.Edges), 0, "Should have relationship edges")
	assert.Equal(t, 4, relResponse.TotalCount, "Should have 4 total relationships (works_at, contributes_to, manages, reports_to)")

	// Verify related entities contain expected nodes
	relatedNodeIDs := make(map[string]bool)
	for _, node := range relResponse.Nodes {
		relatedNodeIDs[node.ID] = true
	}

	// Alice is connected to org-techcorp, org-opensource, and person-bob-minimal
	expectedRelatedIDs := []string{"org-techcorp", "org-opensource", "person-bob-minimal"}
	for _, expectedID := range expectedRelatedIDs {
		assert.True(t, relatedNodeIDs[expectedID], "Should include related entity %s", expectedID)
	}

	// Verify edges contain relationship types
	edgeTypes := make(map[string]bool)
	for _, edge := range relResponse.Edges {
		edgeTypes[edge.Type] = true
	}

	expectedEdgeTypes := []string{"works_at", "contributes_to", "manages", "reports_to"}
	for _, expectedType := range expectedEdgeTypes {
		assert.True(t, edgeTypes[expectedType], "Should include edge type %s", expectedType)
	}

	// Verify each related entity has complete information
	for _, node := range relResponse.Nodes {
		assert.NotEmpty(t, node.ID, "Related node should have an ID")
		assert.NotEmpty(t, node.Type, "Related node should have a type")
		assert.Contains(t, []string{"discovered", "promoted"}, node.Category, "Related node should have valid category")
		assert.NotNil(t, node.Properties, "Related node should have properties")
	}
}

// T104: Integration test - Discovered entity shows metadata in detail panel
func TestDetailPanel_ShowsDiscoveredEntityMetadata(t *testing.T) {
	service := setupDetailTestService(t)
	ctx := context.Background()

	// Get node details for a discovered entity
	nodeDetails, err := service.GetNodeDetails(ctx, "person-alice-detailed")
	require.NoError(t, err, "GetNodeDetails should succeed")
	require.NotNil(t, nodeDetails, "Node details should not be nil")

	// Verify the node is categorized as discovered
	assert.Equal(t, "discovered", nodeDetails.Category, "Node should be categorized as 'discovered'")

	// Verify metadata fields are present in properties (T104)
	metadataFields := []string{
		"confidence",
		"discovered_at",
		"created_at",
		"last_seen",
		"source",
	}

	for _, field := range metadataFields {
		assert.Contains(t, nodeDetails.Properties, field, "Metadata field %s should be present", field)
		assert.NotNil(t, nodeDetails.Properties[field], "Metadata field %s should have a value", field)
	}

	// Verify specific metadata values
	confidence, ok := nodeDetails.Properties["confidence"].(float64)
	assert.True(t, ok, "Confidence should be a float64")
	assert.Equal(t, 0.95, confidence, "Confidence should be 0.95")

	source, ok := nodeDetails.Properties["source"].(string)
	assert.True(t, ok, "Source should be a string")
	assert.Equal(t, "email_analyzer", source, "Source should be 'email_analyzer'")

	// Verify timestamp fields are present and formatted correctly
	discoveredAt, ok := nodeDetails.Properties["discovered_at"].(string)
	assert.True(t, ok, "discovered_at should be a string")
	assert.NotEmpty(t, discoveredAt, "discovered_at should not be empty")
	// Verify it's a valid RFC3339 timestamp
	_, err = time.Parse(time.RFC3339, discoveredAt)
	assert.NoError(t, err, "discovered_at should be a valid RFC3339 timestamp")

	createdAt, ok := nodeDetails.Properties["created_at"].(string)
	assert.True(t, ok, "created_at should be a string")
	assert.NotEmpty(t, createdAt, "created_at should not be empty")
	_, err = time.Parse(time.RFC3339, createdAt)
	assert.NoError(t, err, "created_at should be a valid RFC3339 timestamp")

	lastSeen, ok := nodeDetails.Properties["last_seen"].(string)
	assert.True(t, ok, "last_seen should be a string")
	assert.NotEmpty(t, lastSeen, "last_seen should not be empty")
	_, err = time.Parse(time.RFC3339, lastSeen)
	assert.NoError(t, err, "last_seen should be a valid RFC3339 timestamp")

	// Verify degree information is included
	assert.GreaterOrEqual(t, nodeDetails.Degree, 0, "Degree should be non-negative")
	// Alice has 3 relationships, so degree should be at least 3
	assert.GreaterOrEqual(t, nodeDetails.Degree, 3, "Alice should have at least 3 relationships")
}
