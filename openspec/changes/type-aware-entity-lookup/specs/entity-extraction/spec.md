## MODIFIED Requirements

### Requirement: Extractor passes type hints to repository methods
**ADDED**: The system SHALL provide entity type context when calling repository lookup methods.

#### Scenario: Pass type hint in createOrUpdateEntity
- **WHEN** extractor calls FindEntityByUniqueID to check for existing entity
- **THEN** extractor passes entity.TypeCategory as type hint parameter

#### Scenario: Pass type hint in deduplication
- **WHEN** dedup logic calls FindEntityByUniqueID or SimilaritySearch
- **THEN** dedup passes known entity type as hint

#### Scenario: Type hint from extraction context
- **WHEN** extractor has entity metadata with TypeCategory field
- **THEN** extractor uses TypeCategory value as type hint for repository calls

### Requirement: Extractor uses registry for type-aware operations
**ADDED**: The system SHALL leverage registry EntityFinder functions when entity type is known.

#### Scenario: Direct registry lookup for promoted types
- **WHEN** extractor needs to find entity of promoted type
- **THEN** extractor checks registry for EntityFinder and uses it if available

#### Scenario: Fallback to repository for non-promoted types
- **WHEN** extractor needs to find entity of non-promoted type
- **THEN** extractor calls repository methods without registry lookup

#### Scenario: Type determination from entity metadata
- **WHEN** extractor processes entity metadata
- **THEN** extractor determines whether type is promoted before choosing lookup strategy

### Requirement: Deduplication leverages type hints
**ADDED**: The system SHALL improve deduplication accuracy by providing type context.

#### Scenario: Similarity search with type filter
- **WHEN** dedup performs similarity search for potential duplicates
- **THEN** dedup passes entity type to focus search on relevant table

#### Scenario: Exact match lookup with type hint
- **WHEN** dedup checks for exact unique_id match
- **THEN** dedup passes type hint for optimized lookup

### Requirement: Extractor handles promoted type entity creation
**ADDED**: The system SHALL create entities in promoted tables when type is promoted.

#### Scenario: Check registry before creation
- **WHEN** extractor needs to create new entity
- **THEN** extractor checks registry for EntityCreator for that type

#### Scenario: Use registry creator for promoted types
- **WHEN** entity type has registered EntityCreator
- **THEN** extractor calls EntityCreator instead of creating in discovered_entities

#### Scenario: Fallback to discovered_entities
- **WHEN** entity type has no registered EntityCreator
- **THEN** extractor creates entity in discovered_entities table

### Requirement: Type hint availability throughout extraction pipeline
**ADDED**: The system SHALL maintain entity type information through entire extraction flow.

#### Scenario: Type flows from LLM response
- **WHEN** LLM returns entity with type field
- **THEN** extractor preserves type in EntityMetadata.TypeCategory

#### Scenario: Type available during dedup
- **WHEN** dedup processes entity
- **THEN** dedup has access to TypeCategory for type-aware queries

#### Scenario: Type available during creation
- **WHEN** extractor creates entity
- **THEN** extractor has TypeCategory for registry lookup
