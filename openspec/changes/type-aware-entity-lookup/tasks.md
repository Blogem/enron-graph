## 1. Registry Enhancement (EntityFinder) - TDD

- [x] 1.1 Write test for EntityFinder type definition and registration in internal/registry/registry_test.go
- [x] 1.2 Add EntityFinder type definition to internal/registry/registry.go (mirrors EntityCreator signature)
- [x] 1.3 Add PromotedFinders global map to internal/registry/registry.go
- [x] 1.4 Add RegisterFinder() function to internal/registry/registry.go
- [x] 1.5 Verify test passes for EntityFinder registration
- [x] 1.6 Write test for findPerson() function in ent/registry_test.go
- [x] 1.7 Generate findPerson() function in ent/registry.go (queries persons table by unique_id)
- [x] 1.8 Generate findOrganization() function in ent/registry.go (queries organizations table by unique_id)
- [x] 1.9 Generate findContract() function in ent/registry.go (queries contracts table by unique_id)
- [x] 1.10 Add finder registration calls to init() in ent/registry.go for all promoted types
- [x] 1.11 Verify all finder tests pass

## 2. Repository Interface Changes (Type Hints)

- [x] 2.1 Add variadic typeHint parameter to FindEntityByID in internal/graph/repository.go interface
- [x] 2.2 Add variadic typeHint parameter to FindEntityByUniqueID in internal/graph/repository.go interface
- [x] 2.3 Add variadic typeHint parameter to FindEntitiesByType in internal/graph/repository.go interface
- [x] 2.4 Update repository method documentation to explain type hint usage and fallback behavior

## 3. Repository Type-Aware Queries (FindEntityByUniqueID) - TDD

- [x] 3.1 Write test for FindEntityByUniqueID with type hint (direct table lookup)
- [x] 3.2 Add helper function to check if type is promoted (queries PromotedTypes registry)
- [x] 3.3 Add helper function to get table name from type name (handles pluralization)
- [x] 3.4 Implement direct table lookup when type hint provided
- [x] 3.5 Verify test passes for type hint lookup
- [x] 3.6 Write test for tier 1 fallback (discovered_entities)
- [x] 3.7 Ensure tier 1 logic exists (should already work)
- [x] 3.8 Write test for tier 2 fallback (relationships inference)
- [x] 3.9 Implement tier 2: Query relationships table to infer entity type
- [x] 3.10 Verify test passes for tier 2
- [x] 3.11 Write test for tier 3 fallback (parallel search)
- [x] 3.12 Add worker pool implementation for parallel search with max 5 concurrent queries
- [x] 3.13 Add context cancellation for parallel search (early exit on first match)
- [x] 3.14 Implement tier 3: Parallel search across promoted tables using EntityFinder registry
- [x] 3.15 Verify test passes for tier 3

## 4. Repository Type-Aware Queries (FindEntityByID) - TDD

- [x] 4.1 Write test for FindEntityByID with type hint
- [x] 4.2 Implement type-aware FindEntityByID with direct table lookup when hint provided
- [x] 4.3 Verify test passes
- [x] 4.4 Write test for FindEntityByID without hint (relationships inference)
- [x] 4.5 Implement relationships inference fallback for FindEntityByID
- [x] 4.6 Verify test passes

## 5. Repository Type-Aware Queries (Other Methods) - TDD

- [x] 5.1 Write test for FindEntitiesByType with promoted type
- [x] 5.2 Implement type-aware FindEntitiesByType to query promoted tables
- [x] 5.3 Verify test passes
- [x] 5.4 Write test for SimilaritySearch across promoted tables
- [x] 5.5 Update SimilaritySearch to generate UNION query across all tables with embeddings
- [x] 5.6 Verify test passes
- [x] 5.7 Write test for TraverseRelationships with promoted entities
- [x] 5.8 Update TraverseRelationships to pass to_type as hint when fetching target entities
- [x] 5.9 Verify test passes
- [x] 5.10 Write test for GetDistinctEntityTypes returning promoted types from relationships
- [x] 5.11 Update GetDistinctEntityTypes to query relationships table for promoted types
- [x] 5.12 Verify test passes

## 6. Extractor Integration (Type Hint Passing) - TDD

- [x] 6.1 Write test for extractor passing type hints to repository methods
- [x] 6.2 Update createOrUpdateEntity in internal/extractor/extractor.go to pass TypeCategory as type hint to FindEntityByUniqueID
- [x] 6.3 Update deduplication logic in internal/extractor/dedup.go to pass type hint to SimilaritySearch
- [x] 6.4 Update deduplication exact match check to pass type hint to FindEntityByUniqueID
- [x] 6.5 Verify test passes
- [x] 6.6 Write test for extractor using registry for promoted type creation
- [x] 6.7 Add logic to check PromotedTypes registry before creating entity in extractor
- [x] 6.8 Route entity creation through registry EntityCreator for promoted types
- [x] 6.9 Fallback to discovered_entities creation for non-promoted types
- [x] 6.10 Verify test passes

