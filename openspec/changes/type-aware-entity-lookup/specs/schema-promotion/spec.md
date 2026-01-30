## MODIFIED Requirements

### Requirement: Promoter updates relationships table during entity migration
**ADDED**: The system SHALL maintain relationship integrity when migrating entities from discovered_entities to promoted tables.

#### Scenario: Track ID mapping during INSERT
- **WHEN** promoter copies entities to promoted table
- **THEN** promoter uses RETURNING clause to capture new auto-increment IDs

#### Scenario: Build old-to-new ID mapping
- **WHEN** entities are inserted into promoted table
- **THEN** promoter creates map[int]int of discovered.id → promoted.id

#### Scenario: Update FROM references in relationships
- **WHEN** relationships table has rows with from_type matching migrated type
- **THEN** promoter updates from_type to promoted type name and from_id to new ID

#### Scenario: Update TO references in relationships
- **WHEN** relationships table has rows with to_type matching migrated type
- **THEN** promoter updates to_type to promoted type name and to_id to new ID

#### Scenario: Preserve relationship metadata
- **WHEN** updating relationship references
- **THEN** promoter preserves relationship_type, email_source_id, and all other metadata

### Requirement: Entity migration and relationship updates are atomic
**ADDED**: The system SHALL execute all migration steps within a single database transaction.

#### Scenario: Transaction includes INSERT
- **WHEN** promoter starts migration
- **THEN** entity INSERT operations occur within transaction

#### Scenario: Transaction includes relationship UPDATE
- **WHEN** promoter updates relationships table
- **THEN** UPDATE operations occur within same transaction as INSERT

#### Scenario: Transaction includes cleanup
- **WHEN** promoter deletes old discovered_entities rows
- **THEN** DELETE operations occur within same transaction

#### Scenario: Rollback on any failure
- **WHEN** any step fails (INSERT, UPDATE, or DELETE)
- **THEN** transaction rolls back and no changes are committed

#### Scenario: Commit only when all steps succeed
- **WHEN** all migration steps complete successfully
- **THEN** transaction commits and all changes persist

### Requirement: Promoter cleans up migrated entities from discovered_entities
**ADDED**: The system SHALL remove migrated entities from discovered_entities table by default.

#### Scenario: Delete after successful migration
- **WHEN** entities are successfully copied to promoted table
- **THEN** promoter deletes corresponding rows from discovered_entities

#### Scenario: Use ID mapping for cleanup
- **WHEN** deleting old entities
- **THEN** promoter uses tracked old IDs to identify rows for deletion

#### Scenario: Cleanup within transaction
- **WHEN** cleanup executes
- **THEN** DELETE operation is part of migration transaction

#### Scenario: Skip cleanup if disabled
- **WHEN** cleanup is explicitly disabled via configuration
- **THEN** promoter leaves discovered_entities rows unchanged

### Requirement: ID mapping uses RETURNING clause for efficiency
**ADDED**: The system SHALL capture new IDs in single roundtrip using PostgreSQL RETURNING clause.

#### Scenario: INSERT with RETURNING id
- **WHEN** promoter executes batch INSERT
- **THEN** query includes RETURNING id clause

#### Scenario: Parse RETURNING results
- **WHEN** INSERT completes
- **THEN** promoter parses returned IDs in order matching input rows

#### Scenario: Match old IDs to new IDs
- **WHEN** building ID mapping
- **THEN** promoter pairs each input row's old ID with corresponding RETURNING ID

### Requirement: Relationship updates handle partial matches
**ADDED**: The system SHALL correctly update relationships even when only some entities are promoted.

#### Scenario: Update only relevant relationships
- **WHEN** updating from_type references
- **THEN** promoter filters WHERE from_type = old_type before UPDATE

#### Scenario: Preserve non-migrated references
- **WHEN** relationship references entity that was not migrated
- **THEN** promoter leaves that reference unchanged

#### Scenario: Handle mixed type references
- **WHEN** relationship connects promoted entity to non-promoted entity
- **THEN** promoter updates only the promoted side of reference

### Requirement: Promoter validates relationship integrity after migration
**ADDED**: The system SHALL verify relationship consistency after completing migration.

#### Scenario: Check orphaned references
- **WHEN** migration completes
- **THEN** promoter queries for relationships referencing non-existent IDs

#### Scenario: Log validation errors
- **WHEN** validation finds inconsistencies
- **THEN** promoter logs error details with affected relationship IDs

#### Scenario: Optional strict mode
- **WHEN** strict validation is enabled
- **THEN** promoter returns error and prevents migration if validation fails

### Requirement: Batch updates for performance
**ADDED**: The system SHALL update relationships in efficient batches using SQL.

#### Scenario: Bulk UPDATE with IN clause
- **WHEN** updating multiple relationship rows
- **THEN** promoter uses single UPDATE with WHERE id IN (...) clause

#### Scenario: Batch size limit
- **WHEN** number of relationships exceeds batch threshold (e.g., 1000)
- **THEN** promoter splits updates into multiple batches

#### Scenario: Concurrent batch execution
- **WHEN** multiple batches exist
- **THEN** promoter executes batches sequentially within transaction

### Requirement: Migration handles duplicate unique_ids
**ADDED**: The system SHALL detect and handle entities with duplicate unique_ids during migration.

#### Scenario: Detect duplicates before INSERT
- **WHEN** promoter queries discovered_entities for migration
- **THEN** promoter checks for duplicate unique_id values

#### Scenario: Skip duplicate inserts
- **WHEN** promoted table already contains entity with unique_id
- **THEN** promoter skips INSERT and logs warning

#### Scenario: Update relationships for skipped entities
- **WHEN** entity is skipped due to duplicate
- **THEN** promoter still updates relationships to reference existing promoted entity

### Requirement: Promoter logs migration statistics
**ADDED**: The system SHALL provide detailed reporting of migration operations.

#### Scenario: Log entity counts
- **WHEN** migration completes
- **THEN** promoter logs number of entities migrated, skipped, and deleted

#### Scenario: Log relationship update counts
- **WHEN** relationship updates complete
- **THEN** promoter logs number of from_type and to_type references updated

#### Scenario: Log transaction duration
- **WHEN** transaction commits
- **THEN** promoter logs total time spent in migration transaction
