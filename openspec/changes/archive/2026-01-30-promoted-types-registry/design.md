## Context

The Enron graph application follows a two-phase workflow:
1. **Extraction**: LLM extracts entities from emails and stores them in a generic `DiscoveredEntity` table
2. **Promotion**: High-frequency entity types are promoted to dedicated Ent schemas with optimized tables

Currently, these phases are disconnected. When the promoter creates a new schema (e.g., `Person.go`), runs `go generate ./ent`, and migrates data, the extractor continues to use the generic table for new entities of that type because it has no awareness of promoted schemas at runtime.

The project uses:
- **Ent framework** for schema-first ORM code generation
- **Go code generation** (`go generate`) to regenerate Ent code when schemas change
- **External command execution** for running migrations and codegen from the promoter

Current file structure:
- `ent/schema/`: Contains schema definitions (e.g., `discoveredentity.go`, `email.go`)
- `ent/generate.go`: Single line directive for `go generate`
- `internal/promoter/`: Handles promotion workflow
- `internal/extractor/`: Extracts entities and stores them via repository

## Goals / Non-Goals

**Goals:**
- Enable runtime detection of promoted schemas by the extractor
- Auto-generate registration code when schemas are promoted (no manual maintenance)
- Route extracted entities to promoted tables when available, falling back to generic storage
- Maintain compatibility with existing generic entity storage path
- Keep the registry in sync with actual Ent schemas (no drift)

**Non-Goals:**
- Changing the promotion workflow or criteria (stays the same)
- Modifying existing Ent schema structures (only adds registration logic)
- Supporting external schema registries or dynamic schema loading
- Hot-reloading or runtime schema changes without recompilation

## Decisions

### 1. Ent Template-Based Code Generation
**Decision**: Use Ent's built-in template system to auto-generate registry initialization code during `go generate ./ent`.

**Rationale**:
- Ent already supports custom templates via `entc.TemplateDir()` option
- Templates run during code generation, ensuring registry stays in sync with schemas
- No manual maintenance required - registration happens automatically when schemas are created
- The promoter already calls `go generate ./ent` after creating schemas

**Alternatives considered**:
- Manual registration in each schema file → Error-prone, easy to forget
- Runtime reflection to discover schemas → Complex, performance overhead, fragile
- Build-time code analysis tool → Additional tooling complexity

**Implementation**:
- Create `ent/template/registry.tmpl` with Go template logic
- Modify `ent/generate.go` to include template directory option
- Template generates `ent/registry_generated.go` with init() function

### 2. Type-Erased EntityCreator Interface
**Decision**: Use `func(ctx context.Context, data map[string]any) (any, error)` as the creation function signature.

**Rationale**:
- Enables a single registry map: `map[string]EntityCreator`
- Extractor already works with `map[string]interface{}` for entity properties from LLM
- Ent's Create() API can be called generically via generated code
- Return type `any` allows uniform handling - caller doesn't need to know specific types

**Alternatives considered**:
- Type-safe generics → Would require registry to know all types at compile time, defeats the purpose
- Interface-based entities → Would require changing all Ent schemas to implement an interface

**Trade-offs**:
- Less type safety (runtime panics if data types mismatch) → Mitigated by schema validation in extractor
- Requires type assertions when using returned entity → Acceptable since caller knows the type name

### 3. Schema Name as Registry Key
**Decision**: Register entities using their Ent schema name (e.g., "Person", "Organization"), matching the table name convention.

**Rationale**:
- Ent schema names already map to table names in a predictable way
- Extractor receives type names from LLM (e.g., "Person") that should match schema names
- Avoids need for separate type mapping or normalization
- Template has direct access to `$node.Name` for registration

**Alternatives considered**:
- Use table names → Extra conversion step, schema name is more canonical
- Use fully qualified Go type names → Unnecessarily verbose, harder to match with LLM output

**Implementation detail**: The registry will normalize type names to match Ent conventions (PascalCase).

### 4. Registry Package Location
**Decision**: Create registry at `internal/registry/registry.go` (not in `ent/` package).

**Rationale**:
- Avoids circular dependencies: `ent/` depends on registry, registry shouldn't depend on `ent/`
- The generated registration code in `ent/registry_generated.go` imports `internal/registry` to call `Register()`
- Keeps registry logic separate from Ent framework code
- Allows extractor to import registry without importing all of `ent/`

**Package structure**:
```
internal/registry/
  registry.go          # Defines EntityCreator, PromotedTypes map, Register()

ent/
  generate.go          # Updated with TemplateDir option
  registry_generated.go # Generated by template, calls registry.Register() in init()
  template/
    registry.tmpl      # Template that generates registration code
```

### 5. Extractor Integration Point
**Decision**: Check registry in `createOrUpdateEntity()` before calling repository's `CreateDiscoveredEntity()`.

**Rationale**:
- This is the centralized creation point for all entity types in the extractor
- Minimal code changes - single lookup before existing fallback logic
- Preserves existing deduplication and merge logic (happens before creation)
- Clean separation: registry check → promoted path OR generic path

**Pseudocode**:
```go
func (e *Extractor) createOrUpdateEntity(ctx, typeCategory, name, ...) {
    // Existing deduplication logic...
    
    // NEW: Check registry first
    if createFn, exists := registry.PromotedTypes[typeCategory]; exists {
        entity, err := createFn(ctx, properties)
        return entity, err
    }
    
    // Existing generic storage fallback
    entity, err := e.repo.CreateDiscoveredEntity(ctx, ...)
    return entity, err
}
```

### 6. Template Field Mapping Strategy
**Decision**: Template generates property setters for all Ent fields using field name matching against the `data` map.

**Rationale**:
- LLM extraction already produces property maps with keys matching expected field names
- Ent template provides field metadata (`$f.Name`, `$f.Type`, `$f.StructField`)
- Type assertions can be generated based on Ent field types
- Missing fields are skipped gracefully (optional properties)

**Generated code pattern**:
```go
if val, ok := data["email"]; ok && val != nil {
    builder.SetEmail(val.(string))
}
if val, ok := data["name"]; ok && val != nil {
    builder.SetName(val.(string))
}
```

**Challenge**: Edge relationships and complex types (time.Time, enums) need special handling in the template.

## Risks / Trade-offs

**Risk**: Template generates invalid code for complex field types (edges, embeds, time.Time)  
**Mitigation**: Start with scalar fields only (string, int, float, bool). Template skips edge fields and special types initially. Can be extended later.

**Risk**: Type assertion panics if LLM provides wrong data types  
**Mitigation**: Wrap creation in recovery, log errors, fall back to generic storage on panic. Add validation layer before calling registry functions.

**Risk**: Registry empty at startup if promoted schemas don't import properly  
**Mitigation**: Generated `registry_generated.go` uses `init()` which runs automatically. Add startup check to log registered types for debugging.

**Risk**: Promoter must remember to run `go generate ./ent` after creating schema  
**Mitigation**: This is already required in the current workflow (promoter calls `RunEntGenerate`). No new dependency introduced.

**Trade-off**: Code generation adds complexity to the build process  
**Accepted**: This is the Ent framework's standard pattern. The project already uses `go generate`. Template approach keeps registry automatically synchronized.

**Trade-off**: Less type safety due to `map[string]any` interface  
**Accepted**: Aligns with existing extractor architecture. Benefits (flexibility, uniform handling) outweigh costs (runtime type checks).
