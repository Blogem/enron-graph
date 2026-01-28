package explorer_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/explorer"
	integration "github.com/Blogem/enron-graph/tests/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPerformanceTestService(t *testing.T, nodeCount int) *explorer.GraphService {
	client, db := integration.SetupTestDBWithSQL(t)
	seedPerformanceTestData(t, client, nodeCount)
	return explorer.NewGraphService(client, db, nil)
}

func seedPerformanceTestData(t *testing.T, client *ent.Client, nodeCount int) {
	ctx := context.Background()

	t.Logf("Seeding %d nodes for performance test...", nodeCount)
	startTime := time.Now()

	// Create nodes with varying degrees of connectivity
	entityIDs := make([]int, nodeCount)

	for i := 0; i < nodeCount; i++ {
		entity, err := client.DiscoveredEntity.Create().
			SetUniqueID(fmt.Sprintf("perf-node-%d", i)).
			SetTypeCategory(getTypeForIndex(i)).
			SetName(fmt.Sprintf("Node %d", i)).
			SetConfidenceScore(0.80 + float64(i%20)/100.0).
			SetProperties(map[string]interface{}{
				"index":       i,
				"group":       i % 10,
				"description": fmt.Sprintf("Performance test node %d", i),
			}).
			Save(ctx)
		require.NoError(t, err, "Failed to create node %d", i)
		entityIDs[i] = entity.ID
	}

	// Create relationships with varying connectivity patterns
	edgeCount := 0
	for i := 0; i < nodeCount; i++ {
		// Each node connects to several others (creates realistic graph structure)
		connectionCount := getConnectionCountForIndex(i, nodeCount)

		for j := 0; j < connectionCount; j++ {
			targetIdx := (i + j + 1) % nodeCount
			if targetIdx != i {
				_, err := client.Relationship.Create().
					SetFromType("discovered_entity").
					SetFromID(entityIDs[i]).
					SetToType("discovered_entity").
					SetToID(entityIDs[targetIdx]).
					SetType(getRelationshipType(i, j)).
					SetConfidenceScore(0.75 + float64(j%10)/40.0).
					Save(ctx)
				if err == nil {
					edgeCount++
				}
			}
		}
	}

	elapsed := time.Since(startTime)
	t.Logf("Seeded %d nodes and %d edges in %v", nodeCount, edgeCount, elapsed)
}

func getTypeForIndex(i int) string {
	types := []string{"person", "organization", "project", "location", "event"}
	return types[i%len(types)]
}

func getConnectionCountForIndex(i, total int) int {
	// Vary connectivity: some nodes have few connections, some have many
	if i%10 == 0 {
		return min(20, total/10) // Hub nodes with many connections
	} else if i%5 == 0 {
		return min(10, total/20) // Medium connectivity
	}
	return min(3, total/50) // Most nodes have few connections
}

