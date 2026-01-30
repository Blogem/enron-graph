## Why

The promotion workflow is incomplete, causing two interconnected failures:

**Problem 1 - Repository Can't Find Promoted Entities**: Repository methods (FindEntityByID, FindEntityByUniqueID, FindEntitiesByType, SimilaritySearch, TraverseRelationships) only query the `discovered_entities` table, unable to find entities that have been promoted to dedicated tables. This breaks deduplication in the extractor (causing duplicate storage), prevents relationship traversal across promoted types, and makes the promoted types registry incomplete—it can create promoted entities but can't retrieve them.

**Problem 2 - Promoter Doesn't Update Relationships**: The promoter's `CopyEntities` method has a TODO (line 165) indicating relationships are NOT updated when entities are promoted to ANY type (person, organization, contract, etc.). When entities are promoted:
- New rows are inserted into the promoted table with NEW auto-increment IDs (e.g., discovered_entities.id=123 → persons.id=5)
- Relationships table still references `discovered_entity` type and old ID=123
- Should reference `person` type and new ID=5
- This affects ANY type that gets promoted, as the promoter is fully type-agnostic
- TraverseRelationships filters by entity type and misses all promoted entities entirely

These two problems compound: even if we fix the repository to query promoted tables, the relationship data would be inconsistent. We must fix both to make promotion work correctly.

## What Changes

**Repository Changes (Query Across Tables):**
- Add optional `typeHint` parameter to FindEntityByID and FindEntityByUniqueID (backward compatible via variadic params)
- Implement type-aware lookup strategy: direct table query when type known (O(1)), fallback to relationship table inference, last resort parallel goroutine search
- Extend registry with EntityFinder functions (mirrors existing EntityCreator pattern)
- Update FindEntitiesByType to query promoted tables when type matches
- Update SimilaritySearch to use UNION query across all tables with embeddings
- Update TraverseRelationships to pass entity type hints from relationship metadata
- Update extractor and deduplicator to pass type hints at all call sites

**Promoter Changes (Update Relationships During Migration):**
- Track old_id → new_id mapping during entity insertion (use RETURNING clause)
- After migrating entities, update relationships table to reference promoted type and new IDs
- Update both from_type/from_id and to_type/to_id for all affected relationships
- Optionally delete migrated entities from discovered_entities table (configurable)
- Add migration statistics to PromotionResult (relationships updated count)

- `schema-promotion`: Promoter updates relationships table during migration (from_type/to_type and ID mapping), completing the TODO at line 165. Works for ANY promoted type (person, organization, contract, etc.)
- `type-aware-entity-queries`: Query entities across promoted and discovered tables with performance-optimized lookup strategies (type hints, relationship inference, parallel search)

### Modified Capabilities
- `entity-extraction`: Extractor passes typeCategory as hint to FindEntityByUniqueID, enabling O(1) promoted table lookups
- `promoted-type-registry`: Registry extended with EntityFinder functions alongside EntityCreator functions, generated via Ent templates

**Future Work** (separate change):
- `schema-promotion`: Promoter should update relationships table when migrating ANY entity type (from_type/to_type and ID mapping). Currently has TODO at line 165 in promoter.go. The promoter is type-agnostic (works with person, organization, contract, or any discovered type), but relationship updates are missing for all types. This change provides workaround by querying both tables.

## Impact

**Affected Code:**
- `internal/registry/registry.go`: Add EntityFinder type and PromotedFinders map (~20 lines)
- `ent/template/registry.tmpl`: Generate findXxx() functions for each schema (~30-40 template lines)
- `internal/graph/repository.go`: Update 2 method signatures with optional typeHint parameter
- `internal/graph/repository_impl.go`: Implement type-aware logic for 6 methods (~150-200 lines)
- `internal/extractor/extractor.go`: Pass typeCategory to 3 FindEntityByUniqueID calls (3 line changes)
- `internal/extractor/dedup.go`: Pass type hints to 3 FindEntityByUniqueID calls (3 line changes)
- Test files: Update mocks and add optional parameters (~50-100 lines)

**`internal/promoter/promoter.go`: Update CopyEntities to track ID mapping and update relationships (~80-100 lines)
- Test files: Update mocks, add optional parameters, test relationship updates (~100-15
- With type hint: 1 query (direct table lookup) - current best case maintained
- Without type hint: 1-3 queries (discovered → relationships → promoted) - acceptable for rare cases
- No performance degradation for existing code paths

**Dependencies:**
- No external dependency changes
- Requires `go generate ./ent` to regenerate registry with finders

**Database:**
- No schema changes
- Promoter will UPDATE relationships table during migration (DML, not DDL)
- Option to DELETE migrated rows from discovered_entities (configurable cleanup)
- Leverages existing relationship table's from_type/to_type metadata

**APIs:**
- Backward compatible: optional parameters don't break existing callers
- Internal API only (repository interface)
