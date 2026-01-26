# Feature Specification: Graph Explorer

**Feature Branch**: `002-graph-explorer`  
**Created**: January 26, 2026  
**Status**: Draft  
**Input**: User description: "i want to add a graph explorer to the encron graph. It needs to: expose the current schema, both the promoted/concrete types/tables and the discovered entities. - browse the graph contents in a graphical interface"

## Clarifications

### Session 2026-01-26

- Q: How should users specify which graph data to load when they first open the explorer? → A: Auto-load a small sample (e.g., 50-100 random nodes) immediately on startup
- Q: Which graph layout strategy should be used to position nodes? → A: Force-directed layout - nodes positioned using physics simulation (connected nodes attract, all nodes repel)
- Q: How should relationship directionality be displayed in the graph? → A: Arrows on edges - draw directional arrows pointing from source to target node
- Q: What information should be displayed for each property in the schema view? → A: Names, types, and sample values - include example data from actual entities
- Q: How should the system handle expanding nodes that have many relationships (e.g., >50)? → A: Auto-batch with load-more - initially show first 50 relationships with "Load 50 more" button showing remaining count

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Schema Overview (Priority: P1)

As a data analyst or developer, I need to understand what types of entities exist in the Enron graph system, including both the formally defined schema (promoted/concrete types and tables) and the dynamically discovered entities, so I can understand the data structure before querying or analyzing it.

**Why this priority**: Understanding the schema is foundational - users cannot effectively explore graph data without knowing what entity types exist. This delivers immediate value by providing visibility into the system's data model.

**Independent Test**: Can be fully tested by opening the explorer, viewing the schema panel, and verifying that both promoted types (from the formal schema) and discovered entity types are displayed with their properties and counts.

**Acceptance Scenarios**:

1. **Given** the graph contains promoted entity types (Email, Relationship, etc.), **When** I open the schema view, **Then** I see a list of all promoted types with their property names, data types, and sample values from actual entities
2. **Given** the graph contains discovered entities, **When** I view the schema, **Then** I see discovered entity types listed separately with their occurrence counts
3. **Given** I select a specific entity type from the schema view, **When** I click on it, **Then** I see detailed information about that type including its properties and relationships
4. **Given** the schema has been updated with new discovered entities, **When** I refresh the schema view, **Then** I see the latest entity types without restarting the application

---

### User Story 2 - Browse Graph Visually (Priority: P1)

As a user investigating the Enron dataset, I need to visualize the graph structure in an interactive graphical interface, so I can explore relationships between entities, understand connection patterns, and navigate the data intuitively.

**Why this priority**: Visual graph browsing is the core value proposition of a graph explorer. This enables users to discover insights that would be difficult to find through queries alone. Equal priority to schema view as both are essential MVP features.

**Independent Test**: Can be fully tested by loading a subset of graph data, displaying it as nodes and edges, and verifying that users can pan, zoom, click nodes to see details, and expand/collapse connections.

**Acceptance Scenarios**:

1. **Given** I have opened the graph explorer, **When** the application starts, **Then** I see a small sample of 50-100 random nodes automatically loaded and displayed using force-directed layout
2. **Given** nodes are displayed in the graph view, **When** I click on a node, **Then** I see its properties and can identify what entity type it represents
3. **Given** I see a node with connected relationships, **When** I click to expand it, **Then** I see its connected nodes rendered in the graph
4. **Given** I expand a node with more than 50 relationships, **When** the initial batch loads, **Then** I see the first 50 relationships and a "Load 50 more" button showing the remaining count
5. **Given** the graph visualization is displayed, **When** I use pan and zoom controls, **Then** I can navigate large graphs smoothly
5. **Given** I have multiple entity types in the view, **When** I look at the graph, **Then** different entity types are visually distinguishable (color, shape, or icon)

---

### User Story 3 - Filter and Search Graph Data (Priority: P2)

As a user analyzing the graph, I need to filter the displayed entities by type, properties, or relationship patterns, so I can focus on relevant subsets of data and reduce visual clutter.

**Why this priority**: Once users can view the schema and browse the graph, filtering becomes important for working with larger datasets. However, users can still derive value from exploring small subsets without filtering initially.

**Independent Test**: Can be fully tested by applying filters to hide/show specific entity types, searching for entities by property values, and verifying the graph updates to show only matching entities.

**Acceptance Scenarios**:

1. **Given** the graph contains multiple entity types, **When** I filter to show only Email entities, **Then** only Email nodes and their relationships are displayed
2. **Given** I want to find a specific entity, **When** I search by a property value (e.g., email address), **Then** matching entities are highlighted or isolated in the view
3. **Given** I have applied filters, **When** I clear them, **Then** the full graph view is restored
4. **Given** the graph has both promoted and discovered entities, **When** I filter by entity source (promoted vs discovered), **Then** only entities from that source are shown

