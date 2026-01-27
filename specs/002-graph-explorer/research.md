# Phase 0: Research - Graph Explorer

**Feature**: 002-graph-explorer  
**Date**: 2026-01-26  
**Status**: Complete

## Research Questions

Based on the Technical Context section of plan.md, the following areas required research to resolve uncertainties and validate technology choices.

---

## 1. Wails Framework for Go + Web Hybrid Applications

**Question**: Can Wails provide the bridge between Go backend (ent + PostgreSQL) and React frontend while producing a single-binary desktop application?

**Decision**: **Yes - Use Wails v2**

**Rationale**:
- **Single Binary**: Wails bundles Go backend + webview + frontend assets into one executable per platform (macOS/Linux/Windows)
- **Go-Frontend Bridge**: Provides automatic binding of Go functions to JavaScript without REST/gRPC overhead
- **Existing Stack Compatibility**: Works seamlessly with existing ent ORM and PostgreSQL connections
- **Development Experience**: Hot reload for frontend, standard Go tooling for backend
- **Production Ready**: Used in production applications; active community; version 2.x is stable

**Alternatives Considered**:
- **Electron + Go subprocess**: Rejected due to larger bundle size, complex IPC, and multiple processes
- **Web server + browser**: Rejected because spec requires desktop application, not web app
- **Pure bubbletea TUI**: Cannot achieve force-directed graph visualization or performance requirements

**Implementation Notes**:
- Install wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Use `wails init` to generate project structure, then adapt to existing monorepo
- Frontend build integrated via `wails build` command

---

## 2. Force-Directed Graph Visualization Libraries

**Question**: Which React library can deliver force-directed layout with directional arrows, WebGL performance for 1000+ nodes, and batched loading?

**Decision**: **Use react-force-graph (3D/2D version)**

**Rationale**:
- **Physics Simulation**: Built on d3-force, providing attraction/repulsion out-of-the-box (FR-003b requirement)
- **Performance**: Uses Three.js for WebGL rendering, proven to handle 10k+ nodes smoothly (exceeds SC-008 requirement of 1000 nodes <500ms)
- **Directional Arrows**: Native support via `linkDirectionalArrowLength` and `linkDirectionalArrowRelPos` (FR-003c requirement)
- **Node Differentiation**: Supports custom node colors, shapes, and sizes via callbacks (FR-005)
- **Interactive**: Built-in pan, zoom, hover, click handlers (FR-003, FR-004)
- **Active Maintenance**: 3k+ GitHub stars, regular updates, comprehensive documentation

**Alternatives Considered**:
- **vis.js / vis-network**: Older library, Canvas-based (not WebGL), struggles with 500+ nodes
- **cytoscape.js**: Excellent for biological networks but force-directed layouts are slower than react-force-graph; primarily Canvas-based
- **sigma.js**: Fast WebGL but requires more manual physics setup; react-force-graph provides better React integration
- **D3.js raw force layout**: Would require months of custom work to match react-force-graph features

**Implementation Notes**:
- Install: `npm install react-force-graph`
- Use 2D version initially (simpler); can upgrade to 3D if desired
- Lazy load node data on expand to handle high-degree nodes (FR-006a)

---

## 3. Schema Introspection for Ent-Generated Types

**Question**: How can we programmatically extract promoted schema type information (property names, types, sample values) from ent-generated code?

**Decision**: **Use ent's Schema Descriptor API + SQL reflection for sample values**

**Rationale**:
- **Schema Descriptor**: Ent provides `client.Schema` which exposes schema metadata programmatically
- **Table Metadata**: Can iterate `Schema.Tables` to get table names, columns, types, and relationships
- **Runtime Reflection**: Ent's generated code includes methods like `Table()`, `Columns()`, `Fields()` on each type
- **Sample Values**: Query database directly with `SELECT * FROM table_name LIMIT 1` to get real sample data
- **Type Safety**: All metadata extraction happens in strongly-typed Go code

