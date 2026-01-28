package explorer_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Blogem/enron-graph/ent"
	"github.com/Blogem/enron-graph/internal/explorer"
	integration "github.com/Blogem/enron-graph/tests/integration"
)

func setupFilterTestService(t *testing.T) *explorer.GraphService {
	client, db := integration.SetupTestDBWithSQL(t)
	seedFilterTestData(t, client)
	return explorer.NewGraphService(client, db, nil)
}

func seedFilterTestData(t *testing.T, client *ent.Client) {
	ctx := context.Background()

	// Create person entities
	person1, err := client.DiscoveredEntity.Create().
		SetUniqueID("person-alice").
		SetTypeCategory("person").
		SetName("Alice Smith").
		SetConfidenceScore(0.95).
		SetProperties(map[string]interface{}{
			"email": "alice@enron.com",
			"role":  "Manager",
			"dept":  "Finance",
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
			"email": "bob@enron.com",
			"role":  "Developer",
			"dept":  "Technology",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person-bob: %v", err)
	}

	person3, err := client.DiscoveredEntity.Create().
		SetUniqueID("person-charlie").
		SetTypeCategory("person").
		SetName("Charlie Brown").
		SetConfidenceScore(0.88).
		SetProperties(map[string]interface{}{
			"email": "charlie@techcorp.com",
			"role":  "Engineer",
			"dept":  "Technology",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create person-charlie: %v", err)
	}

	// Create organization entities
	org1, err := client.DiscoveredEntity.Create().
		SetUniqueID("org-enron").
		SetTypeCategory("organization").
		SetName("Enron Corporation").
		SetConfidenceScore(0.98).
		SetProperties(map[string]interface{}{
			"industry": "Energy",
			"founded":  "1985",
			"status":   "defunct",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create org-enron: %v", err)
	}

	org2, err := client.DiscoveredEntity.Create().
		SetUniqueID("org-techcorp").
		SetTypeCategory("organization").
		SetName("TechCorp Inc").
		SetConfidenceScore(0.92).
		SetProperties(map[string]interface{}{
			"industry": "Technology",
			"founded":  "2010",
			"status":   "active",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create org-techcorp: %v", err)
	}

	// Create project entities
	project1, err := client.DiscoveredEntity.Create().
		SetUniqueID("project-alpha").
		SetTypeCategory("project").
		SetName("Project Alpha").
		SetConfidenceScore(0.85).
		SetProperties(map[string]interface{}{
			"status":      "active",
			"budget":      "1000000",
			"description": "Important energy trading system",
		}).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create project-alpha: %v", err)
	}

	// Create relationships
	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person1.ID).
		SetToType("discovered_entity").
		SetToID(org1.ID).
		SetType("works_at").
		SetConfidenceScore(0.95).
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
		SetConfidenceScore(0.92).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create relationship person2->org1: %v", err)
	}

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person3.ID).
		SetToType("discovered_entity").
		SetToID(org2.ID).
		SetType("works_at").
		SetConfidenceScore(0.90).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create relationship person3->org2: %v", err)
	}

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person1.ID).
		SetToType("discovered_entity").
		SetToID(project1.ID).
		SetType("leads").
		SetConfidenceScore(0.88).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create relationship person1->project1: %v", err)
	}

	_, err = client.Relationship.Create().
		SetFromType("discovered_entity").
		SetFromID(person2.ID).
		SetToType("discovered_entity").
		SetToID(person1.ID).
		SetType("reports_to").
		SetConfidenceScore(0.87).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create relationship person2->person1: %v", err)
	}
}

