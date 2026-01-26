package contract

import (
	"context"
	"strings"
	"testing"

	"github.com/Blogem/enron-graph/internal/explorer"
)

// T043: Test GetRandomNodes returns limited nodes
func TestGraphService_GetRandomNodes_ReturnsLimitedNodes(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedGraphTestData(t, client)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// When: GetRandomNodes(100)
	resp, err := service.GetRandomNodes(ctx, 100)

	// Then: Returns up to 100 nodes without error
	if err != nil {
		t.Fatalf("GetRandomNodes failed: %v", err)
	}

	if len(resp.Nodes) > 100 {
		t.Errorf("Expected at most 100 nodes, got %d", len(resp.Nodes))
	}

	if len(resp.Nodes) == 0 {
		t.Error("Expected at least some nodes if DB has data")
	}

	// Verify TotalNodes is set
	if resp.TotalNodes == 0 {
		t.Error("Expected TotalNodes to be set")
	}
}

// T044: Test GetRandomNodes validates node fields
func TestGraphService_GetRandomNodes_ValidatesNodeFields(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedGraphTestData(t, client)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	resp, err := service.GetRandomNodes(ctx, 100)
	if err != nil {
		t.Fatalf("GetRandomNodes failed: %v", err)
	}

	// Each node has required fields
	for i, node := range resp.Nodes {
		if node.ID == "" {
			t.Errorf("Node %d: ID is required", i)
		}
		if node.Type == "" {
			t.Errorf("Node %d: Type is required", i)
		}
		if node.Category != "promoted" && node.Category != "discovered" {
			t.Errorf("Node %d: Category must be 'promoted' or 'discovered', got '%s'", i, node.Category)
		}
		if node.Properties == nil {
			t.Errorf("Node %d: Properties map is required (can be empty)", i)
		}
	}
}

// T045: Test GetRandomNodes edges reference returned nodes
func TestGraphService_GetRandomNodes_EdgesReferenceReturnedNodes(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedGraphTestData(t, client)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	resp, err := service.GetRandomNodes(ctx, 100)
	if err != nil {
		t.Fatalf("GetRandomNodes failed: %v", err)
	}

	// Build map of returned node IDs
	nodeIDs := make(map[string]bool)
	for _, node := range resp.Nodes {
		nodeIDs[node.ID] = true
	}

	// Verify edges reference existing nodes
	for i, edge := range resp.Edges {
		if !nodeIDs[edge.Source] {
			t.Errorf("Edge %d: Source '%s' does not reference a returned node", i, edge.Source)
		}
		if !nodeIDs[edge.Target] {
			t.Errorf("Edge %d: Target '%s' does not reference a returned node", i, edge.Target)
		}
		if edge.Type == "" {
			t.Errorf("Edge %d: Type is required", i)
		}
	}
}

// T046: Test GetRelationships paginates correctly
func TestGraphService_GetRelationships_PaginatesCorrectly(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()

	// Create a node with exactly 120 relationships
	nodeID := SeedNodeWithManyRelationships(t, client, 120)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// First batch: offset=0, limit=50
	resp1, err := service.GetRelationships(ctx, nodeID, 0, 50)
	if err != nil {
		t.Fatalf("GetRelationships (batch 1) failed: %v", err)
	}

	if len(resp1.Edges) != 50 {
		t.Errorf("Expected exactly 50 edges in first batch, got %d", len(resp1.Edges))
	}
	if resp1.TotalCount != 120 {
		t.Errorf("Expected TotalCount=120, got %d", resp1.TotalCount)
	}
	if !resp1.HasMore {
		t.Error("Expected HasMore=true for first batch")
	}
	if resp1.Offset != 0 {
		t.Errorf("Expected Offset=0, got %d", resp1.Offset)
	}

	// Second batch: offset=50, limit=50
	resp2, err := service.GetRelationships(ctx, nodeID, 50, 50)
	if err != nil {
		t.Fatalf("GetRelationships (batch 2) failed: %v", err)
	}

	if len(resp2.Edges) != 50 {
		t.Errorf("Expected exactly 50 edges in second batch, got %d", len(resp2.Edges))
	}
	if resp2.TotalCount != 120 {
		t.Errorf("Expected TotalCount=120, got %d", resp2.TotalCount)
	}
	if !resp2.HasMore {
		t.Error("Expected HasMore=true for second batch")
	}
	if resp2.Offset != 50 {
		t.Errorf("Expected Offset=50, got %d", resp2.Offset)
	}
}

// T047: Test GetRelationships handles final batch
func TestGraphService_GetRelationships_HandlesFinalBatch(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()

	// Create a node with exactly 120 relationships
	nodeID := SeedNodeWithManyRelationships(t, client, 120)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// Final batch: offset=100, limit=50
	resp, err := service.GetRelationships(ctx, nodeID, 100, 50)
	if err != nil {
		t.Fatalf("GetRelationships (final batch) failed: %v", err)
	}

	if len(resp.Edges) != 20 {
		t.Errorf("Expected 20 edges in final batch (120 - 100), got %d", len(resp.Edges))
	}
	if resp.TotalCount != 120 {
		t.Errorf("Expected TotalCount=120, got %d", resp.TotalCount)
	}
	if resp.HasMore {
		t.Error("Expected HasMore=false for final batch")
	}
	if resp.Offset != 100 {
		t.Errorf("Expected Offset=100, got %d", resp.Offset)
	}

	// Verify connected nodes are included
	if len(resp.Nodes) == 0 {
		t.Error("Expected connected nodes to be included in response")
	}

	// Verify nodes match edge targets
	nodeIDs := make(map[string]bool)
	for _, node := range resp.Nodes {
		nodeIDs[node.ID] = true
	}

	for _, edge := range resp.Edges {
		// Either source or target should be in the returned nodes
		// (source is the expanded node, targets are the connected nodes)
		if edge.Source != nodeID && !nodeIDs[edge.Source] {
			if !nodeIDs[edge.Target] {
				t.Errorf("Edge has neither source nor target in returned nodes")
			}
		}
	}
}

