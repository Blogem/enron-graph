## ADDED Requirements

### Requirement: Repository queries across promoted and discovered tables
The system SHALL enable repository methods to query entities from both promoted type tables and the discovered_entities table.

#### Scenario: Find entity in promoted table with type hint
- **WHEN** FindEntityByUniqueID is called with a unique_id and type hint for a promoted type
- **THEN** system queries the promoted table directly and returns the entity

#### Scenario: Find entity in discovered_entities with type hint
- **WHEN** FindEntityByUniqueID is called with a unique_id and type hint for a non-promoted type
- **THEN** system queries the discovered_entities table and returns the entity

#### Scenario: Find entity without type hint
- **WHEN** FindEntityByUniqueID is called with only a unique_id (no type hint)
- **THEN** system searches discovered_entities first, then uses fallback strategies to locate the entity

### Requirement: Type hint parameter is backward compatible
The system SHALL accept type hints as optional variadic parameters that do not break existing code.

#### Scenario: Legacy call without type hint
- **WHEN** existing code calls FindEntityByUniqueID with only uniqueID parameter
- **THEN** system accepts the call and uses fallback lookup strategies

#### Scenario: New call with type hint
- **WHEN** code calls FindEntityByUniqueID with uniqueID and type hint parameters
- **THEN** system uses the type hint for optimized lookup

### Requirement: Three-tier fallback strategy for lookups without type hints
The system SHALL implement a prioritized fallback chain when no type hint is provided.

#### Scenario: Entity found in discovered_entities (tier 1)
- **WHEN** FindEntityByUniqueID is called without type hint and entity exists in discovered_entities
- **THEN** system returns entity after single query to discovered_entities table

#### Scenario: Type inferred from relationships table (tier 2)
- **WHEN** entity not found in discovered_entities and relationships table has references to the entity
- **THEN** system queries relationships to infer entity type and queries the appropriate table

#### Scenario: Parallel search across promoted tables (tier 3)
- **WHEN** entity not found in discovered_entities and not referenced in relationships
- **THEN** system searches all promoted tables in parallel and returns first match

#### Scenario: Entity not found anywhere
- **WHEN** entity does not exist in any table
- **THEN** system returns not found error

### Requirement: FindEntityByID supports type hints
The system SHALL accept optional type hint for FindEntityByID to enable direct table lookup by numeric ID.

#### Scenario: Find by ID with type hint
- **WHEN** FindEntityByID is called with numeric ID and type hint
- **THEN** system queries the specified table directly

#### Scenario: Find by ID without type hint uses relationships inference
- **WHEN** FindEntityByID is called without type hint
- **THEN** system queries relationships table to determine entity type before fetching

### Requirement: FindEntitiesByType queries promoted tables when applicable
The system SHALL query promoted tables when FindEntitiesByType is called with a promoted type name.

#### Scenario: Query promoted type
- **WHEN** FindEntitiesByType is called with a type that has been promoted (e.g., "person")
- **THEN** system queries the promoted table (persons) and returns entities

#### Scenario: Query non-promoted type
- **WHEN** FindEntitiesByType is called with a non-promoted type
- **THEN** system queries discovered_entities table with type filter

#### Scenario: Query all types (empty type parameter)
- **WHEN** FindEntitiesByType is called with empty type parameter
- **THEN** system queries all tables (discovered_entities and all promoted tables) and returns combined results

### Requirement: SimilaritySearch queries all tables with embeddings
The system SHALL perform vector similarity search across all tables that contain embedding columns.

#### Scenario: UNION search across tables
- **WHEN** SimilaritySearch is called
- **THEN** system generates UNION query across discovered_entities and all promoted tables with embedding columns

#### Scenario: Skip tables without embeddings
- **WHEN** SimilaritySearch identifies tables without embedding columns
- **THEN** system excludes those tables from the UNION query

#### Scenario: Results ordered by similarity
- **WHEN** UNION query returns results from multiple tables
- **THEN** system orders combined results by similarity score and returns top K matches

### Requirement: TraverseRelationships passes type hints from relationship metadata
The system SHALL use entity type information from relationships table when traversing the graph.

