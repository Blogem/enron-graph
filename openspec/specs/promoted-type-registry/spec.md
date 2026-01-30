## ADDED Requirements

### Requirement: Registry maintains promoted type mappings
The system SHALL maintain a global registry that maps entity type names to their corresponding Ent creation functions.

#### Scenario: Registry initialized at startup
- **WHEN** the application starts
- **THEN** all promoted schemas are automatically registered via init() functions

#### Scenario: Lookup existing promoted type
- **WHEN** code queries the registry for a promoted type name (e.g., "Person")
- **THEN** registry returns the corresponding EntityCreator function

#### Scenario: Lookup non-promoted type
- **WHEN** code queries the registry for a type that has not been promoted
- **THEN** registry indicates the type does not exist in the registry

### Requirement: Auto-registration during code generation
The system SHALL automatically generate registration code for all promoted schemas during Ent code generation.

#### Scenario: Code generation with promoted schemas
- **WHEN** `go generate ./ent` runs with promoted schemas in `ent/schema/`
- **THEN** generated code includes registration calls for each promoted schema in an init() function

#### Scenario: Registration code maps type names to creators
- **WHEN** registration code is generated for a schema (e.g., "Person")
- **THEN** generated code registers the schema name with a function that creates entities of that type

#### Scenario: Generated code handles field mapping
- **WHEN** registration code is generated for a schema with fields
- **THEN** generated creator function maps data properties to Ent builder setter methods

### Requirement: EntityCreator function signature
The system SHALL define EntityCreator as a function that accepts a context and property map and returns an entity.

#### Scenario: EntityCreator accepts context and data
- **WHEN** an EntityCreator function is called
- **THEN** it receives a context.Context and map[string]any containing entity properties

#### Scenario: EntityCreator returns created entity
- **WHEN** an EntityCreator function successfully creates an entity
- **THEN** it returns the created entity and nil error

#### Scenario: EntityCreator returns error on failure
- **WHEN** an EntityCreator function fails to create an entity
- **THEN** it returns nil entity and an error describing the failure

### Requirement: Extractor checks registry before generic storage
The system SHALL check the promoted types registry before falling back to generic DiscoveredEntity storage.

#### Scenario: Entity type exists in registry
- **WHEN** extractor creates an entity with a type that exists in the registry
- **THEN** extractor uses the registered EntityCreator to store the entity in the promoted table

#### Scenario: Entity type not in registry
- **WHEN** extractor creates an entity with a type that does not exist in the registry
- **THEN** extractor falls back to storing the entity in the DiscoveredEntity table

#### Scenario: Registry creation fails
- **WHEN** extractor calls a registered EntityCreator and it returns an error
- **THEN** extractor logs the error and falls back to storing in DiscoveredEntity table

### Requirement: Template supports scalar field types
The system SHALL generate registration code that handles scalar field types from entity property maps.

#### Scenario: String field mapping
- **WHEN** generated code processes a string field from the property map
- **THEN** it type-asserts the value and calls the corresponding setter method

#### Scenario: Integer field mapping
- **WHEN** generated code processes an integer field from the property map
- **THEN** it type-asserts the value and calls the corresponding setter method

#### Scenario: Float field mapping
- **WHEN** generated code processes a float field from the property map
- **THEN** it type-asserts the value and calls the corresponding setter method

#### Scenario: Boolean field mapping
- **WHEN** generated code processes a boolean field from the property map
- **THEN** it type-asserts the value and calls the corresponding setter method

#### Scenario: Missing optional field
- **WHEN** a field is missing from the property map
- **THEN** generated code skips that field without error

#### Scenario: Nil field value
- **WHEN** a field has a nil value in the property map
- **THEN** generated code skips that field without error

### Requirement: Promoter triggers registry regeneration
The system SHALL regenerate the entity registry when a type is promoted.

#### Scenario: Schema promotion triggers codegen
- **WHEN** promoter successfully creates a new schema in `ent/schema/`
- **THEN** promoter runs `go generate ./ent` to regenerate Ent code including registry

#### Scenario: Registry available after promotion
- **WHEN** a type is promoted and code generation completes
- **THEN** subsequent application restarts include the newly promoted type in the registry
