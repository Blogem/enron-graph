## MODIFIED Requirements

### Requirement: Registry provides EntityCreator functions for promoted types
The system SHALL maintain a global registry of EntityCreator functions that map type names to entity creation logic.

#### Scenario: EntityCreator is registered for promoted type
- **WHEN** a type schema is promoted and code generation runs
- **THEN** system registers an EntityCreator function for that type in the global PromotedTypes map

#### Scenario: EntityCreator creates entity in promoted table
- **WHEN** EntityCreator is called with metadata
- **THEN** function creates entity in the promoted table and returns created entity

#### Scenario: EntityCreator maps fields from metadata
- **WHEN** EntityCreator receives metadata with standard fields (unique_id, name, description, embedding, metadata_json)
- **THEN** function maps all applicable fields to promoted table columns

### Requirement: Registry auto-registers during initialization
The system SHALL automatically register all promoted types without manual configuration.

#### Scenario: Generated code registers types in init()
- **WHEN** application starts and ent/registry.go is loaded
- **THEN** all promoted types are registered via init() function execution

#### Scenario: Registration happens before first entity creation
- **WHEN** extractor attempts to create an entity
- **THEN** registry is already populated with all promoted types

### Requirement: EntityCreator signature is consistent across types
The system SHALL enforce a standardized function signature for all EntityCreator functions.

#### Scenario: Creator accepts client and metadata
- **WHEN** EntityCreator is called
- **THEN** function accepts (client *ent.Client, metadata EntityMetadata) parameters

#### Scenario: Creator returns entity and error
- **WHEN** EntityCreator completes
- **THEN** function returns (*ent.DiscoveredEntity, error) or promoted type equivalent

**ADDED**: EntityFinder signature mirrors EntityCreator

#### Scenario: Finder accepts client and lookup key
- **WHEN** EntityFinder is called
- **THEN** function accepts (client *ent.Client, uniqueID string) parameters

#### Scenario: Finder returns entity and error
- **WHEN** EntityFinder completes
- **THEN** function returns (*ent.DiscoveredEntity, error) or promoted type equivalent

### Requirement: Registry provides EntityFinder functions for promoted types
**ADDED**: The system SHALL maintain EntityFinder functions in parallel to EntityCreator functions.

#### Scenario: EntityFinder is registered for promoted type
- **WHEN** a type schema is promoted and code generation runs
- **THEN** system registers an EntityFinder function for that type in the global PromotedFinders map

#### Scenario: EntityFinder queries entity by unique_id
- **WHEN** EntityFinder is called with unique_id
- **THEN** function queries the promoted table by unique_id and returns matching entity

#### Scenario: EntityFinder returns not found error
- **WHEN** EntityFinder is called with non-existent unique_id
- **THEN** function returns sql.ErrNoRows or appropriate not found error

#### Scenario: Repository uses EntityFinder for type-aware lookups
- **WHEN** repository method receives type hint and needs to find entity
- **THEN** repository retrieves EntityFinder from registry and calls it

### Requirement: Generated code includes both creator and finder functions
**ADDED**: The system SHALL generate both create and find functions for each promoted type.

#### Scenario: Generated createXxx function
- **WHEN** code generation runs for promoted type
- **THEN** ent/registry.go contains createXxx function that creates entity

#### Scenario: Generated findXxx function
- **WHEN** code generation runs for promoted type
- **THEN** ent/registry.go contains findXxx function that queries by unique_id

#### Scenario: Both functions registered in init()
- **WHEN** application starts
- **THEN** init() calls registry.Register() for both creator and finder

### Requirement: Extractor uses registry for entity operations
The system SHALL use the registry for both creating and finding entities of promoted types.

#### Scenario: Extractor creates via registry
- **WHEN** extractor needs to create entity of promoted type
- **THEN** extractor retrieves EntityCreator from registry and calls it

**ADDED**: Extractor finds via registry

#### Scenario: Extractor finds via registry
- **WHEN** extractor needs to find entity of promoted type
- **THEN** extractor retrieves EntityFinder from registry and calls it

### Requirement: Field types match between creator and finder
**ADDED**: The system SHALL ensure EntityCreator and EntityFinder work with consistent data types.

#### Scenario: Creator accepts metadata struct
- **WHEN** EntityCreator is called
- **THEN** function accepts EntityMetadata with all standard fields

#### Scenario: Finder returns compatible entity
- **WHEN** EntityFinder returns entity
- **THEN** entity contains same fields that EntityCreator would populate

### Requirement: Promoter triggers registration regeneration
The system SHALL regenerate registry code when new types are promoted.

#### Scenario: Promoter runs code generation
- **WHEN** promoter promotes a new type schema
- **THEN** promoter executes entgen to regenerate registry code

**ADDED**: Registration includes finder functions

#### Scenario: Registration regeneration includes finders
- **WHEN** code regeneration completes
- **THEN** ent/registry.go contains both creator and finder for newly promoted type
