package explorer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/explorer"
	integration "github.com/Blogem/enron-graph/tests/integration"
)

func setupGraphTestService(t *testing.T) *explorer.GraphService {
	client, db := integration.SetupTestDBWithSQL(t)
	seedGraphTestData(t, client)
	return explorer.NewGraphService(client, db, nil)
}

func setupGraphTestServiceWithManyNodes(t *testing.T, nodeCount int) *explorer.GraphService {
	client, db := integration.SetupTestDBWithSQL(t)
	seedManyNodes(t, client, nodeCount)
	return explorer.NewGraphService(client, db, nil)
}

func setupGraphTestServiceWithHighDegreeNode(t *testing.T, relationshipCount int) (*explorer.GraphService, string) {
	client, db := integration.SetupTestDBWithSQL(t)
	nodeID := seedHighDegreeNode(t, client, relationshipCount)
	return explorer.NewGraphService(client, db, nil), nodeID
}

func seedGraphTestData(t *testing.T, client *ent.Client) {
	ctx := context.Background()

	person1, err := client.DiscoveredEntity.Create().
		SetUniqueID("person-alice").
		SetTypeCategory("person").
		SetName("Alice Smith").
		SetConfidenceScore(0.95).
		SetProperties(map[string]interface{}{
			"email": "alice@example.com",
			"role":  "Manager",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person-alice: %v", err)
	}

	person2, err := client.DiscoveredEntity.Create().
		SetUniqueID("person-bob").
		SetTypeCategory("person").
		SetName("Bob Johnson").
		SetConfidenceScore(0.90).
		SetProperties(map[string]interface{}{
			"email": "bob@example.com",
			"role":  "Developer",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person-bob: %v", err)
	}

	org1, err := client.DiscoveredEntity.Create().
		SetUniqueID("org-techcorp").
		SetTypeCategory("organization").
		SetName("TechCorp Inc").
		SetConfidenceScore(0.88).
		SetProperties(map[string]interface{}{
			"industry": "Technology",
			"founded":  "2010",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create org-techcorp: %v", err)
	}

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person1.ID).
		SetToType("discovered_entity").
		SetToID(org1.ID).
		SetType("works_at").
		SetConfidenceScore(0.92).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create relationship person1->org1: %v", err)
	}

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person2.ID).
		SetToType("discovered_entity").
		SetToID(org1.ID).
		SetType("works_at").
		SetConfidenceScore(0.91).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create relationship person2->org1: %v", err)
	}

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person1.ID).
		SetToType("discovered_entity").
		SetToID(person2.ID).
		SetType("colleague_of").
		SetConfidenceScore(0.85).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create relationship person1->person2: %v", err)
	}
}

func seedManyNodes(t *testing.T, client *ent.Client, count int) {
	ctx := context.Background()

	var previousID int
	for i := 0; i < count; i++ {
		entity, err := client.DiscoveredEntity.Create().
			SetUniqueID(fmt.Sprintf("node-%d", i)).
			SetTypeCategory("person").
			SetName(fmt.Sprintf("Person %d", i)).
			SetConfidenceScore(0.80 + float64(i%20)/100.0).
			SetProperties(map[string]interface{}{
				"index": i,
				"group": i % 10,
			}).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create node-%d: %v", i, err)
		}

		if i > 0 {
			_, err = client.Relationship.Create().
				SetFromType("discovered_entity").
				SetFromID(previousID).
				SetToType("discovered_entity").
				SetToID(entity.ID).
				SetType("connected_to").
				SetConfidenceScore(0.75).
				Save(ctx)
			if err != nil {
				t.Fatalf("Failed to create relationship %d->%d: %v", previousID, entity.ID, err)
			}
		}
		previousID = entity.ID
	}
}

func seedHighDegreeNode(t *testing.T, client *ent.Client, relationshipCount int) string {
	ctx := context.Background()

	centralNode, err := client.DiscoveredEntity.Create().
		SetUniqueID("central-hub").
		SetTypeCategory("organization").
		SetName("Central Hub").
		SetConfidenceScore(0.95).
		SetProperties(map[string]interface{}{
			"description": "High-degree central node",
			"type":        "hub",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create central node: %v", err)
	}

	for i := 0; i < relationshipCount; i++ {
		connectedNode, err := client.DiscoveredEntity.Create().
			SetUniqueID(fmt.Sprintf("connected-%d", i)).
			SetTypeCategory("person").
			SetName(fmt.Sprintf("Connected Person %d", i)).
			SetConfidenceScore(0.80).
			SetProperties(map[string]interface{}{
				"index": i,
			}).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create connected node %d: %v", i, err)
		}

		_, err = client.Relationship.Create().
			SetFromType("discovered_entity").
			SetFromID(centralNode.ID).
			SetToType("discovered_entity").
			SetToID(connectedNode.ID).
			SetType("manages").
			SetConfidenceScore(0.85).
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create relationship to connected-%d: %v", i, err)
		}
	}

	return "central-hub"
}

