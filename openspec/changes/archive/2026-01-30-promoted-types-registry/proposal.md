## Why

Currently, when the promoter creates a new Ent type and migrates data to dedicated tables, the extractor has no way to know about these promoted types at runtime. This causes extracted entities to be stored in the generic `DiscoveredEntity` table even when a promoted schema exists for that entity type, defeating the purpose of promotion.

## What Changes

- Introduce a dynamic entity registry that maps promoted type names to their Ent creation functions
- Add an Ent code generation template that auto-registers promoted types during code generation
- Modify the promoter to trigger Ent code generation with the registry template when promoting a type
- Update the extractor to check the registry first before falling back to generic `DiscoveredEntity` storage
- Create a registry package (`internal/registry`) that maintains the mapping of type names to `EntityCreator` functions

## Capabilities

### New Capabilities
- `promoted-type-registry`: Dynamic registration system that maps entity type names to their Ent creation functions, enabling runtime routing of discovered entities to promoted schemas

### Modified Capabilities
- `entity-extraction`: Extraction logic will check the promoted types registry before falling back to generic storage (implementation change, not requirement change - no delta spec needed)

## Impact

**Affected Code:**
- `internal/promoter/`: Must trigger Ent codegen with registry template after schema promotion
- `internal/extractor/`: Must query registry to route entities to correct tables
- `ent/`: New codegen template directory and updated `ent/generate.go`
- New package: `internal/registry/` for the global registry

**Dependencies:**
- Requires Ent template support (already available in entgo.io/ent/entc)
- No external dependency changes

**Database:**
- No schema changes - this is about routing logic, not storage structure

**APIs:**
- No public API changes - internal routing mechanism only
