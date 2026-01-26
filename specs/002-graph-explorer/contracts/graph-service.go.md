# Graph Service Contract (Go)

**Package**: `internal/explorer`  
**File**: `graph_service.go`  
**Purpose**: Backend service interface for graph data access in the Graph Explorer

---

## Interface Definition

```go
package explorer

import (
    "context"
)

// GraphService provides graph data access for the explorer frontend
type GraphService interface {
    // GetRandomNodes returns a random sample of nodes for initial graph display
    // Implements FR-003a: auto-load 50-100 nodes on startup
    // Parameters:
    //   - ctx: request context
    //   - limit: number of nodes to return (50-100 recommended)
    // Returns:
    //   - GraphResponse with random node sample and connecting edges
    //   - error if database query fails
    GetRandomNodes(ctx context.Context, limit int) (*GraphResponse, error)
    
    // GetNodes returns nodes filtered by type with optional search
    // Implements FR-007 (filtering) and FR-008 (search)
    // Parameters:
    //   - ctx: request context
    //   - filter: NodeFilter specifying types and search criteria
    // Returns:
    //   - GraphResponse with filtered nodes and edges
    //   - error if database query fails
    GetNodes(ctx context.Context, filter NodeFilter) (*GraphResponse, error)
    
    // GetRelationships returns paginated relationships for a specific node
    // Implements FR-006a: batched loading for high-degree nodes
    // Parameters:
    //   - ctx: request context
    //   - nodeID: ID of the node to expand
    //   - offset: pagination offset (0 for first batch)
    //   - limit: batch size (50 recommended)
    // Returns:
    //   - RelationshipsResponse with edges, connected nodes, and pagination info
    //   - error if database query fails or node not found
    GetRelationships(ctx context.Context, nodeID string, offset int, limit int) (*RelationshipsResponse, error)
    
    // GetNodeDetails returns complete information for a single node
    // Implements FR-011: detail panel for selected entities
    // Parameters:
    //   - ctx: request context
    //   - nodeID: ID of the node to retrieve
    // Returns:
    //   - GraphNode with all properties and metadata
    //   - error if node not found
    GetNodeDetails(ctx context.Context, nodeID string) (*GraphNode, error)
}

// NodeFilter specifies criteria for filtering graph nodes
type NodeFilter struct {
    // Types to include (empty means all types)
    // e.g., []string{"Email", "Person"}
    Types []string `json:"types"`
    
    // Category filter: "promoted", "discovered", or empty for all
    Category string `json:"category"`
    
    // SearchQuery for property value matching
    // Searches across all property values (case-insensitive)
    SearchQuery string `json:"search_query"`
    
    // Limit on number of nodes to return
    Limit int `json:"limit"`
}

// GraphResponse contains a subset of the graph for visualization
type GraphResponse struct {
    Nodes      []GraphNode `json:"nodes"`
    Edges      []GraphEdge `json:"edges"`
    TotalNodes int         `json:"total_nodes"`
    HasMore    bool        `json:"has_more"`
}

// GraphNode represents an entity in the graph
type GraphNode struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`       // e.g., "Email", "Person"
    Category   string                 `json:"category"`   // "promoted" | "discovered"
    Properties map[string]interface{} `json:"properties"` // All entity attributes
    IsGhost    bool                   `json:"is_ghost"`   // True if placeholder
    Degree     int                    `json:"degree"`     // Relationship count (optional)
}

// GraphEdge represents a relationship between two nodes
type GraphEdge struct {
    Source     string                 `json:"source"` // Source node ID
    Target     string                 `json:"target"` // Target node ID
    Type       string                 `json:"type"`   // Relationship type
    Properties map[string]interface{} `json:"properties"`
}

// RelationshipsResponse contains paginated relationships for a node
type RelationshipsResponse struct {
    Edges      []GraphEdge `json:"edges"`
    Nodes      []GraphNode `json:"nodes"`       // Connected nodes being added
    TotalCount int         `json:"total_count"` // Total relationships for this node
    HasMore    bool        `json:"has_more"`
    Offset     int         `json:"offset"` // Current pagination offset
}
```

---

## Contract Tests

**Test File**: `tests/contract/graph_service_contract_test.go`

### Test Cases

```go
package contract_test