func TestGraphExplorer_AutoLoadNodesOnStartup(t *testing.T) {
	service := setupGraphTestServiceWithManyNodes(t, 150)
	ctx := context.Background()

	response, err := service.GetRandomNodes(ctx, 100)
	if err != nil {
		t.Fatalf("GetRandomNodes failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected graph response, got nil")
	}

	if len(response.Nodes) == 0 {
		t.Error("Expected at least some nodes, got 0")
	}
	if len(response.Nodes) > 100 {
		t.Errorf("Expected at most 100 nodes, got %d", len(response.Nodes))
	}

	if response.TotalNodes != 150 {
		t.Errorf("Expected TotalNodes=150, got %d", response.TotalNodes)
	}

	if !response.HasMore {
		t.Error("Expected HasMore=true since we have 150 total nodes and only loaded 100")
	}

	nodeIDs := make(map[string]bool)
	for _, node := range response.Nodes {
		if node.ID == "" {
			t.Error("Node missing ID")
		}
		if node.Type == "" {
			t.Error("Node missing Type")
		}
		if node.Category != "discovered" && node.Category != "promoted" {
			t.Errorf("Invalid category: %s (expected 'discovered' or 'promoted')", node.Category)
		}
		if node.Properties == nil {
			t.Error("Node missing Properties map")
		}
		nodeIDs[node.ID] = true
	}

	for _, edge := range response.Edges {
		if edge.Source == "" {
			t.Error("Edge missing Source")
		}
		if edge.Target == "" {
			t.Error("Edge missing Target")
		}
		if edge.Type == "" {
			t.Error("Edge missing Type")
		}

		if !nodeIDs[edge.Source] {
			t.Errorf("Edge source %s not in returned nodes", edge.Source)
		}
		if !nodeIDs[edge.Target] {
			t.Errorf("Edge target %s not in returned nodes", edge.Target)
		}
	}

	t.Logf("Auto-loaded %d nodes with %d edges from total of %d nodes",
		len(response.Nodes), len(response.Edges), response.TotalNodes)
}

func TestGraphExplorer_NodeClickShowsDetails(t *testing.T) {
	service := setupGraphTestService(t)
	ctx := context.Background()

	nodeID := "person-alice"
	nodeDetails, err := service.GetNodeDetails(ctx, nodeID)
	if err != nil {
		t.Fatalf("GetNodeDetails failed: %v", err)
	}

	if nodeDetails == nil {
		t.Fatal("Expected node details, got nil")
	}

	if nodeDetails.ID != nodeID {
		t.Errorf("Expected ID=%s, got %s", nodeID, nodeDetails.ID)
	}

	if nodeDetails.Type == "" {
		t.Error("Node Type is empty")
	}

	if nodeDetails.Category != "discovered" && nodeDetails.Category != "promoted" {
		t.Errorf("Invalid category: %s", nodeDetails.Category)
	}

	if nodeDetails.Properties == nil {
		t.Fatal("Node Properties is nil")
	}

	if email, ok := nodeDetails.Properties["email"].(string); !ok || email != "alice@example.com" {
		t.Errorf("Expected email=alice@example.com, got %v", nodeDetails.Properties["email"])
	}

	if role, ok := nodeDetails.Properties["role"].(string); !ok || role != "Manager" {
		t.Errorf("Expected role=Manager, got %v", nodeDetails.Properties["role"])
	}

	t.Logf("Retrieved node details: ID=%s, Type=%s, Category=%s, Properties=%+v",
		nodeDetails.ID, nodeDetails.Type, nodeDetails.Category, nodeDetails.Properties)
}

func TestGraphExplorer_ExpandNodeLoadsRelationships(t *testing.T) {
	service := setupGraphTestService(t)
	ctx := context.Background()

	nodeID := "person-alice"
	response, err := service.GetRelationships(ctx, nodeID, 0, 50)
	if err != nil {
		t.Fatalf("GetRelationships failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected relationships response, got nil")
	}

	if len(response.Edges) == 0 {
		t.Error("Expected at least some edges, got 0")
	}

	if len(response.Nodes) == 0 {
		t.Error("Expected at least some connected nodes, got 0")
	}

	if response.TotalCount == 0 {
		t.Error("Expected TotalCount > 0")
	}

	if response.Offset != 0 {
		t.Errorf("Expected Offset=0, got %d", response.Offset)
	}

	if response.HasMore {
		t.Error("Expected HasMore=false for node with only 2 relationships")
	}

	foundOutgoing := false
	foundIncoming := false
	for _, edge := range response.Edges {
		if edge.Source == nodeID {
			foundOutgoing = true
		}
		if edge.Target == nodeID {
			foundIncoming = true
		}
	}

	if !foundOutgoing && !foundIncoming {
		t.Error("Expected at least one edge to reference the expanded node")
	}

	nodeIDs := make(map[string]bool)
	for _, node := range response.Nodes {
		nodeIDs[node.ID] = true
	}

	for _, edge := range response.Edges {
		if edge.Source == nodeID {
			if !nodeIDs[edge.Target] {
				t.Errorf("Edge target %s not in returned nodes", edge.Target)
			}
		} else if edge.Target == nodeID {
			if !nodeIDs[edge.Source] {
				t.Errorf("Edge source %s not in returned nodes", edge.Source)
			}
		}
	}

	t.Logf("Expanded node %s: loaded %d edges, %d nodes, total=%d, hasMore=%v",
		nodeID, len(response.Edges), len(response.Nodes), response.TotalCount, response.HasMore)
}

