# Phase 1: Data Model - Graph Explorer

**Feature**: 002-graph-explorer  
**Date**: 2026-01-26  
**Status**: Complete

## Overview

This document defines the entities, their properties, relationships, and validation rules for the Graph Explorer feature. The data model bridges between the existing database schema (ent-managed) and the frontend visualization layer.

---

## Entity Definitions

### 1. GraphNode (DTO)

**Purpose**: Unified representation of any entity (promoted or discovered) for visualization.

**Source**: Dynamically created from `ent` entities (Email, Relationship, SchemaPromotion) OR `DiscoveredEntity` table.

**Properties**:

| Field | Type | Description | Validation | Source |
|-------|------|-------------|------------|--------|
| `id` | string | Unique identifier (UUID or ent ID) | Required, non-empty | Database primary key |
| `type` | string | Entity type name (e.g., "Email", "Person", "Organization") | Required, non-empty | Ent table name OR `DiscoveredEntity.type` |
| `category` | string | "promoted" or "discovered" | Required, enum | Determined by service layer |
| `properties` | map[string]interface{} | All entity attributes as key-value pairs | Required (can be empty map) | Ent fields OR `DiscoveredEntity.properties` JSONB |
| `is_ghost` | bool | True if node is placeholder for unloaded relationships | Required, default false | Set during batched expansion |
| `degree` | int | Total relationship count (for high-degree node warning) | Optional, ≥0 | COUNT query on relationships |

**Relationships**: None (DTO representing database entities)

**Validation Rules**:
- `category` MUST be either "promoted" or "discovered"
- `properties` MUST be a valid JSON object
- `id` MUST be unique within a graph response
- If `is_ghost` is true, `properties` MAY be empty (node not yet loaded)

**State Transitions**: Immutable (recreated on each query)

---

### 2. GraphEdge (DTO)

**Purpose**: Represents a relationship between two entities in the graph visualization.

**Source**: Created from `Relationship` table OR inferred from ent edges (e.g., Email.to, Email.from).

**Properties**:

| Field | Type | Description | Validation | Source |
|-------|------|-------------|------------|--------|
| `source` | string | ID of source node | Required, non-empty | Relationship.from_id or ent edge source |
| `target` | string | ID of target node | Required, non-empty | Relationship.to_id or ent edge target |
| `type` | string | Relationship type (e.g., "sent_to", "works_for", "mentions") | Required, non-empty | Relationship.type or ent edge name |
| `properties` | map[string]interface{} | Edge attributes (weight, timestamp, etc.) | Optional (can be empty) | Relationship.properties JSONB or ent edge fields |

**Relationships**:
- References GraphNode via `source` (many-to-one)
- References GraphNode via `target` (many-to-one)

**Validation Rules**:
- `source` and `target` MUST reference existing GraphNode IDs in the same response
- `source` and `target` MUST NOT be equal (no self-loops in v1)
- `type` MUST be a valid relationship type (validated against schema)

**State Transitions**: Immutable (recreated on each query)

---

### 3. SchemaType (DTO)

**Purpose**: Describes an entity type's schema for the schema browser panel.

**Source**: Aggregated from `ent.Schema` metadata (promoted types) OR `DiscoveredEntity` table (discovered types).

**Properties**:

| Field | Type | Description | Validation | Source |
|-------|------|-------------|------------|--------|
| `name` | string | Type name (e.g., "Email", "Person") | Required, non-empty | Ent table name OR `DiscoveredEntity.type` |
| `category` | string | "promoted" or "discovered" | Required, enum | Determined by source |
| `count` | int | Number of entities of this type | Required, ≥0 | COUNT(*) query |
| `properties` | []PropertyDefinition | List of property metadata | Required (can be empty array) | Ent schema columns OR JSONB key aggregation |
| `relationships` | []string | Names of relationship types involving this entity | Optional | Ent edges OR Relationship.type aggregation |

**Nested Type: PropertyDefinition**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | string | Property name | Required, non-empty |
| `data_type` | string | Data type (e.g., "string", "int", "timestamp", "jsonb") | Required, non-empty |
| `sample_value` | interface{} | Example value from an actual entity | Optional | (can be nil for empty tables) |
| `nullable` | bool | Whether property can be null | Required, default true |

