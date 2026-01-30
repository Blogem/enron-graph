## ADDED Requirements

### Requirement: Initiate entity type promotion
The system SHALL allow users to initiate promotion of a discovered entity type to a formal Ent schema.

#### Scenario: User selects type for promotion
- **WHEN** user selects a type from analysis results and clicks "Promote"
- **THEN** system SHALL display promotion preview with detected properties

#### Scenario: User initiates promotion directly by type name
- **WHEN** user provides a valid entity type name for promotion
- **THEN** system SHALL generate schema definition and display promotion preview

#### Scenario: Invalid type name provided
- **WHEN** user provides a type name that does not exist in discovered entities
- **THEN** system SHALL return error "Type not found in discovered entities"

### Requirement: Generate schema definition
The system SHALL generate a schema definition from discovered entity properties.

#### Scenario: Generate properties from discovered entities
- **WHEN** generating schema for a type
- **THEN** system SHALL analyze all entities of that type and derive property definitions

#### Scenario: Detect property types
- **WHEN** analyzing entity properties
- **THEN** system SHALL infer data types (string, int, float, bool, time) based on property values

#### Scenario: Determine required properties
- **WHEN** generating schema definition
- **THEN** properties appearing in >80% of entities SHALL be marked as required

#### Scenario: Include validation rules
- **WHEN** properties have consistent patterns
- **THEN** system SHALL generate appropriate validation rules (e.g., email format, min/max length)

### Requirement: Display promotion preview
The system SHALL show users a preview of the schema before executing promotion.

#### Scenario: Show detected properties
- **WHEN** displaying promotion preview
- **THEN** system SHALL list all detected properties with their types and required status

#### Scenario: Show property sample values
- **WHEN** viewing property details
- **THEN** system SHALL display sample values from existing entities

#### Scenario: Show affected entity count
- **WHEN** previewing promotion
- **THEN** system SHALL display the count of entities that will be migrated

### Requirement: Confirm promotion action
The system SHALL require user confirmation before executing promotion.

#### Scenario: User confirms promotion
- **WHEN** user reviews schema and clicks "Confirm Promote"
- **THEN** system SHALL execute the promotion workflow

#### Scenario: User cancels promotion
- **WHEN** user clicks "Cancel" in promotion preview
- **THEN** system SHALL abort promotion and return to analysis view

### Requirement: Execute promotion workflow
The system SHALL execute the complete promotion workflow including schema generation, file creation, and data migration.

#### Scenario: Successful promotion execution
- **WHEN** executing promotion for valid type
- **THEN** system SHALL create schema file, run migrations, and migrate entities

#### Scenario: Create Ent schema file
- **WHEN** promotion executes
- **THEN** system SHALL create schema file at ent/schema/<typename>.go with generated properties

#### Scenario: Migrate discovered entities to promoted type
- **WHEN** promotion completes schema generation
- **THEN** system SHALL migrate all entities of that type from discovered_entities to new promoted table

#### Scenario: Track migration statistics
- **WHEN** executing data migration
- **THEN** system SHALL count entities migrated and validation errors encountered

### Requirement: Handle promotion errors
The system SHALL handle errors during promotion and provide clear feedback.

#### Scenario: Schema generation fails
- **WHEN** schema generation encounters an error
- **THEN** system SHALL return error with details and NOT create any files

#### Scenario: File write permission error
- **WHEN** system cannot write to ent/schema directory
- **THEN** system SHALL return error indicating file permission issue

#### Scenario: Database migration fails
- **WHEN** data migration encounters database error
- **THEN** system SHALL rollback changes and return error with details

#### Scenario: Validation errors during migration
- **WHEN** some entities fail validation during migration
- **THEN** system SHALL continue migration but report validation error count

### Requirement: Display promotion results
The system SHALL display the results of promotion execution to the user.

#### Scenario: Successful promotion results
- **WHEN** promotion completes successfully
- **THEN** system SHALL display success message with schema file path and entity count

#### Scenario: Show entities migrated count
- **WHEN** promotion completes
- **THEN** system SHALL display the number of entities successfully migrated

#### Scenario: Show validation errors count
- **WHEN** promotion completes with validation errors
- **THEN** system SHALL display the count of entities that failed validation

#### Scenario: Show schema file location
- **WHEN** promotion creates schema file
- **THEN** system SHALL display the absolute path to the created file

#### Scenario: Promotion failure results
- **WHEN** promotion fails
- **THEN** system SHALL display error message with failure reason

### Requirement: Calculate project root for file generation
The system SHALL correctly determine the project root directory for schema file creation.

#### Scenario: Determine project root from working directory
- **WHEN** promotion needs to write schema file
- **THEN** system SHALL calculate project root relative to the explorer executable location

#### Scenario: Verify schema directory exists
- **WHEN** preparing to write schema file
- **THEN** system SHALL verify that ent/schema directory exists in project root

#### Scenario: Handle incorrect project root
- **WHEN** project root calculation is incorrect
- **THEN** system SHALL return error indicating schema directory not found

### Requirement: Integrate with promoter package
The system SHALL delegate promotion logic to the existing internal/promoter package.

#### Scenario: Call promoter for schema generation
- **WHEN** executing promotion
- **THEN** system SHALL call analyst.GenerateSchemaForType to get schema definition

#### Scenario: Call promoter for type promotion
- **WHEN** executing promotion workflow
- **THEN** system SHALL call promoter.PromoteType with schema definition and project paths

#### Scenario: Convert analyst schema to promoter format
- **WHEN** passing schema to promoter
- **THEN** system SHALL convert analyst.SchemaDefinition to promoter.SchemaDefinition format

#### Scenario: Provide database connections to promoter
- **WHEN** calling promoter
- **THEN** system SHALL provide both ent.Client and sql.DB connections for queries and migrations

### Requirement: Update schema cache after promotion
The system SHALL refresh the schema cache after successful promotion.

#### Scenario: Refresh schema after promotion
- **WHEN** promotion completes successfully
- **THEN** system SHALL trigger schema refresh to include newly promoted type

#### Scenario: Display notification to refresh
- **WHEN** promotion succeeds
- **THEN** system SHALL display notification suggesting user refresh schema view

### Requirement: Handle concurrent promotions
The system SHALL prevent concurrent promotion operations that could conflict.

#### Scenario: Disable promotion during active promotion
- **WHEN** a promotion is in progress
- **THEN** system SHALL disable promotion UI and reject new promotion requests

#### Scenario: Show in-progress state
- **WHEN** promotion is executing
- **THEN** system SHALL display progress indicator and block other promotions

#### Scenario: Allow promotion after completion
- **WHEN** previous promotion completes (success or failure)
- **THEN** system SHALL re-enable promotion functionality

### Requirement: Validate promotion performance
The system SHALL complete promotion within acceptable time limits.

#### Scenario: Promotion completes within time limit
- **WHEN** promoting a type with <1000 entities
- **THEN** promotion SHALL complete within 10 seconds

#### Scenario: Show progress during promotion
- **WHEN** promotion is executing
- **THEN** system SHALL display loading state with "Promoting..." message

#### Scenario: Handle promotion timeout
- **WHEN** promotion exceeds reasonable time limit (60 seconds)
- **THEN** system SHALL attempt to cancel and display timeout warning