func TestGraphExplorer_BatchedLoadingForHighDegreeNodes(t *testing.T) {
	service, nodeID := setupGraphTestServiceWithHighDegreeNode(t, 120)
	ctx := context.Background()

	batch1, err := service.GetRelationships(ctx, nodeID, 0, 50)
	if err != nil {
		t.Fatalf("GetRelationships (batch 1) failed: %v", err)
	}

	if batch1 == nil {
		t.Fatal("Expected relationships response for batch 1, got nil")
	}

	if len(batch1.Edges) != 50 {
		t.Errorf("Expected exactly 50 edges in batch 1, got %d", len(batch1.Edges))
	}

	if batch1.TotalCount != 120 {
		t.Errorf("Expected TotalCount=120, got %d", batch1.TotalCount)
	}

	if batch1.Offset != 0 {
		t.Errorf("Expected Offset=0 for batch 1, got %d", batch1.Offset)
	}

	if !batch1.HasMore {
		t.Error("Expected HasMore=true for batch 1 (120 total, loaded 50)")
	}

	batch2, err := service.GetRelationships(ctx, nodeID, 50, 50)
	if err != nil {
		t.Fatalf("GetRelationships (batch 2) failed: %v", err)
	}

	if len(batch2.Edges) != 50 {
		t.Errorf("Expected exactly 50 edges in batch 2, got %d", len(batch2.Edges))
	}

	if batch2.TotalCount != 120 {
		t.Errorf("Expected TotalCount=120 for batch 2, got %d", batch2.TotalCount)
	}

	if batch2.Offset != 50 {
		t.Errorf("Expected Offset=50 for batch 2, got %d", batch2.Offset)
	}

	if !batch2.HasMore {
		t.Error("Expected HasMore=true for batch 2 (120 total, loaded 100)")
	}

	batch3, err := service.GetRelationships(ctx, nodeID, 100, 50)
	if err != nil {
		t.Fatalf("GetRelationships (batch 3) failed: %v", err)
	}

	if len(batch3.Edges) != 20 {
		t.Errorf("Expected exactly 20 edges in batch 3 (remaining), got %d", len(batch3.Edges))
	}

	if batch3.TotalCount != 120 {
		t.Errorf("Expected TotalCount=120 for batch 3, got %d", batch3.TotalCount)
	}

	if batch3.Offset != 100 {
		t.Errorf("Expected Offset=100 for batch 3, got %d", batch3.Offset)
	}

	if batch3.HasMore {
		t.Error("Expected HasMore=false for batch 3 (final batch)")
	}

	edgeIDs := make(map[string]bool)
	duplicates := 0

	for _, edge := range batch1.Edges {
		edgeKey := fmt.Sprintf("%s->%s:%s", edge.Source, edge.Target, edge.Type)
		if edgeIDs[edgeKey] {
			duplicates++
		}
		edgeIDs[edgeKey] = true
	}

	for _, edge := range batch2.Edges {
		edgeKey := fmt.Sprintf("%s->%s:%s", edge.Source, edge.Target, edge.Type)
		if edgeIDs[edgeKey] {
			duplicates++
		}
		edgeIDs[edgeKey] = true
	}

	for _, edge := range batch3.Edges {
		edgeKey := fmt.Sprintf("%s->%s:%s", edge.Source, edge.Target, edge.Type)
		if edgeIDs[edgeKey] {
			duplicates++
		}
		edgeIDs[edgeKey] = true
	}

	if duplicates > 0 {
		t.Errorf("Found %d duplicate edges across batches", duplicates)
	}

	totalLoaded := len(batch1.Edges) + len(batch2.Edges) + len(batch3.Edges)
	if totalLoaded != 120 {
		t.Errorf("Expected to load 120 total edges, got %d", totalLoaded)
	}

	t.Logf("Batched loading verified: batch1=%d, batch2=%d, batch3=%d, total=%d, duplicates=%d",
		len(batch1.Edges), len(batch2.Edges), len(batch3.Edges), totalLoaded, duplicates)
}
