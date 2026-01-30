## Context

The Enron graph application uses a two-phase entity lifecycle:
1. **Discovery**: LLM extracts entities from emails → stored in generic `discovered_entities` table
2. **Promotion**: High-frequency types promoted to dedicated Ent schemas with optimized tables

Current architecture has a promoted types registry that routes *creation* to promoted tables, but repository methods only query `discovered_entities`, missing promoted entities entirely. Additionally, when the promoter migrates entities to promoted tables, it doesn't update the relationships table, leaving stale references.

**Current State:**
- Registry has `PromotedTypes` map with `EntityCreator` functions (create path works)
- Repository methods hardcoded to query `discovered_entities` (fetch path broken)
- Promoter's `CopyEntities` has TODO at line 165 about relationship updates
- Entities get new auto-increment IDs in promoted tables (discovered.id=123 → persons.id=5)
- Relationships still reference old type and ID (`discovered_entity`, 123 instead of `person`, 5)

**Constraints:**
- Ent generates code, so we work within generated structures
- Performance critical: extractor processes 1000s of entities
- Postgres-specific (can use RETURNING, UNION, etc.)
- Relationships table is source of truth for graph structure

## Goals / Non-Goals

**Goals:**
- Enable repository to query both discovered and promoted entity tables
- Provide O(1) lookups when type is known (performance optimization)
- Fix promoter to update relationships during migration (complete the workflow)
- Work with ANY promoted type (type-agnostic solution)

**Non-Goals:**
- Changing Ent's code generation patterns
- Modifying database schema (no DDL)
- Fixing existing broken data (future migration tool can handle that)
- Polymorphic entity interface (future refactor, keep *ent.DiscoveredEntity for now)
- Real-time sync between tables (acceptable to have temporary duplicates)

## Decisions

### Decision 1: Optional Type Hint Pattern

**Choice:** Add optional variadic `typeHint` parameter to repository methods

```go
// Before
FindEntityByUniqueID(ctx, uniqueID string) (*ent.DiscoveredEntity, error)

// After (backward compatible)
FindEntityByUniqueID(ctx, uniqueID string, typeHint ...string) (*ent.DiscoveredEntity, error)
```

**Rationale:**
- Call sites that have type information can pass it for O(1) lookups
- Mirrors Go stdlib patterns (e.g., `context.WithTimeout`)
- Alternative (new method like `FindEntityByUniqueIDWithType`) would fragment the API

**Alternatives Considered:**
- Separate `FindPromotedEntity` method → duplicates logic, forces callers to know if promoted
- Return `interface{}` instead of `*ent.DiscoveredEntity` → breaks all callers, type unsafe
- Always scan all tables → too slow, defeats purpose of promotion

### Decision 2: Three-Tier Lookup Strategy

**Choice:** Implement fallback chain when type hint absent

1. **Fast path**: Type hint provided → query specific table (O(1))
2. **Smart fallback**: Query `discovered_entities` first (likely location)
3. **Relationship inference**: Query relationships table for type hint
4. **Last resort**: Parallel goroutine search across all promoted tables

**Rationale:**
- Prioritizes common case (90%+ of entities still in discovered_entities)
- Relationship table has `from_type`/`to_type` metadata we can leverage
- Parallel search amortizes worst-case cost across multiple tables
- Each tier exits early on success

**Alternatives Considered:**
- Always scan all tables in sequence → O(N) every time, too slow
- UNION query across all tables → complex with different schemas, SQL injection risk
- Application-level cache → memory overhead, invalidation complexity, multi-instance issues

### Decision 3: Registry Symmetry (EntityFinder)

**Choice:** Add `EntityFinder` functions to registry alongside `EntityCreator`

```go
type EntityFinder func(ctx context.Context, uniqueID string) (any, error)
var PromotedFinders = make(map[string]EntityFinder)
```

**Rationale:**
- Mirrors existing `EntityCreator` pattern (developers already understand it)
- Generated via same Ent templates (consistency)
- Registry becomes complete: can both create AND find promoted entities
- Enables dynamic discovery of promoted types at runtime

**Alternatives Considered:**
- Reflection to call Ent's generated Query methods → fragile, breaks on Ent changes
- Manual registration in init() → error-prone, easy to forget
- SQL-only approach → bypasses Ent's abstractions, loses type safety

### Decision 4: Promoter ID Mapping with RETURNING

**Choice:** Use PostgreSQL RETURNING clause to track old_id → new_id mapping

```go
INSERT INTO persons (unique_id, name, ...) VALUES ($1, $2, ...) RETURNING id
// Store: oldToNewIDMap[entity.ID] = returnedID
```

**Rationale:**
- Single query to insert and get new ID (efficient)
- Postgres supports RETURNING natively
- Maps naturally to Go map for batch updates
- Alternative (query after insert) requires extra round-trips