**Relationships**: None (DTO representing schema metadata)

**Validation Rules**:
- `category` MUST be either "promoted" or "discovered"
- `name` MUST be unique within the schema response
- `properties` array MUST have unique `name` fields
- For discovered types, `data_type` is inferred from sample values (e.g., typeof in JSON)

**State Transitions**: Immutable (recreated on each schema query)

---

### 4. GraphResponse (DTO)

**Purpose**: Top-level response containing a graph subset for visualization.

**Source**: Assembled by `GraphService` from database queries.

**Properties**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `nodes` | []GraphNode | All nodes in this graph subset | Required (can be empty array) |
| `edges` | []GraphEdge | All edges connecting nodes in this subset | Required (can be empty array) |
| `total_nodes` | int | Total nodes available (for pagination info) | Required, ≥0 |
| `has_more` | bool | Whether more nodes exist beyond this response | Required |

**Relationships**:
- Contains multiple GraphNode instances
- Contains multiple GraphEdge instances

**Validation Rules**:
- All `edges[].source` and `edges[].target` MUST reference a node in `nodes[]`
- `nodes[].id` MUST be unique within the array
- `has_more` should be true if `len(nodes) < total_nodes`

**State Transitions**: Immutable (recreated on each query)

---

### 5. RelationshipsResponse (DTO)

**Purpose**: Response for batched relationship loading ("Load 50 more" functionality).

**Source**: Assembled by `GraphService.GetRelationships()` with pagination.

**Properties**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `edges` | []GraphEdge | Relationships in this batch | Required (can be empty array) |
| `nodes` | []GraphNode | Connected nodes (targets of relationships) | Required (can be empty array) |
| `total_count` | int | Total relationships for this node | Required, ≥0 |
| `has_more` | bool | Whether more batches exist | Required |
| `offset` | int | Current offset in pagination | Required, ≥0 |

**Relationships**:
- Contains multiple GraphEdge instances
- Contains multiple GraphNode instances (nodes being added to graph)

**Validation Rules**:
- `offset + len(edges)` should equal next offset if `has_more` is true
- All `edges[].source` MUST be the same (the node being expanded)
- All `edges[].target` MUST reference a node in `nodes[]`

**State Transitions**: Immutable (recreated on each paginated query)

---

### 6. SchemaResponse (DTO)

**Purpose**: Complete schema metadata for the schema browser panel.

**Source**: Assembled by `SchemaService` from ent metadata and DiscoveredEntity queries.

**Properties**:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `promoted_types` | []SchemaType | Formally defined ent schema types | Required (can be empty array) |
| `discovered_types` | []SchemaType | Dynamically discovered entity types | Required (can be empty array) |
| `total_entities` | int | Total entity count across all types | Required, ≥0 |

**Relationships**:
- Contains multiple SchemaType instances in two categories

**Validation Rules**:
- `promoted_types[].name` MUST be unique
- `discovered_types[].name` MUST be unique
- `promoted_types[].name` and `discovered_types[].name` MUST NOT overlap
- `total_entities` MUST equal sum of all SchemaType.count values

**State Transitions**: Immutable (recreated on each schema query; may change if new entities discovered)

---

## Existing Database Entities (Reference)

### DiscoveredEntity (ent schema - existing)

**Table**: `discovered_entities`

**Key Fields**:
- `id` (UUID)
- `type` (string) - Entity type name (e.g., "Person", "Organization")
- `properties` (JSONB) - Dynamic properties extracted by LLM
- `confidence` (float) - Discovery confidence score
- `discovered_at` (timestamp)

**Usage in Graph Explorer**:
- Queried via `GROUP BY type` to generate discovered SchemaType entries
- Individual entities converted to GraphNode with `category="discovered"`
- Properties extracted from JSONB for GraphNode.properties

---

### Relationship (ent schema - existing)

**Table**: `relationships`

**Key Fields**:
- `id` (UUID)
- `from_id` (UUID) - Source entity
- `to_id` (UUID) - Target entity  
- `type` (string) - Relationship type (e.g., "sent_to", "reports_to")
- `properties` (JSONB) - Optional edge attributes
- `created_at` (timestamp)

**Usage in Graph Explorer**:
- Directly mapped to GraphEdge instances
- Batched via OFFSET/LIMIT for high-degree nodes (FR-006a)
- Filtered by type for relationship type filtering