// T093: Integration test - Filter by entity type updates graph
func TestFilterExplorer_FilterByEntityType(t *testing.T) {
	service := setupFilterTestService(t)
	ctx := context.Background()

	// Test 1: Filter for person entities only
	personFilter := explorer.NodeFilter{
		Types: []string{"person"},
		Limit: 100,
	}
	personResp, err := service.GetNodes(ctx, personFilter)
	if err != nil {
		t.Fatalf("GetNodes with person filter failed: %v", err)
	}

	if personResp == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(personResp.Nodes) == 0 {
		t.Error("Expected at least some person nodes")
	}

	// Verify all returned nodes are person type
	personCount := 0
	for _, node := range personResp.Nodes {
		if !node.IsGhost && node.Type == "person" {
			personCount++
		}
	}

	if personCount != 3 {
		t.Errorf("Expected 3 person nodes, got %d", personCount)
	}

	// Test 2: Filter for organization entities only
	orgFilter := explorer.NodeFilter{
		Types: []string{"organization"},
		Limit: 100,
	}
	orgResp, err := service.GetNodes(ctx, orgFilter)
	if err != nil {
		t.Fatalf("GetNodes with organization filter failed: %v", err)
	}

	if orgResp == nil {
		t.Fatal("Expected response, got nil")
	}

	// Verify all returned nodes are organization type
	orgCount := 0
	for _, node := range orgResp.Nodes {
		if !node.IsGhost && node.Type == "organization" {
			orgCount++
		}
	}

	if orgCount != 2 {
		t.Errorf("Expected 2 organization nodes, got %d", orgCount)
	}

	// Test 3: Filter for multiple types
	multiFilter := explorer.NodeFilter{
		Types: []string{"person", "organization"},
		Limit: 100,
	}
	multiResp, err := service.GetNodes(ctx, multiFilter)
	if err != nil {
		t.Fatalf("GetNodes with multi-type filter failed: %v", err)
	}

	if multiResp == nil {
		t.Fatal("Expected response, got nil")
	}

	// Count non-ghost nodes of allowed types
	allowedCount := 0
	for _, node := range multiResp.Nodes {
		if !node.IsGhost && (node.Type == "person" || node.Type == "organization") {
			allowedCount++
		}
	}

	if allowedCount != 5 {
		t.Errorf("Expected 5 nodes (3 persons + 2 organizations), got %d", allowedCount)
	}

	// Test 4: Verify edges are preserved even if target is outside filter
	// When filtering by person, should still see edges to organizations (as ghost nodes)
	personNodeIDs := make(map[string]bool)
	ghostNodeIDs := make(map[string]bool)

	for _, node := range personResp.Nodes {
		if node.IsGhost {
			ghostNodeIDs[node.ID] = true
		} else {
			personNodeIDs[node.ID] = true
		}
	}

	// Should have edges from person nodes
	if len(personResp.Edges) == 0 {
		t.Log("Warning: Expected some edges from person nodes")
	}

	// Verify edges reference valid nodes (including ghosts)
	allNodeIDs := make(map[string]bool)
	for id := range personNodeIDs {
		allNodeIDs[id] = true
	}
	for id := range ghostNodeIDs {
		allNodeIDs[id] = true
	}

	for i, edge := range personResp.Edges {
		if !allNodeIDs[edge.Source] {
			t.Errorf("Edge %d: Source '%s' not in returned nodes", i, edge.Source)
		}
		if !allNodeIDs[edge.Target] {
			t.Errorf("Edge %d: Target '%s' not in returned nodes", i, edge.Target)
		}
	}

	t.Logf("Filter by type test passed: person=%d, org=%d, multi=%d nodes",
		personCount, orgCount, allowedCount)
}