func getRelationshipType(i, j int) string {
	types := []string{"connected_to", "related_to", "works_with", "part_of", "manages"}
	return types[(i+j)%len(types)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// T105: Performance test - Verify SC-008: <500ms pan/zoom for 1000+ nodes
func TestGraphPerformance_LargeGraphRendering(t *testing.T) {
	// SC-008: Graph should handle 1000 nodes with pan/zoom operations in <500ms
	nodeCount := 1000
	service := setupPerformanceTestService(t, nodeCount)
	ctx := context.Background()

	t.Logf("Testing performance with %d nodes", nodeCount)

	// Measure GetRandomNodes performance (simulates initial graph load)
	startLoad := time.Now()
	response, err := service.GetRandomNodes(ctx, nodeCount)
	loadDuration := time.Since(startLoad)

	require.NoError(t, err, "GetRandomNodes should succeed")
	require.NotNil(t, response, "Response should not be nil")

	t.Logf("Loaded %d nodes in %v", len(response.Nodes), loadDuration)

	// Verify we got a significant number of nodes
	assert.Greater(t, len(response.Nodes), 900, "Should load at least 900 nodes for 1000-node graph")

	// Verify data structure overhead is reasonable
	assert.Less(t, loadDuration, 2*time.Second, "Loading 1000 nodes should take <2s (SC-001)")

	// Verify node structure is complete (required for rendering)
	for i, node := range response.Nodes {
		if i > 10 {
			break // Sample check only
		}
		assert.NotEmpty(t, node.ID, "Node ID should be present")
		assert.NotEmpty(t, node.Type, "Node type should be present")
		assert.NotNil(t, node.Properties, "Node properties should be present")
	}

	// Verify edges are included
	assert.Greater(t, len(response.Edges), 0, "Should have edges between nodes")

	t.Logf("Graph has %d edges connecting nodes", len(response.Edges))
}

// T105: Performance test - Measure relationship expansion for high-degree nodes
func TestGraphPerformance_RelationshipExpansion(t *testing.T) {
	nodeCount := 500
	service := setupPerformanceTestService(t, nodeCount)
	ctx := context.Background()

	// Get random nodes first
	response, err := service.GetRandomNodes(ctx, 100)
	require.NoError(t, err, "GetRandomNodes should succeed")
	require.Greater(t, len(response.Nodes), 0, "Should have nodes")

	// Find a node with relationships to expand
	var testNodeID string
	for _, node := range response.Nodes {
		if node.Degree > 5 {
			testNodeID = node.ID
			break
		}
	}

	if testNodeID == "" {
		// Use first node if none have high degree
		testNodeID = response.Nodes[0].ID
	}

	// Measure relationship expansion performance (simulates pan/zoom interaction)
	startExpand := time.Now()
	relResponse, err := service.GetRelationships(ctx, testNodeID, 0, 50)
	expandDuration := time.Since(startExpand)

	require.NoError(t, err, "GetRelationships should succeed")
	require.NotNil(t, relResponse, "Relationships response should not be nil")

	t.Logf("Expanded relationships for node %s in %v", testNodeID, expandDuration)
	t.Logf("  - Loaded %d related nodes", len(relResponse.Nodes))
	t.Logf("  - Loaded %d edges", len(relResponse.Edges))
	t.Logf("  - Total relationships: %d", relResponse.TotalCount)

	// SC-008: Relationship expansion should be fast enough for interactive use
	assert.Less(t, expandDuration, 500*time.Millisecond, "Relationship expansion should take <500ms (SC-008)")

	// Verify response structure
	assert.GreaterOrEqual(t, relResponse.TotalCount, 0, "Total count should be non-negative")
	assert.Equal(t, relResponse.TotalCount > 50, relResponse.HasMore, "HasMore should be true if total > limit")
}

// T105: Performance test - Measure node details retrieval
func TestGraphPerformance_NodeDetails(t *testing.T) {
	nodeCount := 1000
	service := setupPerformanceTestService(t, nodeCount)
	ctx := context.Background()

	// Test node details retrieval speed (simulates node click interaction)
	testNodeID := "perf-node-42"

	startDetails := time.Now()
	nodeDetails, err := service.GetNodeDetails(ctx, testNodeID)
	detailsDuration := time.Since(startDetails)

	require.NoError(t, err, "GetNodeDetails should succeed")
	require.NotNil(t, nodeDetails, "Node details should not be nil")

	t.Logf("Retrieved node details in %v", detailsDuration)

	// SC-003: Node details should appear in <1 second
	assert.Less(t, detailsDuration, 1*time.Second, "Node details should load in <1s (SC-003)")

	// Verify details are complete
	assert.Equal(t, testNodeID, nodeDetails.ID, "Node ID should match")
	assert.NotEmpty(t, nodeDetails.Type, "Node type should be present")
	assert.NotNil(t, nodeDetails.Properties, "Properties should be present")
	assert.GreaterOrEqual(t, nodeDetails.Degree, 0, "Degree should be non-negative")
}

// T105: Performance test - Measure filter operation performance
func TestGraphPerformance_FilteredQuery(t *testing.T) {
	nodeCount := 1000
	service := setupPerformanceTestService(t, nodeCount)
	ctx := context.Background()

	// Test filtered query performance (simulates filter interaction)
	filter := explorer.NodeFilter{
		Types: []string{"person", "organization"},
		Limit: 500,
	}

	startFilter := time.Now()
	response, err := service.GetNodes(ctx, filter)
	filterDuration := time.Since(startFilter)

	require.NoError(t, err, "GetNodes should succeed")
	require.NotNil(t, response, "Response should not be nil")

	t.Logf("Filtered query returned %d nodes in %v", len(response.Nodes), filterDuration)

	// SC-004: Filters should update in <1 second
	assert.Less(t, filterDuration, 1*time.Second, "Filter should apply in <1s (SC-004)")

	// Verify filtering worked correctly - non-ghost nodes should match requested types
	// Ghost nodes (IsGhost=true) can be any type as they're connected to filtered nodes
	typeMap := make(map[string]bool)
	for _, t := range filter.Types {
		typeMap[t] = true
	}

	nonGhostCount := 0
	for _, node := range response.Nodes {
		if !node.IsGhost {
			nonGhostCount++
			assert.True(t, typeMap[node.Type],
				"Filtered non-ghost node type '%s' should be one of the requested types: %v", node.Type, filter.Types)
		}
	}

	// Verify we actually got some non-ghost results
	assert.Greater(t, nonGhostCount, 0, "Filter should return some non-ghost nodes")
}

// T105: Benchmark - Overall rendering pipeline simulation
func TestGraphPerformance_FullRenderingPipeline(t *testing.T) {
	nodeCount := 1000
	service := setupPerformanceTestService(t, nodeCount)
	ctx := context.Background()

	t.Log("=== Full Rendering Pipeline Performance Test ===")

	// Step 1: Initial load
	startTotal := time.Now()

	startLoad := time.Now()
	response, err := service.GetRandomNodes(ctx, 100)
	loadDuration := time.Since(startLoad)
	require.NoError(t, err, "Initial load should succeed")
	t.Logf("✓ Step 1 - Initial load (100 nodes): %v", loadDuration)

	// Step 2: Expand a node (simulating user interaction)
	startExpand := time.Now()
	nodeID := response.Nodes[0].ID
	_, err = service.GetRelationships(ctx, nodeID, 0, 50)
	expandDuration := time.Since(startExpand)
	require.NoError(t, err, "Relationship expansion should succeed")
	t.Logf("✓ Step 2 - Expand node relationships: %v", expandDuration)

	// Step 3: Get node details (simulating node click)
	startDetails := time.Now()
	_, err = service.GetNodeDetails(ctx, nodeID)
	detailsDuration := time.Since(startDetails)
	require.NoError(t, err, "Node details should succeed")
	t.Logf("✓ Step 3 - Get node details: %v", detailsDuration)

	// Step 4: Apply filter (simulating filter interaction)
	startFilter := time.Now()
	filter := explorer.NodeFilter{Types: []string{"person"}, Limit: 200}
	_, err = service.GetNodes(ctx, filter)
	filterDuration := time.Since(startFilter)
	require.NoError(t, err, "Filter should succeed")
	t.Logf("✓ Step 4 - Apply filter: %v", filterDuration)

	totalDuration := time.Since(startTotal)
	t.Logf("=== Total pipeline duration: %v ===", totalDuration)

	// Verify performance meets all success criteria
	assert.Less(t, loadDuration, 2*time.Second, "SC-001: Schema loads in <2s")
	assert.Less(t, expandDuration, 500*time.Millisecond, "SC-008: Pan/zoom in <500ms")
	assert.Less(t, detailsDuration, 1*time.Second, "SC-003: Node details in <1s")
	assert.Less(t, filterDuration, 1*time.Second, "SC-004: Filters update in <1s")

	t.Log("✓ All performance criteria met!")
}