#### Scenario: Traverse with known target type
- **WHEN** TraverseRelationships encounters a relationship with to_type specified
- **THEN** system passes to_type as hint when fetching target entity

#### Scenario: Traverse accepts any entity type
- **WHEN** relationship references a promoted type (not just "discovered_entity")
- **THEN** system successfully fetches and includes the entity in traversal results

### Requirement: Registry provides EntityFinder functions
The system SHALL maintain EntityFinder functions in the registry for all promoted types, mirroring EntityCreator functions.

#### Scenario: Finder registered for promoted type
- **WHEN** a type is promoted and code generation runs
- **THEN** registry contains both EntityCreator and EntityFinder for that type

#### Scenario: Finder queries by unique_id
- **WHEN** EntityFinder is called with unique_id
- **THEN** finder queries the promoted table and returns matching entity

#### Scenario: Finder returns not found for non-existent entity
- **WHEN** EntityFinder is called with unique_id that doesn't exist
- **THEN** finder returns appropriate not found error

### Requirement: Performance optimization via type hints
The system SHALL achieve O(1) lookup performance when type hints are provided.

#### Scenario: Direct table access with hint
- **WHEN** repository method receives accurate type hint
- **THEN** system executes single query to correct table without fallback logic

#### Scenario: Fallback remains efficient
- **WHEN** type hint is not provided
- **THEN** system executes at most 3 queries (discovered, relationships inference, parallel search)

### Requirement: Parallel search limits concurrency
The system SHALL limit concurrent queries during parallel promoted table search to prevent connection pool exhaustion.

#### Scenario: Worker pool for parallel search
- **WHEN** parallel search is triggered across N promoted tables
- **THEN** system uses bounded concurrency (e.g., max 5 concurrent queries)

#### Scenario: Early exit on first match
- **WHEN** parallel search finds entity in any table
- **THEN** system cancels remaining queries and returns immediately

### Requirement: GetDistinctEntityTypes returns both discovered and promoted types
The system SHALL enumerate all entity types across both discovered_entities and promoted tables using the relationships table.

#### Scenario: Query discovered types from discovered_entities
- **WHEN** GetDistinctEntityTypes is called
- **THEN** system queries distinct type_category values from discovered_entities table

#### Scenario: Query promoted types from relationships table
- **WHEN** GetDistinctEntityTypes is called
- **THEN** system queries distinct from_type and to_type values from relationships table

#### Scenario: Combine and deduplicate types
- **WHEN** both queries complete
- **THEN** system combines discovered and relationship types, removes duplicates, and returns unique list

#### Scenario: Exclude generic type marker
- **WHEN** processing relationship types
- **THEN** system filters out "discovered_entity" generic marker to avoid duplication

### Requirement: FindRelationshipsByEntity supports promoted entity types
The system SHALL find relationships for entities in both discovered_entities and promoted tables.

#### Scenario: Find relationships by promoted entity type
- **WHEN** FindRelationshipsByEntity is called with promoted entity type (e.g., "person") and ID
- **THEN** system queries relationships where from_type or to_type matches the promoted type

#### Scenario: Find relationships by discovered entity type
- **WHEN** FindRelationshipsByEntity is called with "discovered_entity" type and ID
- **THEN** system queries relationships using discovered_entity type filter

#### Scenario: Return all matching relationships
- **WHEN** relationships table has references using both old and new type names
- **THEN** system returns all relationships matching the specified entity type and ID

### Requirement: FindShortestPath handles promoted entities in path
The system SHALL find shortest paths between entities regardless of whether they are promoted or discovered.

#### Scenario: Path traversal uses type-aware entity lookup
- **WHEN** FindShortestPath traverses relationships to build path
- **THEN** system uses type-aware FindEntityByID to verify entity existence

#### Scenario: Path includes promoted entities
- **WHEN** shortest path includes entities from promoted tables
- **THEN** system successfully resolves all entities and returns complete path

#### Scenario: BFS respects entity types in relationships
- **WHEN** breadth-first search encounters relationships with promoted types
- **THEN** system correctly follows relationships across promoted and discovered entities