## 7. Promoter ID Mapping (RETURNING clause) - TDD

- [x] 7.1 Write test for ID mapping with RETURNING clause
- [x] 7.2 Add ID mapping tracking to CopyEntities in internal/promoter/promoter.go (map[int]int for old→new IDs)
- [x] 7.3 Modify INSERT query to include RETURNING id clause
- [x] 7.4 Parse RETURNING results and build ID mapping
- [x] 7.5 Verify test passes

## 8. Promoter Relationship Updates - TDD

- [x] 8.1 Write test for updating from_type and from_id references
- [x] 8.2 Add updateRelationships() function to update from_type and from_id references
- [x] 8.3 Use WHERE from_type = ? filter for efficient FROM reference updates
- [x] 8.4 Implement batch UPDATE with IN clause for relationship IDs
- [x] 8.5 Add batch size limit of 1000 rows per UPDATE statement
- [x] 8.6 Verify test passes
- [x] 8.7 Write test for updating to_type and to_id references
- [x] 8.8 Add updateRelationships() function to update to_type and to_id references
- [x] 8.9 Use WHERE to_type = ? filter for efficient TO reference updates
- [x] 8.10 Verify test passes

## 9. Entity Cleanup - TDD

- [x] 9.1 Write test for entity cleanup DELETE operation
- [x] 9.2 Add deleteOldEntities() function to internal/promoter/promoter.go
- [x] 9.3 Use ID mapping to identify discovered_entities rows for deletion
- [x] 9.4 Execute DELETE with WHERE id IN (...) using old IDs
- [x] 9.5 Add configuration flag to disable cleanup if needed (default: enabled)
- [x] 9.6 Verify test passes

## 10. Transaction Management - TDD

- [x] 10.1 Write test for transaction atomicity (all steps succeed together)
- [x] 10.2 Wrap CopyEntities INSERT in database transaction
- [x] 10.3 Execute relationship updates within same transaction
- [x] 10.4 Execute entity cleanup DELETE within same transaction
- [x] 10.5 Commit transaction only when all steps succeed
- [x] 10.6 Verify test passes
- [x] 10.7 Write test for transaction rollback on failure
- [x] 10.8 Add rollback handling for any step failure
- [x] 10.9 Add transaction timeout of 60 seconds
- [x] 10.10 Verify rollback test passes

## 11. Validation & Error Handling - TDD

- [x] 11.1 Write test for duplicate unique_id handling
- [x] 11.2 Add duplicate unique_id detection before INSERT in promoter
- [x] 11.3 Skip INSERT for entities with duplicate unique_ids (log warning)
- [x] 11.4 Update relationships even for skipped entities (use existing promoted ID)
- [x] 11.5 Verify test passes
- [x] 11.6 Write test for orphaned relationship validation
- [x] 11.7 Add validation to check for orphaned relationship references after migration
- [x] 11.8 Log validation errors with affected relationship IDs
- [x] 11.9 Add optional strict validation mode (fails migration if inconsistencies found)
- [x] 11.10 Add error wrapping with context for all promoter errors
- [x] 11.11 Verify validation tests pass

## 12. Integration Tests

- [x] 12.1 Create integration test for full promotion workflow (person type)
- [x] 12.2 Implement full workflow if needed to pass integration test
- [x] 12.3 Verify entities copied to persons table
- [x] 12.4 Verify relationships updated to reference persons table
- [x] 12.5 Verify old entities deleted from discovered_entities
- [x] 12.6 Verify transaction atomicity (rollback on failure)
- [x] 12.7 Test FindEntityByUniqueID can find promoted entities
- [x] 12.8 Test TraverseRelationships works with promoted entities
- [x] 12.9 Test extractor creates entities in promoted tables
- [x] 12.10 Test extractor dedup finds promoted entities
- [x] 12.11 Test SimilaritySearch returns results from promoted tables

## 13. Logging & Observability

- [x] 13.1 Add logging for entity migration counts (migrated, skipped, deleted)
- [x] 13.2 Add logging for relationship update counts (from_type and to_type)
- [x] 13.3 Add logging for transaction duration
- [x] 13.4 Add logging for tier 1/2/3 fallback path usage in repository
- [x] 13.5 Add metrics for parallel search performance (queries executed, early exits)

## 14. Documentation & Cleanup

- [x] 14.1 Remove TODO comment at line 165 in internal/promoter/promoter.go
- [x] 14.2 Add godoc comments for new EntityFinder type
- [x] 14.3 Add godoc comments for three-tier fallback strategy
- [x] 14.4 Add godoc comments for ID mapping approach
- [x] 14.5 Update DATABASE.md with promoted tables relationship handling
- [x] 14.6 Update README.md with type-aware query capabilities
- [x] 14.7 Add code examples showing type hint usage in repository calls