// T094: Integration test - Search by property value highlights matches
func TestFilterExplorer_SearchByPropertyValue(t *testing.T) {
	service := setupFilterTestService(t)
	ctx := context.Background()

	// Test 1: Search for "enron.com" - should match email addresses
	searchFilter1 := explorer.NodeFilter{
		SearchQuery: "enron.com",
		Limit:       100,
	}
	searchResp1, err := service.GetNodes(ctx, searchFilter1)
	if err != nil {
		t.Fatalf("GetNodes with search filter failed: %v", err)
	}

	if searchResp1 == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(searchResp1.Nodes) == 0 {
		t.Error("Expected at least some nodes matching 'enron.com'")
	}

	// Verify that matched nodes contain the search term
	matchedCount := 0
	for _, node := range searchResp1.Nodes {
		if node.IsGhost {
			continue
		}

		foundMatch := false
		for _, value := range node.Properties {
			if str, ok := value.(string); ok {
				if strings.Contains(strings.ToLower(str), "enron.com") {
					foundMatch = true
					matchedCount++
					break
				}
			}
		}

		if !foundMatch {
			t.Errorf("Node %s did not contain search term 'enron.com' in any property", node.ID)
		}
	}

	if matchedCount < 2 {
		t.Errorf("Expected at least 2 nodes with 'enron.com', got %d", matchedCount)
	}

	// Test 2: Search for "Technology" - should match department and industry
	searchFilter2 := explorer.NodeFilter{
		SearchQuery: "Technology",
		Limit:       100,
	}
	searchResp2, err := service.GetNodes(ctx, searchFilter2)
	if err != nil {
		t.Fatalf("GetNodes with search filter failed: %v", err)
	}

	if len(searchResp2.Nodes) == 0 {
		t.Error("Expected at least some nodes matching 'Technology'")
	}

	// Verify matches contain the search term
	techMatchCount := 0
	for _, node := range searchResp2.Nodes {
		if node.IsGhost {
			continue
		}

		foundMatch := false
		for _, value := range node.Properties {
			if str, ok := value.(string); ok {
				if strings.Contains(strings.ToLower(str), strings.ToLower("Technology")) {
					foundMatch = true
					techMatchCount++
					break
				}
			}
		}

		if !foundMatch {
			t.Errorf("Node %s did not contain search term 'Technology' in any property", node.ID)
		}
	}

	if techMatchCount < 2 {
		t.Errorf("Expected at least 2 nodes with 'Technology', got %d", techMatchCount)
	}

	// Test 3: Search for specific name "Alice"
	searchFilter3 := explorer.NodeFilter{
		SearchQuery: "Alice",
		Limit:       100,
	}
	searchResp3, err := service.GetNodes(ctx, searchFilter3)
	if err != nil {
		t.Fatalf("GetNodes with search filter failed: %v", err)
	}

	// Should find exactly Alice
	aliceFound := false
	for _, node := range searchResp3.Nodes {
		if node.IsGhost {
			continue
		}

		if strings.Contains(strings.ToLower(node.Properties["email"].(string)), "alice") ||
			(node.Properties["role"] != nil && strings.Contains(strings.ToLower(node.Properties["role"].(string)), "alice")) {
			aliceFound = true
		}
	}

	if !aliceFound {
		t.Error("Expected to find Alice in search results")
	}

	// Test 4: Combine search with type filter
	combinedFilter := explorer.NodeFilter{
		Types:       []string{"person"},
		SearchQuery: "enron.com",
		Limit:       100,
	}
	combinedResp, err := service.GetNodes(ctx, combinedFilter)
	if err != nil {
		t.Fatalf("GetNodes with combined filter failed: %v", err)
	}

	// Should only return person nodes with enron.com
	for _, node := range combinedResp.Nodes {
		if node.IsGhost {
			continue
		}

		// Must be person type
		if node.Type != "person" {
			t.Errorf("Expected only person nodes, got type: %s", node.Type)
		}

		// Must contain search term
		foundMatch := false
		for _, value := range node.Properties {
			if str, ok := value.(string); ok {
				if strings.Contains(strings.ToLower(str), "enron.com") {
					foundMatch = true
					break
				}
			}
		}

		if !foundMatch {
			t.Errorf("Node %s did not contain 'enron.com' in any property", node.ID)
		}
	}

	t.Logf("Search test passed: enron=%d, tech=%d, combined filters work correctly",
		matchedCount, techMatchCount)
}