import (
    "context"
    "testing"
    
    "github.com/Blogem/enron-graph/internal/explorer"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestGraphService_GetRandomNodes_ReturnsLimitedNodes verifies random node sampling
func TestGraphService_GetRandomNodes_ReturnsLimitedNodes(t *testing.T) {
    // Given: A GraphService with database containing at least 100 entities
    svc := setupGraphService(t)
    ctx := context.Background()
    
    // When: GetRandomNodes(100)
    resp, err := svc.GetRandomNodes(ctx, 100)
    
    // Then: Returns up to 100 nodes without error
    require.NoError(t, err)
    assert.LessOrEqual(t, len(resp.Nodes), 100)
    assert.NotEmpty(t, resp.Nodes, "should return at least some nodes if DB has data")
    
    // And: Each node has required fields
    for _, node := range resp.Nodes {
        assert.NotEmpty(t, node.ID, "node ID required")
        assert.NotEmpty(t, node.Type, "node type required")
        assert.Contains(t, []string{"promoted", "discovered"}, node.Category, "category must be promoted or discovered")
        assert.NotNil(t, node.Properties, "properties map required (can be empty)")
    }
    
    // And: Edges reference existing nodes
    nodeIDs := make(map[string]bool)
    for _, node := range resp.Nodes {
        nodeIDs[node.ID] = true
    }
    for _, edge := range resp.Edges {
        assert.True(t, nodeIDs[edge.Source], "edge source must reference a returned node")
        assert.True(t, nodeIDs[edge.Target], "edge target must reference a returned node")
        assert.NotEmpty(t, edge.Type, "edge type required")
    }
}

// TestGraphService_GetRelationships_PaginatesCorrectly verifies batch loading
func TestGraphService_GetRelationships_PaginatesCorrectly(t *testing.T) {
    // Given: A node with >50 relationships
    svc, nodeID := setupNodeWithManyRelationships(t, 120) // 120 relationships
    ctx := context.Background()
    
    // When: GetRelationships with offset=0, limit=50
    resp1, err := svc.GetRelationships(ctx, nodeID, 0, 50)
    
    // Then: Returns first 50 relationships
    require.NoError(t, err)
    assert.Len(t, resp1.Edges, 50, "should return exactly 50 edges")
    assert.Equal(t, 120, resp1.TotalCount, "should report total count")
    assert.True(t, resp1.HasMore, "should indicate more relationships exist")
    assert.Equal(t, 0, resp1.Offset, "should echo offset")
    
    // When: GetRelationships with offset=50, limit=50 (second batch)
    resp2, err := svc.GetRelationships(ctx, nodeID, 50, 50)
    
    // Then: Returns next 50 relationships
    require.NoError(t, err)
    assert.Len(t, resp2.Edges, 50, "should return next 50 edges")
    assert.Equal(t, 120, resp2.TotalCount, "total count unchanged")
    assert.True(t, resp2.HasMore, "should indicate more relationships exist")
    assert.Equal(t, 50, resp2.Offset)
    
    // When: GetRelationships with offset=100, limit=50 (final batch)
    resp3, err := svc.GetRelationships(ctx, nodeID, 100, 50)
    
    // Then: Returns remaining 20 relationships
    require.NoError(t, err)
    assert.Len(t, resp3.Edges, 20, "should return remaining 20 edges")
    assert.Equal(t, 120, resp3.TotalCount)
    assert.False(t, resp3.HasMore, "should indicate no more relationships")
    assert.Equal(t, 100, resp3.Offset)
}

// TestGraphService_GetNodes_FiltersBy Type verifies type filtering
func TestGraphService_GetNodes_FiltersByType(t *testing.T) {
    // Given: Database with Email and DiscoveredEntity types
    svc := setupGraphServiceWithMixedTypes(t)
    ctx := context.Background()
    
    // When: Filter by types=["Email"]
    filter := explorer.NodeFilter{
        Types: []string{"Email"},
        Limit: 100,
    }
    resp, err := svc.GetNodes(ctx, filter)
    
    // Then: Returns only Email nodes
    require.NoError(t, err)
    for _, node := range resp.Nodes {
        assert.Equal(t, "Email", node.Type, "should only return Email types")
        assert.Equal(t, "promoted", node.Category, "Email is promoted type")
    }
}

// TestGraphService_GetNodes_SearchesProperties verifies search functionality
func TestGraphService_GetNodes_SearchesProperties(t *testing.T) {
    // Given: Database with nodes containing "john@enron.com"
    svc := setupGraphServiceWithTestData(t)
    ctx := context.Background()
    
    // When: Search for "john@enron.com"
    filter := explorer.NodeFilter{
        SearchQuery: "john@enron.com",
        Limit:       100,
    }
    resp, err := svc.GetNodes(ctx, filter)
    
    // Then: Returns nodes with matching property values
    require.NoError(t, err)
    assert.NotEmpty(t, resp.Nodes, "should find matching nodes")
    
    // And: At least one node contains the search term in properties
    foundMatch := false
    for _, node := range resp.Nodes {
        for _, value := range node.Properties {
            if str, ok := value.(string); ok && contains(str, "john@enron.com") {
                foundMatch = true
                break
            }
        }
    }
    assert.True(t, foundMatch, "at least one node should match search query")
}

// TestGraphService_GetNodeDetails_ReturnsCompleteNode verifies detail retrieval
func TestGraphService_GetNodeDetails_ReturnsCompleteNode(t *testing.T) {
    // Given: A specific node ID
    svc, nodeID := setupGraphServiceWithNode(t)
    ctx := context.Background()
    
    // When: GetNodeDetails(nodeID)
    node, err := svc.GetNodeDetails(ctx, nodeID)
    
    // Then: Returns complete node data
    require.NoError(t, err)
    assert.Equal(t, nodeID, node.ID)
    assert.NotEmpty(t, node.Type)
    assert.NotEmpty(t, node.Category)
    assert.NotNil(t, node.Properties)
    assert.False(t, node.IsGhost, "real node should not be ghost")
}

// TestGraphService_GetNodeDetails_ErrorsOnMissingNode verifies error handling
func TestGraphService_GetNodeDetails_ErrorsOnMissingNode(t *testing.T) {
    // Given: A GraphService
    svc := setupGraphService(t)
    ctx := context.Background()
    
    // When: GetNodeDetails with non-existent ID
    _, err := svc.GetNodeDetails(ctx, "non-existent-id-12345")
    
    // Then: Returns error
    assert.Error(t, err, "should error on non-existent node")
}
```

---

## Implementation Notes

### Performance Expectations

Based on Success Criteria:
- `GetRandomNodes(100)` MUST complete in <2 seconds (SC-001)
- `GetRelationships()` MUST complete in <2 seconds even for high-degree nodes (SC-006)
- `GetNodes()` with filters MUST complete in <1 second (SC-004)

### Error Handling

All methods SHOULD return errors for:
- Database connection failures
- Invalid node IDs (GetNodeDetails, GetRelationships)
- Invalid pagination parameters (offset < 0, limit < 1)

### Thread Safety

All methods MUST be safe for concurrent use (multiple frontend requests in parallel).

### Wails Binding

The implementation struct will be bound to Wails frontend using:

```go
// In cmd/explorer/main.go
func main() {
    app := NewApp() // App struct contains GraphService
    
    err := wails.Run(&options.App{
        Title:  "Enron Graph Explorer",
        Width:  1400,
        Height: 1000,
        Bind: []interface{}{
            app.GraphService, // Exposes all methods to JavaScript
        },
    })
}
```

Frontend can then call:

```typescript
import { GetRandomNodes } from '../wailsjs/go/explorer/GraphService';

const response = await GetRandomNodes(100);
// response is typed as GraphResponse
```