// T048: Test GetNodeDetails returns complete info
func TestGraphService_GetNodeDetails_ReturnsCompleteInfo(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedGraphTestData(t, client)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// Get first entity to use as test subject
	entities, err := client.DiscoveredEntity.Query().Limit(1).All(ctx)
	if err != nil || len(entities) == 0 {
		t.Fatal("Failed to get test entity")
	}

	nodeID := entities[0].UniqueID

	// When: GetNodeDetails
	node, err := service.GetNodeDetails(ctx, nodeID)
	if err != nil {
		t.Fatalf("GetNodeDetails failed: %v", err)
	}

	// Then: Returns complete node data
	if node.ID != nodeID {
		t.Errorf("Expected ID=%s, got %s", nodeID, node.ID)
	}
	if node.Type == "" {
		t.Error("Expected Type to be set")
	}
	if node.Category == "" {
		t.Error("Expected Category to be set")
	}
	if node.Properties == nil {
		t.Error("Expected Properties map to be set")
	}
	if node.IsGhost {
		t.Error("Real node should not be marked as ghost")
	}

	// Verify properties match entity
	if node.Type != entities[0].TypeCategory {
		t.Errorf("Expected Type=%s, got %s", entities[0].TypeCategory, node.Type)
	}
}

// TestGraphService_GetNodeDetails_ErrorsOnMissingNode verifies error handling
func TestGraphService_GetNodeDetails_ErrorsOnMissingNode(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// When: GetNodeDetails with non-existent ID
	_, err := service.GetNodeDetails(ctx, "non-existent-id-12345")

	// Then: Returns error
	if err == nil {
		t.Error("Expected error for non-existent node")
	}

	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// T076: Test GetNodes filters by type
func TestGraphService_GetNodes_FiltersByType(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedMixedTypeGraphData(t, client)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// When: Filter by types=["person"]
	filter := explorer.NodeFilter{
		Types: []string{"person"},
		Limit: 100,
	}
	resp, err := service.GetNodes(ctx, filter)

	// Then: Returns only person nodes
	if err != nil {
		t.Fatalf("GetNodes failed: %v", err)
	}

	if len(resp.Nodes) == 0 {
		t.Error("Expected at least some person nodes")
	}

	for i, node := range resp.Nodes {
		if node.Type != "person" {
			t.Errorf("Node %d: Expected type 'person', got '%s'", i, node.Type)
		}
	}
}

// T077: Test GetNodes filters by category
func TestGraphService_GetNodes_FiltersByCategory(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedMixedTypeGraphData(t, client)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// When: Filter by category="discovered"
	filter := explorer.NodeFilter{
		Category: "discovered",
		Limit:    100,
	}
	resp, err := service.GetNodes(ctx, filter)

	// Then: Returns only discovered nodes
	if err != nil {
		t.Fatalf("GetNodes failed: %v", err)
	}

	if len(resp.Nodes) == 0 {
		t.Error("Expected at least some discovered nodes")
	}

	for i, node := range resp.Nodes {
		if node.Category != "discovered" {
			t.Errorf("Node %d: Expected category 'discovered', got '%s'", i, node.Category)
		}
	}
}

// T078: Test GetNodes searches by property value
func TestGraphService_GetNodes_SearchesByPropertyValue(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedMixedTypeGraphData(t, client)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// When: Search for "skilling@enron.com"
	filter := explorer.NodeFilter{
		SearchQuery: "skilling@enron.com",
		Limit:       100,
	}
	resp, err := service.GetNodes(ctx, filter)

	// Then: Returns nodes with matching property values
	if err != nil {
		t.Fatalf("GetNodes failed: %v", err)
	}

	if len(resp.Nodes) == 0 {
		t.Error("Expected at least one matching node")
	}

	// At least one node should contain the search term
	foundMatch := false
	for _, node := range resp.Nodes {
		for _, value := range node.Properties {
			if str, ok := value.(string); ok && strings.Contains(strings.ToLower(str), strings.ToLower("skilling@enron.com")) {
				foundMatch = true
				break
			}
		}
		if foundMatch {
			break
		}
	}

	if !foundMatch {
		t.Error("Expected at least one node to match search query")
	}
}

// T079: Test GetNodes combines multiple filters
func TestGraphService_GetNodes_CombinesMultipleFilters(t *testing.T) {
	client, db := NewTestClientWithDB(t)
	defer client.Close()
	SeedMixedTypeGraphData(t, client)

	service := explorer.NewGraphService(client, db)
	ctx := context.Background()

	// When: Filter by type="person" AND search="enron.com"
	filter := explorer.NodeFilter{
		Types:       []string{"person"},
		SearchQuery: "enron.com",
		Limit:       100,
	}
	resp, err := service.GetNodes(ctx, filter)

	// Then: Returns only person nodes matching search query
	if err != nil {
		t.Fatalf("GetNodes failed: %v", err)
	}

	for i, node := range resp.Nodes {
		// Must be person type
		if node.Type != "person" {
			t.Errorf("Node %d: Expected type 'person', got '%s'", i, node.Type)
		}

		// Must contain search term in at least one property
		foundMatch := false
		for _, value := range node.Properties {
			if str, ok := value.(string); ok && strings.Contains(strings.ToLower(str), "enron.com") {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			t.Errorf("Node %d: Expected to match search query 'enron.com'", i)
		}
	}
}