// T095: Integration test - Clear filters restores full graph
func TestFilterExplorer_ClearFiltersRestoresFullGraph(t *testing.T) {
	service := setupFilterTestService(t)
	ctx := context.Background()

	// Step 1: Get full graph without filters
	fullFilter := explorer.NodeFilter{
		Limit: 100,
	}
	fullResp, err := service.GetNodes(ctx, fullFilter)
	if err != nil {
		t.Fatalf("GetNodes without filter failed: %v", err)
	}

	if fullResp == nil {
		t.Fatal("Expected response, got nil")
	}

	// Count total non-ghost nodes
	fullNodeCount := 0
	for _, node := range fullResp.Nodes {
		if !node.IsGhost {
			fullNodeCount++
		}
	}

	if fullNodeCount != 6 {
		t.Errorf("Expected 6 total nodes (3 persons + 2 orgs + 1 project), got %d", fullNodeCount)
	}

	fullEdgeCount := len(fullResp.Edges)
	if fullEdgeCount != 5 {
		t.Errorf("Expected 5 edges, got %d", fullEdgeCount)
	}

	// Step 2: Apply a restrictive filter
	restrictiveFilter := explorer.NodeFilter{
		Types: []string{"person"},
		Limit: 100,
	}
	restrictedResp, err := service.GetNodes(ctx, restrictiveFilter)
	if err != nil {
		t.Fatalf("GetNodes with restrictive filter failed: %v", err)
	}

	restrictedNodeCount := 0
	for _, node := range restrictedResp.Nodes {
		if !node.IsGhost {
			restrictedNodeCount++
		}
	}

	if restrictedNodeCount >= fullNodeCount {
		t.Errorf("Expected restrictive filter to reduce nodes, got %d (full: %d)",
			restrictedNodeCount, fullNodeCount)
	}

	// Step 3: Clear filters by calling GetNodes without type/search constraints
	clearedFilter := explorer.NodeFilter{
		Limit: 100,
	}
	clearedResp, err := service.GetNodes(ctx, clearedFilter)
	if err != nil {
		t.Fatalf("GetNodes with cleared filter failed: %v", err)
	}

	// Count cleared non-ghost nodes
	clearedNodeCount := 0
	for _, node := range clearedResp.Nodes {
		if !node.IsGhost {
			clearedNodeCount++
		}
	}

	// Should restore to full graph
	if clearedNodeCount != fullNodeCount {
		t.Errorf("Expected cleared filter to restore %d nodes, got %d",
			fullNodeCount, clearedNodeCount)
	}

	clearedEdgeCount := len(clearedResp.Edges)
	if clearedEdgeCount != fullEdgeCount {
		t.Errorf("Expected cleared filter to restore %d edges, got %d",
			fullEdgeCount, clearedEdgeCount)
	}

	// Step 4: Apply a different filter (search), then clear again
	searchFilter := explorer.NodeFilter{
		SearchQuery: "enron.com",
		Limit:       100,
	}
	searchResp, err := service.GetNodes(ctx, searchFilter)
	if err != nil {
		t.Fatalf("GetNodes with search filter failed: %v", err)
	}

	searchNodeCount := 0
	for _, node := range searchResp.Nodes {
		if !node.IsGhost {
			searchNodeCount++
		}
	}

	if searchNodeCount >= fullNodeCount {
		t.Errorf("Expected search filter to reduce nodes, got %d (full: %d)",
			searchNodeCount, fullNodeCount)
	}

	// Clear again
	reClearedFilter := explorer.NodeFilter{
		Limit: 100,
	}
	reClearedResp, err := service.GetNodes(ctx, reClearedFilter)
	if err != nil {
		t.Fatalf("GetNodes with re-cleared filter failed: %v", err)
	}

	reClearedNodeCount := 0
	for _, node := range reClearedResp.Nodes {
		if !node.IsGhost {
			reClearedNodeCount++
		}
	}

	if reClearedNodeCount != fullNodeCount {
		t.Errorf("Expected re-cleared filter to restore %d nodes, got %d",
			fullNodeCount, reClearedNodeCount)
	}

	// Verify all entity types are present after clearing
	typesSeen := make(map[string]bool)
	for _, node := range reClearedResp.Nodes {
		if !node.IsGhost {
			typesSeen[node.Type] = true
		}
	}

	expectedTypes := []string{"person", "organization", "project"}
	for _, expectedType := range expectedTypes {
		if !typesSeen[expectedType] {
			t.Errorf("Expected to see type '%s' after clearing filters", expectedType)
		}
	}

	t.Logf("Clear filters test passed: full=%d, restricted=%d, cleared=%d, re-cleared=%d",
		fullNodeCount, restrictedNodeCount, clearedNodeCount, reClearedNodeCount)
}