**Implementation:**
```go
oldToNewIDMap := make(map[int]int)
for _, entity := range entities {
    var newID int
    err := tx.QueryRow("INSERT ... RETURNING id", values...).Scan(&newID)
    oldToNewIDMap[entity.ID] = newID
}

// Then update relationships
for oldID, newID := range oldToNewIDMap {
    UPDATE relationships 
    SET from_type=$1, from_id=$2 
    WHERE from_type='discovered_entity' AND from_id=$3
}
```

**Alternatives Considered:**
- Use `unique_id` for correlation → works but complex nested queries, slower
- Sequential SELECT after INSERT → N extra queries
- Don't track mapping, keep duplicates → leaves data inconsistent

### Decision 5: Relationship Updates in Same Transaction

**Choice:** Update relationships within the same transaction as entity migration

**Rationale:**
- Atomic: either all succeeds or all rolls back
- Prevents partial migration state
- Relationships always consistent with entity storage
- Existing code already uses transactions in `CopyEntities`

**Trade-off:** Longer transaction time, but promotions are infrequent (acceptable)

### Decision 6: Cleanup Migrated Entities

**Choice:** Delete old `discovered_entities` rows after successful promotion (default: delete)

```go
type PromotionRequest struct {
    ...
    KeepMigratedEntities bool // default: false (i.e., delete old rows)
}
```

**Rationale:**
- Clean data model (no duplicates)
- Single source of truth after promotion
- Reduces storage and query complexity
- Migration is atomic (transaction ensures consistency)
- Optional flag allows keeping old rows if needed for debugging/rollback testing

**Trade-off:** Cannot easily rollback promotion without re-migration, but transaction safety ensures data integrity

## Risks / Trade-offs

### Risk: Performance degradation when type hint missing
**Impact:** Without hint, 1-3 queries instead of 1  
**Mitigation:** 
- Relationship table lookup is fast (indexed on from_id/to_id)
- Parallel search amortizes worst case
- Extractor/dedup hot paths always have type available
- Rare case (API, TUI) acceptable to be slightly slower

### Risk: Promoted entities get new IDs, breaking external references
**Impact:** If anything caches entity IDs, they become stale  
**Mitigation:**
- All internal code uses repository methods (will query correct table)
- `unique_id` is stable identifier across promotion
- External systems should use `unique_id`, not numeric ID

### Risk: RETURNING clause requires PostgreSQL
**Impact:** Won't work with other databases  
**Mitigation:**
- Project already PostgreSQL-only (pgvector dependency)
- Document this as Postgres-specific in promoter code
- Fallback: could do SELECT after INSERT for other DBs (slower)

### Risk: Parallel goroutine search could spike connections
**Impact:** If many promoted tables, many concurrent queries  
**Mitigation:**
- Limit concurrency (use worker pool or semaphore)
- Exit early on first match (most won't need full scan)
- Connection pooling handles burst

### Risk: Relationship updates might miss complex edge cases
**Impact:** Some relationships might not update correctly  
**Mitigation:**
- Update both from_type/from_id and to_type/to_id in separate statements
- Transaction ensures consistency
- Add comprehensive tests for edge cases (self-references, circular, etc.)

### Risk: Type hint could be wrong (caller passes incorrect type)
**Impact:** Query wrong table, return nil when entity exists elsewhere  
**Mitigation:**
- Fallback chain catches this (tries discovered_entities if promoted fails)
- Log warning when hint fails but entity found in fallback
- Tests verify hint accuracy

## Migration Plan

**Deployment Steps:**

1. **Phase 1: Registry Extension**
   - Deploy registry changes with EntityFinder functions
   - Run `go generate ./ent` to regenerate code
   - No runtime impact (new code unused)

2. **Phase 2: Repository Updates**
   - Deploy repository methods with optional typeHint parameter
   - Backward compatible (existing calls work)
   - New extractor/dedup calls use hints

3. **Phase 3: Promoter Updates**
   - Deploy promoter relationship updates
   - Test on low-frequency type first
   - Monitor relationship table updates

4. **Phase 4: Existing Data (Future)**
   - Separate migration tool to fix existing broken relationships
   - Out of scope for this change

**Rollback Strategy:**
- Phase 1-2: Safe to revert (backward compatible)
- Phase 3: Revert promoter, but new promotions won't have relationship updates
- No data corruption risk (transactions prevent partial state)

**Monitoring:**
- Log when type hint fails but fallback succeeds (indicates incorrect hints)
- Track query counts per lookup type (measure fallback frequency)
- Monitor promotion statistics (relationships updated count)

## Open Questions

None - all major decisions resolved during exploration phase.
