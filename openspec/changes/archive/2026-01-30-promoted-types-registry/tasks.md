## 1. Write Registry Package Tests (RED)

- [x] 1.1 Create `internal/registry/registry_test.go`
- [x] 1.2 Write test for Register() function storing EntityCreator in map
- [x] 1.3 Write test for looking up existing promoted type
- [x] 1.4 Write test for looking up non-existent type (returns nil/false)
- [x] 1.5 Write test for multiple registrations (no collision)
- [x] 1.6 Run tests and verify they fail (no implementation yet)

## 2. Implement Registry Package (GREEN)

- [x] 2.1 Create `internal/registry/registry.go` with EntityCreator type definition
- [x] 2.2 Add PromotedTypes map[string]EntityCreator global variable
- [x] 2.3 Implement Register(typeName string, fn EntityCreator) function
- [x] 2.4 Add package documentation explaining registry purpose
- [x] 2.5 Run tests and verify they pass

## 3. Write Extractor Integration Tests (RED)

- [x] 3.1 Create test in `internal/extractor/extractor_test.go` for promoted type routing
- [x] 3.2 Write test: entity with promoted type should use registry creator
- [x] 3.3 Write test: entity with non-promoted type should use DiscoveredEntity
- [x] 3.4 Write test: registry creator error should fallback to DiscoveredEntity
- [x] 3.5 Write test: registry creator with nil value fields should skip them
- [x] 3.6 Run tests and verify they fail (no implementation yet)

## 4. Implement Extractor Integration (GREEN)

- [x] 4.1 Import `internal/registry` in extractor package
- [x] 4.2 Locate `createOrUpdateEntity` function in `internal/extractor/extractor.go`
- [x] 4.3 Add registry lookup before CreateDiscoveredEntity call
- [x] 4.4 Add context setup with Ent client for EntityCreator functions
- [x] 4.5 Call EntityCreator function if type exists in registry
- [x] 4.6 Add error handling with logging for registry creation failures
- [x] 4.7 Preserve fallback to CreateDiscoveredEntity for non-promoted types
- [x] 4.8 Preserve fallback to CreateDiscoveredEntity on registry errors
- [x] 4.9 Run tests and verify they pass

## 5. Update Ent Code Generation (GREEN)

- [x] 5.1 Modify `ent/generate.go` to import entgo.io/ent/entc and entgo.io/ent/entc/gen
- [x] 5.2 Replace simple directive with main() function
- [x] 5.3 Add entc.TemplateDir("./template") option
- [x] 5.4 Call entc.Generate with schema directory and template option
- [x] 5.5 Add error handling and logging

## 6. Create Ent Template (GREEN)

- [x] 6.1 Create `ent/template/` directory
- [x] 6.2 Create `ent/template/registry.tmpl` with template header
- [x] 6.3 Add template logic to iterate over all schema nodes
- [x] 6.4 Generate registration function for each schema with context.Context parameter
- [x] 6.5 Add Ent client extraction from context logic
- [x] 6.6 Generate Create() builder initialization for each schema
- [x] 6.7 Add field mapping logic for scalar types (string, int, float, bool)
- [x] 6.8 Add type assertion and setter calls for each field
- [x] 6.9 Add nil/missing field checks to skip optional fields
- [x] 6.10 Generate Save() call and return statement
- [x] 6.11 Add init() function that registers all schemas

## 7. Validate Template Output (GREEN)

- [x] 7.1 Run `go generate ./ent` to verify template executes without errors
- [x] 7.2 Verify `ent/registry.go` is created with registration code
- [x] 7.3 Inspect generated code for existing schemas (Email, DiscoveredEntity, etc.)
- [x] 7.4 Verify generated code compiles without errors
- [x] 7.5 Run all tests and verify they still pass

## 8. Write Integration Tests (RED)

- [x] 8.1 Create test promoted schema `ent/schema/testperson.go` for testing
- [x] 8.2 Write integration test: codegen creates registration for TestPerson
- [x] 8.3 Write integration test: extractor routes TestPerson to promoted table
- [x] 8.4 Write integration test: promotion workflow updates registry
- [x] 8.5 Write test: end-to-end extract → promote → extract (uses promoted table)
- [x] 8.6 Run integration tests (may fail initially)

## 9. Complete Integration (GREEN)

- [x] 9.1 Verify promoter already calls RunEntGenerate after creating schema
- [x] 9.2 Fix any issues found by integration tests
- [x] 9.3 Run full promotion workflow: create schema → generate → migrate → verify registry
- [x] 9.4 Verify new promoted type appears in registry after application restart
- [x] 9.5 Run all integration tests and verify they pass

## 10. Refactor and Document

- [x] 10.1 Add comments to registry.go explaining usage
- [x] 10.2 Add comments to template explaining generation logic
- [x] 10.3 Document edge case: only scalar fields supported initially
- [x] 10.4 Add TODO for future edge/relationship field support
- [x] 10.5 Update README if needed to explain registry mechanism
- [x] 10.6 Review all tests for clarity and completeness
- [x] 10.7 Run full test suite and verify all tests pass