**Alternatives Considered**:
- **Parse ent/schema/*.go source files**: Rejected because it's brittle, requires Go AST parsing, and duplicates work ent already does
- **PostgreSQL information_schema**: Doesn't provide ent-specific annotations or relationship metadata
- **Manual schema documentation**: Violates spec requirement to automatically expose current schema (FR-001)

**Implementation Notes**:
```go
// Pseudocode for schema introspection
func GetPromotedSchemaTypes(client *ent.Client) []SchemaType {
    var types []SchemaType
    for _, table := range client.Schema.Tables {
        properties := make([]Property, len(table.Columns))
        for i, col := range table.Columns {
            properties[i] = Property{
                Name: col.Name,
                Type: col.Type.String(),
                SampleValue: queryOne(table.Name, col.Name), // SQL query
            }
        }
        types = append(types, SchemaType{
            Name: table.Name,
            Category: "promoted",
            Properties: properties,
            Count: count(table.Name), // SELECT COUNT(*) FROM table
        })
    }
    return types
}
```

---

## 4. Discovered Entity Representation

**Question**: How do we expose discovered entities (from DiscoveredEntity table) alongside promoted types in a unified schema view?

**Decision**: **Query DiscoveredEntity table, group by type field, and merge with promoted types**

**Rationale**:
- **Existing Schema**: `DiscoveredEntity` already has fields for `type`, `properties` (JSONB), `confidence`, `discovered_at`
- **Grouping**: SQL `GROUP BY type` with `COUNT(*)` gives us entity type counts (FR-009)
- **Property Extraction**: Since properties are JSONB, we can extract unique keys across all entities of a type using PostgreSQL's `jsonb_object_keys`
- **Sample Values**: Query one entity per type to get representative property values
- **Unified Model**: Both promoted and discovered types map to the same `SchemaType` Go struct with a `category` field distinguishing them

**Alternatives Considered**:
- **Separate endpoints for promoted vs discovered**: Rejected because spec requires unified view (US1 acceptance scenario 2)
- **Cache schema metadata in memory**: Deferred to later optimization; initial implementation queries on-demand

**Implementation Notes**:
```sql
-- Get discovered entity type counts
SELECT 
    type, 
    COUNT(*) as count,
    array_agg(DISTINCT jsonb_object_keys(properties)) as property_keys
FROM discovered_entities 
GROUP BY type;

-- Get sample entity per type
SELECT DISTINCT ON (type) type, properties, confidence 
FROM discovered_entities 
ORDER BY type, discovered_at DESC;
```

---

## 5. Batched Relationship Loading Strategy

**Question**: How do we implement "Load 50 more" functionality for high-degree nodes efficiently in the graph traversal?

**Decision**: **Use OFFSET/LIMIT pagination with relationship count caching**

**Rationale**:
- **Existing Repository**: `internal/graph/repository.go` already has graph traversal methods; extend with pagination parameters
- **Efficient Pagination**: ent supports `.Offset(n).Limit(50)` for batched queries
- **Count Caching**: Store relationship counts in memory after first expansion to show "Load 50 more (120 remaining)"
- **Frontend State**: React component maintains `expandedNodes` map tracking which batches have been loaded per node

**Alternatives Considered**:
- **Cursor-based pagination**: More complex for undirected graph edges; OFFSET/LIMIT is simpler for this use case
- **Load all relationships, filter client-side**: Violates performance requirement; could overload browser with 1000+ edges
- **Virtual scrolling in edge list**: Not applicable; we're rendering visual graph, not a list

**Implementation Notes**:
```go
type GetRelationshipsRequest struct {
    NodeID string
    Offset int
    Limit  int
}

type GetRelationshipsResponse struct {
    Edges []GraphEdge
    TotalCount int
    HasMore bool
}

// Usage: LoadMore button calls: getRelationships(nodeID, currentOffset + 50, 50)
```

---

## 6. Frontend Build Integration with Wails

**Question**: How do we integrate the React frontend build process into the existing Go project workflow?

**Decision**: **Use Wails build toolchain with Vite**

**Rationale**:
- **Wails Integration**: `wails dev` runs both Go backend and Vite dev server with hot reload
- **Production Builds**: `wails build` compiles Go, builds React bundle, and embeds assets automatically
- **No Manual Steps**: Developers only run `wails dev`; CI/CD only runs `wails build`
- **Standard React Tools**: Can use Vite, npm, TypeScript with zero custom configuration

**Alternatives Considered**:
- **Separate npm build + Go embed**: Rejected because Wails handles this automatically
- **Serve frontend from Go http server**: Rejected because Wails provides better dev experience

**Implementation Notes**:
- Create `frontend/` directory with `package.json` and `vite.config.ts`
- Configure `wails.json` to point to frontend directory
- Add `frontend/dist` to `.gitignore`
- Document in quickstart.md: "Run `wails dev` to start development"

---

## 7. Testing Strategy for Hybrid Application

**Question**: How do we test the Wails bridge between Go backend and React frontend?

**Decision**: **Three-layer testing approach**

**Rationale**:
- **Contract Tests (Go)**: Test Go service functions in isolation (graph_service, schema_service) - standard Go testing
- **Component Tests (React)**: Test React components with mocked Wails bindings - React Testing Library
- **Integration Tests (Wails)**: Use Wails test mode to verify end-to-end flows - `wails dev -test`

**Testing Layers**:

1. **Contract Tests** (`tests/contract/graph_service_contract_test.go`):
   - Test that `GraphService.GetRandomNodes(100)` returns valid GraphNode slice
   - Test that `SchemaService.GetPromotedTypes()` returns correct ent schema metadata
   - Test batched loading: `GetRelationships(nodeID, offset, limit)` pagination
   - **Framework**: Go standard testing + testify

2. **Component Tests** (`frontend/src/components/*.test.tsx`):
   - Test GraphCanvas renders force-directed layout with mock data
   - Test SchemaPanel displays promoted vs discovered types correctly
   - Test LoadMoreButton shows correct remaining count
   - **Framework**: React Testing Library + Vitest

3. **Integration Tests** (`tests/integration/explorer/graph_explorer_test.go`):
   - Launch Wails app in headless mode
   - Verify schema loads within 2 seconds (SC-001)
   - Verify 100 nodes render on startup (FR-003a)
   - Verify clicking node shows detail panel (US2 acceptance scenario 2)
   - **Framework**: Wails testing utilities + chromedp for UI automation

**Alternatives Considered**:
- **E2E tests with Playwright**: Rejected because Wails provides native testing support
- **Manual testing only**: Violates TDD principle (Constitution III)

---

## 8. Data Type Mapping Between Go and TypeScript

**Question**: How do we ensure type safety across the Wails bridge when passing GraphNode, GraphEdge, SchemaType between Go and React?

**Decision**: **Use Wails TypeScript generation + manual type definitions**

**Rationale**:
- **Automatic Generation**: Wails generates TypeScript types from Go structs when you run `wails generate`
- **Manual Refinement**: Can create `frontend/src/types/graph.ts` to extend or alias generated types for better React ergonomics
- **Runtime Validation**: For development, add zod schemas to validate Wails responses match expected types

**Implementation Notes**:
```go
// internal/explorer/models.go
type GraphNode struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`
    Category   string                 `json:"category"` // "promoted" | "discovered"
    Properties map[string]interface{} `json:"properties"`
    IsGhost    bool                   `json:"is_ghost"` // For unloaded nodes in batching
}

type GraphEdge struct {
    Source string `json:"source"`
    Target string `json:"target"`
    Type   string `json:"type"`
}
```

After `wails generate`, TypeScript file is created:
```typescript
// frontend/wailsjs/go/models.ts (auto-generated)
export interface GraphNode {
    id: string;
    type: string;
    category: string;
    properties: {[key: string]: any};
    is_ghost: boolean;
}
```

---

## Summary of Technology Stack

| Component | Technology | Justification |
|-----------|-----------|---------------|
| **Desktop Framework** | Wails v2 | Single binary, Go-JS bridge, production-ready |
| **Frontend Framework** | React + TypeScript | Industry standard, rich ecosystem, Wails support |
| **Graph Visualization** | react-force-graph | WebGL performance, force-directed physics, 10k+ nodes |
| **Build Tool** | Vite | Fast, modern, Wails default |
| **Backend Services** | Go with ent ORM | Existing stack, reuse graph repository |
| **Schema Introspection** | ent Schema API + SQL | Type-safe, leverages existing ent metadata |
| **Testing** | Go testing + React Testing Library + Wails test mode | Three-layer strategy matching architecture |
| **Type Safety** | Wails TypeScript generation | Auto-generated types from Go structs |

---

## Remaining Unknowns

**None** - All research questions from Technical Context have been resolved. Ready to proceed to Phase 1 (Data Model & Contracts).