---

### Email (ent schema - existing)

**Table**: `emails`

**Key Fields**:
- `id` (UUID)
- `message_id` (string)
- `subject` (string)
- `from_addr` (string)
- `to_addrs` (string array)
- `sent_at` (timestamp)
- `body` (text)

**Usage in Graph Explorer**:
- Converted to GraphNode with `category="promoted"`
- Properties extracted from ent fields
- May have implicit relationships via `from_addr` and `to_addrs`

---

## Data Flow

### Schema Loading (User Story 1)

```
User opens explorer
  → SchemaService.GetSchema()
    → Query ent.Schema.Tables for promoted types
    → Query DiscoveredEntity GROUP BY type for discovered types
    → For each type: COUNT(*) for entity counts
    → For each type: SELECT ... LIMIT 1 for sample values
    → Assemble SchemaResponse
  → Frontend renders SchemaPanel with promoted_types and discovered_types
```

### Initial Graph Load (User Story 2, FR-003a)

```
Application startup
  → GraphService.GetRandomNodes(50-100)
    → SELECT random samples from all entity tables (Email, DiscoveredEntity, etc.)
    → Convert to GraphNode instances
    → Query relationships connecting sampled nodes
    → Convert relationships to GraphEdge instances
    → Assemble GraphResponse
  → Frontend renders force-directed graph in GraphCanvas
```

### Node Expansion (User Story 2, FR-006)

```
User clicks "Expand" on a node
  → GraphService.GetRelationships(nodeID, offset=0, limit=50)
    → Query relationships WHERE from_id=nodeID OR to_id=nodeID
    → Count total relationships (cache for subsequent calls)
    → Apply LIMIT 50 OFFSET 0
    → Fetch connected nodes (targets/sources not already in graph)
    → Assemble RelationshipsResponse with has_more=true if count > 50
  → Frontend adds new nodes and edges to graph
  → If has_more: render "Load 50 more (X remaining)" button
```

### Filter Application (User Story 3)

```
User applies entity type filter (e.g., show only "Email")
  → GraphService.GetNodes(types=["Email"], limit=100)
    → Query only Email table
    → Convert to GraphNode instances with category="promoted"
    → Query relationships WHERE from_id IN (emails) OR to_id IN (emails)
    → Assemble GraphResponse
  → Frontend replaces current graph with filtered subset
```

---

## Validation Summary

| Entity | Key Validation Rules |
|--------|---------------------|
| GraphNode | category ∈ {promoted, discovered}, id unique, properties is valid JSON |
| GraphEdge | source ≠ target, source and target exist in nodes[], type non-empty |
| SchemaType | category ∈ {promoted, discovered}, name unique, count ≥ 0 |
| PropertyDefinition | name unique within SchemaType, data_type non-empty |
| GraphResponse | edge references validate, nodes[] has unique IDs |
| RelationshipsResponse | offset ≥ 0, offset + len(edges) ≤ total_count |
| SchemaResponse | type names unique within categories, no overlap between categories |

---

## Edge Cases Handled

1. **Empty Database**: SchemaResponse returns empty arrays; GraphResponse returns empty nodes/edges
2. **No Discovered Entities**: SchemaResponse.discovered_types is empty array
3. **High-Degree Node**: RelationshipsResponse.has_more=true, batching via offset
4. **Null Property Values**: PropertyDefinition.sample_value is null/nil, frontend displays "(null)"
5. **Ghost Nodes**: GraphNode with is_ghost=true for unloaded relationship targets; properties may be empty
6. **Self-Referential Types**: Relationship table may have from_type=to_type (e.g., Person→Person); GraphEdge allows this
7. **Disconnected Nodes**: GraphResponse may include nodes with zero edges (isolated entities)

---

## Future Enhancements (Out of Scope for v1)

- **Entity Merging**: Merge duplicate discovered entities (requires entity resolution logic)
- **Schema Versioning**: Track schema changes over time (requires migration history)
- **Computed Properties**: Derive properties from graph structure (e.g., centrality scores)
- **Write Operations**: Create, update, delete entities via explorer (v1 is read-only)
- **Real-Time Updates**: WebSocket notifications when new entities discovered (v1 is query-on-demand)