---

### User Story 4 - Navigate Entity Details (Priority: P2)

As a user examining specific entities, I need to view detailed information about individual nodes including all their properties and metadata, so I can understand entity attributes without leaving the visual interface.

**Why this priority**: Detail viewing enhances the exploration experience but is secondary to basic visualization. Users can initially explore by hovering or clicking for basic info before needing full detail panels.

**Independent Test**: Can be fully tested by selecting any node in the graph and verifying that a detail panel shows all properties, timestamps, entity type, and any associated metadata.

**Acceptance Scenarios**:

1. **Given** I have selected a node in the graph, **When** I view the detail panel, **Then** I see all property values for that entity
2. **Given** I am viewing entity details, **When** the entity has relationships, **Then** I see a list of connected entities with relationship types
3. **Given** I am viewing a discovered entity, **When** I check its details, **Then** I see metadata about when it was discovered and its confidence score or source

---

### Edge Cases

- What happens when the graph contains thousands of nodes and rendering all of them would overwhelm the interface?
  - System should limit initial render to a manageable subset (e.g., 100-500 nodes) and provide controls to load more
- How does the system handle entities with missing or null property values in the detail view?
  - Display properties clearly indicating which values are null/missing vs empty strings
- What happens when attempting to expand a node that has hundreds of relationships?
  - System initially loads first 50 relationships and displays a "Load 50 more" button showing remaining count; subsequent clicks load additional batches
- How does the system handle schema changes while the explorer is open?
  - Provide a refresh mechanism; optionally detect schema changes and notify the user
- What happens when there are no discovered entities yet?
  - Display an appropriate message in the schema view indicating no discovered entities exist

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a schema view showing all promoted/concrete entity types from the database schema with property names, data types, and sample values from actual entities
- **FR-002**: System MUST display discovered entity types separately from promoted types in the schema view
- **FR-003**: System MUST provide a graphical visualization of graph nodes and edges with interactive pan and zoom capabilities
- **FR-003a**: System MUST auto-load a sample of 50-100 random nodes on startup to provide immediate visual feedback
- **FR-003b**: System MUST use a force-directed layout algorithm to position nodes (connected nodes attract, all nodes repel)
- **FR-003c**: System MUST display relationship edges with directional arrows pointing from source to target node
- **FR-004**: System MUST allow users to click on nodes to view their properties and connected relationships
- **FR-005**: System MUST visually differentiate between different entity types (through color, shape, size, or icons)
- **FR-006**: System MUST support expanding nodes to load and display their connected entities
- **FR-006a**: System MUST batch relationship loading for high-degree nodes (>50 relationships) by initially showing 50 relationships with a "Load more" control displaying remaining count
- **FR-007**: System MUST provide filtering capabilities to show/hide entities by type
- **FR-008**: System MUST allow searching for entities by property values
- **FR-009**: System MUST display entity counts for each type in the schema view
- **FR-010**: System MUST handle large graphs by limiting initial render size and providing load-more functionality
- **FR-011**: System MUST provide a detail panel showing all properties and metadata for selected entities
- **FR-012**: System MUST support both promoted schema entities (Email, Relationship, SchemaPromotion) and DiscoveredEntity types
- **FR-013**: System MUST refresh schema and graph data without requiring application restart

### Key Entities

- **Promoted Schema Entities**: Email, Relationship, SchemaPromotion, and any other concrete types defined in the ent schema - these are formally defined entities with fixed schemas
- **Discovered Entities**: Dynamically identified entity types from the DiscoveredEntity table that represent patterns found in the data but not yet promoted to formal schema
- **Graph Nodes**: Visual representation of entities in the interface, containing entity type, properties, and visual attributes
- **Graph Edges**: Visual representation of relationships between entities, showing relationship type and directionality via arrows pointing from source to target
- **Schema Metadata**: Information about entity types including property names, data types, sample values, counts, and whether they are promoted or discovered

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can view the complete schema (both promoted and discovered types) within 2 seconds of opening the explorer
- **SC-002**: Users can successfully navigate and explore a graph subset of 100-500 nodes with smooth pan/zoom interactions (no lag or stuttering)
- **SC-003**: Users can identify and click on any node to view its details within 1 second
- **SC-004**: Users can successfully filter the graph to show only specific entity types and see results update within 1 second
- **SC-005**: The interface clearly distinguishes between promoted schema types and discovered entities (visible through separate sections or visual indicators)
- **SC-006**: Users can expand a node's connections and see related entities rendered in the graph within 2 seconds
- **SC-007**: Users can complete a typical exploration task (view schema → load graph subset → filter by type → examine node details) in under 5 minutes on first use
- **SC-008**: The graph visualization remains responsive when displaying up to 1000 nodes (pan/zoom operations complete within 500ms)
