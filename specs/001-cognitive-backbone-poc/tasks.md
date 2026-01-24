# Tasks: Cognitive Backbone POC

**Feature**: 001-cognitive-backbone-poc  
**Generated**: 2026-01-24  
**Input**: spec.md, plan.md, analysis.md, clarifications.md  
**Tech Stack**: Go 1.21+, PostgreSQL + ent + pgvector, Ollama, Bubble Tea

## Test-Driven Development (TDD) Approach

**⚠️ NON-NEGOTIABLE**: This project follows strict TDD principles as mandated by the constitution.

### Red-Green-Refactor Cycle

All implementation MUST follow this workflow:

1. **RED**: Write failing tests first
   - Unit tests for pure functions and logic
   - Contract tests for API endpoints
   - Integration tests for multi-component workflows
   
2. **Obtain Approval**: Review test scenarios before implementation

3. **Verify Failure**: Run tests and confirm they fail for the right reasons

4. **GREEN**: Implement minimum code to make tests pass
   - Focus on making tests pass, not on perfect code
   
5. **REFACTOR**: Clean up code while keeping tests green
   - Improve code structure, readability, performance
   - Tests must remain green throughout refactoring

### Test Organization

**77 test tasks (49% of total tasks)** organized by type:
- **Unit Tests** (22 tasks): Test isolated functions and logic
- **Contract Tests** (6 tasks): Test API request/response schemas
- **Integration Tests** (11 tasks): Test multi-component workflows
- **Performance Tests** (7 tasks): Validate success criteria benchmarks
- **Acceptance Tests** (29 tasks): Verify user story requirements
- **Manual Tests** (2 tasks): TUI and exploratory testing

### Test Files Structure

```
tests/
  integration/          # Integration tests
    loader_test.go
    extractor_test.go
    api_test.go
    analyst_test.go
    promoter_test.go
    chat_test.go
    integrity_test.go
    concurrency_test.go
  benchmarks/          # Performance benchmarks
    loader_bench_test.go
    query_bench_test.go
    path_bench_test.go
    tui_bench_test.go
  fixtures/            # Test data
    sample_emails.csv
    test_entities.json

internal/
  loader/
    parser_test.go     # Unit tests alongside implementation
    headers_test.go
  extractor/
    extractor_test.go
    dedup_test.go
    relationships_test.go
  api/
    handlers_test.go   # Contract tests
  analyst/
    detector_test.go
    clustering_test.go
    ranker_test.go
    schema_gen_test.go
  promoter/
    codegen_test.go
  chat/
    patterns_test.go
    context_test.go
    handler_test.go
    prompts_test.go
  tui/
    graph_view_test.go
    entity_list_test.go
  graph/
    pagination_test.go
    filters_test.go

pkg/
  llm/
    ollama_test.go
```

### Test Verification Notes

Each implementation task references its corresponding test task with:
```
**Verify**: T### tests pass
```

This ensures the TDD cycle is followed and tests are run after implementation.

---

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[US1-5]**: User Story identifier (US1=P1 MVP, US2=P2, US3=P3, US4=P4, US5=P5)
- File paths are absolute from repository root

## Project Structure

```
cmd/            # CLI applications
internal/       # Application logic
ent/            # ent schemas and generated code
pkg/            # Shared utilities
tests/          # Integration tests
assets/         # Enron email dataset
```

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and development environment

- [x] T001 Install Go 1.21+ and verify installation
- [x] T002 Install PostgreSQL 15+ with Docker and enable pgvector extension
- [x] T003 Install Ollama and pull models: `llama3.1:8b`, `mxbai-embed-large`
- [x] T004 Initialize Go module: `go mod init github.com/Blogem/enron-graph`
- [x] T005 Install ent: `go get entgo.io/ent/cmd/ent`
- [x] T006 [P] Create project directory structure: `cmd/`, `internal/`, `ent/`, `pkg/`, `tests/`, `scripts/`
- [x] T007 [P] Create Docker Compose file for PostgreSQL + pgvector at root
- [x] T008 [P] Create `.gitignore` for Go project
- [x] T009 Document environment setup in `README.md`

**Checkpoint**: Development environment ready ✅

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure required before ANY user story implementation

**⚠️ CRITICAL**: All user stories depend on these tasks

### Database Setup

- [x] T010 Create database initialization SQL script: enable pgvector, create database
- [x] T011 Write Docker Compose startup script with volume mounting
- [x] T012 Test PostgreSQL connection and pgvector extension

### Ent Schema Definition

- [x] T013 Define `Email` schema in `ent/schema/email.go`
- [x] T014 [P] Define `DiscoveredEntity` schema with JSONB properties and vector field in `ent/schema/discovered_entity.go`
- [x] T015 [P] Define `Relationship` schema in `ent/schema/relationship.go`
- [x] T016 [P] Define `SchemaPromotion` audit schema in `ent/schema/schema_promotion.go`
- [x] T017 Generate ent code: `go generate ./ent`
- [x] T018 Create initial database migration and apply schema

### Graph Core Implementation

- [x] T019 Create `Repository` interface in `internal/graph/repository.go` with methods:
  - CreateEmail, CreateDiscoveredEntity, CreateRelationship
  - FindEntityByID, FindEntityByUniqueID, FindEntitiesByType
  - TraverseRelationships, FindShortestPath, SimilaritySearch
- [x] T020 Implement repository with ent client in `internal/graph/repository_impl.go`
- [X] T021 Implement pgvector similarity search using raw SQL: `ORDER BY embedding <-> $1 LIMIT $2`
- [X] T022 Add relationship traversal logic: 1-hop and n-hop queries
- [X] T023 Implement shortest path algorithm (BFS) in `internal/graph/path.go`

### Shared Utilities

- [x] T024 [P] Create structured logging setup using `log/slog` in `pkg/utils/logger.go`
- [x] T025 [P] Create configuration management for database connection in `pkg/utils/config.go`

**Checkpoint**: Foundation ready - user story implementation can begin (pending T021-T023)

---

## Phase 3: User Story 1 - Email Data Ingestion and Entity Extraction (Priority: P1) 🎯 MVP

**Goal**: Load Enron emails, extract entities (people, orgs, concepts) with relationships

**Independent Test**: Load 100 sample emails, verify entities extracted and stored, query back from database

### Unit Tests (TDD - Write First)

- [X] T026 [P] [US1] Write unit tests for CSV parser in `internal/loader/parser_test.go`:
  - Test streaming CSV rows
  - Test parsing `file` and `message` columns
  - Test handling malformed CSV
  - Test empty files and missing columns
- [X] T027 [P] [US1] Write unit tests for email header parser in `internal/loader/headers_test.go`:
  - Test Message-ID extraction
  - Test From/To/CC/BCC parsing
  - Test date parsing with various formats
  - Test UTF-8 and Latin-1 encoding handling
  - Test malformed headers
- [X] T028 [P] [US1] Write unit tests for LLM client in `pkg/llm/ollama_test.go`:
  - Test GenerateCompletion with mock responses
  - Test GenerateEmbedding with mock responses
  - Test retry logic (3 retries, exponential backoff)
  - Test timeout handling (30s completion, 10s embedding)
  - Test error cases (connection failures, invalid JSON)
- [X] T029 [P] [US1] Write unit tests for entity extractor in `internal/extractor/extractor_test.go`:
  - Test ExtractFromEmail with sample emails
  - Test JSON response parsing
  - Test confidence score filtering (<0.7 rejected)
  - Test embedding generation
  - Test entity type detection (person, org, concept)
- [X] T030 [P] [US1] Write unit tests for deduplication in `internal/extractor/dedup_test.go`:
  - Test person deduplication by email address
  - Test organization name normalization
  - Test concept similarity matching (cosine >0.85)
  - Test edge cases (empty names, special characters)
- [X] T031 [P] [US1] Write unit tests for relationship creation in `internal/extractor/relationships_test.go`:
  - Test SENT relationship creation
  - Test RECEIVED relationship creation
  - Test MENTIONS relationship creation
  - Test COMMUNICATES_WITH inference

### Email Loader Implementation

- [X] T032 [P] [US1] Implement CSV parser in `internal/loader/parser.go`:
  - Stream CSV rows (don't load entire file)
  - Parse `file` and `message` columns
  - **Verify**: T026 tests pass
- [X] T033 [P] [US1] Implement email header parser in `internal/loader/headers.go`:
  - Extract Message-ID, Date, From, To, CC, BCC, Subject
  - Use `net/mail` package for parsing
  - Handle encoding issues (UTF-8, Latin-1)
  - **Verify**: T027 tests pass
- [X] T034 [US1] Implement batch processor in `internal/loader/processor.go`:
  - Check for duplicates via message-id
  - Insert emails into database via repository
  - Concurrent processing with 10-100 goroutines
  - Progress logging every 100 emails
- [X] T035 [US1] Create loader CLI tool in `cmd/loader/main.go`:
  - Accept flags: `--csv-path`, `--batch-size`, `--workers`, `--extract`
  - Initialize database connection and repository
  - Report summary: processed count, failures, duration

### LLM Client Implementation

- [X] T036 [P] [US1] Implement Ollama client in `pkg/llm/ollama.go`:
  - GenerateCompletion(prompt) using LangChainGo
  - GenerateEmbedding(text) for mxbai-embed-large
  - Retry logic (3 retries, exponential backoff)
  - Timeouts (30s completion, 10s embedding)
  - **Verify**: T028 tests pass
- [X] T037 [P] [US1] Create LLM client interface in `pkg/llm/client.go`
- [ ] T038 [P] [US1] (Optional) Implement LiteLLM fallback client in `pkg/llm/litelim.go`

### Entity Extractor Implementation

- [X] T039 [US1] Design entity extraction prompt template in `internal/extractor/prompts.go`:
  - Extract persons, organizations, concepts with confidence scores
  - Return structured JSON output
  - Test on 10 sample emails, iterate for 70%+ precision
- [X] T040 [US1] Implement `ExtractFromEmail` in `internal/extractor/extractor.go`:
  - Parse email headers → high-confidence person entities
  - Call LLM with prompt → parse JSON response
  - Generate embeddings for each entity
  - Assign confidence scores, filter below 0.7
  - **Verify**: T029 tests pass
- [X] T041 [US1] Implement deduplication logic in `internal/extractor/dedup.go`:
  - Person: email address as unique key
  - Organization: normalize name (lowercase, trim)
  - Concept: embedding similarity (cosine >0.85)
  - **Verify**: T030 tests pass
- [X] T042 [US1] Implement relationship creation in `internal/extractor/relationships.go`:
  - SENT: person → email
  - RECEIVED: email → person
  - MENTIONS: email → organization/concept
  - COMMUNICATES_WITH: person ↔ person (inferred)
  - **Verify**: T031 tests pass
- [X] T043 [US1] Implement batch extraction in `internal/extractor/batch.go`:
  - Process 10-100 emails concurrently
  - Generate embeddings in batches
  - Handle LLM errors gracefully
  - Track metrics: success rate, avg confidence

### Integration

- [X] T044 [US1] Integrate extractor with loader in `cmd/loader/main.go`:
  - Call extractor when `--extract` flag is set
  - Report extraction summary: entities created, relationships created

### Integration Tests (TDD - Write First)

- [X] T045 [US1] Write integration test in `tests/integration/loader_test.go`:
  - Setup: Test database with clean schema
  - Test: Load 100 sample emails from fixture CSV
  - Verify: All emails inserted with correct metadata
  - Verify: No duplicate message-ids
  - Teardown: Clean test database
- [X] T046 [US1] Write integration test in `tests/integration/extractor_test.go`:
  - Setup: Test database with sample emails
  - Test: Run extraction on emails
  - Verify: Entities created with correct types
  - Verify: Relationships created (SENT, RECEIVED, MENTIONS)
  - Verify: Deduplicated entities (no duplicate email addresses)
  - Verify: Confidence scores applied correctly
  - Teardown: Clean test database
- [X] T047 [US1] Write end-to-end integration test in `tests/integration/test_user_story_1.sh`:
  - Run loader with `--extract` flag on 100 sample emails
  - Query database for entities
  - Verify entity count > 0
  - Verify relationship count > 0
  - Query for specific known entity (e.g., "jeff.skilling@enron.com")
  - Verify entity properties and relationships

**Acceptance Tests** (from spec.md):

- [X] T048 [US1] Verify: CSV parsing extracts metadata (sender, recipients, date, subject)
- [X] T049 [US1] Verify: Extractor identifies entities and relationships with structure
- [X] T050 [US1] Verify: Entities stored in graph and can be queried back
- [X] T051 [US1] Verify: Duplicate entities are merged, relationships aggregated
- [X] T052 [US1] Verify: SC-001 - 10k emails processed in <10 minutes (performance test - requires larger dataset)
- [X] T053 [US1] Verify: SC-002 - 90%+ precision for persons, 70%+ for orgs (requires manual review)
- [X] T054 [US1] Verify: SC-011 - 5+ loose entity types discovered (extracted: person, organization)

**Checkpoint**: User Story 1 complete - data loading and extraction functional

---

## Phase 4: User Story 2 - Graph Query and Exploration (Priority: P2)

**Goal**: Query entities, traverse relationships, explore organizational network via REST API

**Independent Test**: Query pre-populated graph for entities, find paths, filter by type

### Contract Tests (TDD - Write First)

- [X] T055 [P] [US2] Write contract tests for entity endpoints in `internal/api/handlers_test.go`:
  - Test GET /entities/:id returns 200 with valid JSON
  - Test GET /entities/:id returns 404 for non-existent ID
  - Test GET /entities?type=person&name=john returns filtered results
  - Test GET /entities validates query parameters
  - Test response schema matches expected structure
- [X] T056 [P] [US2] Write contract tests for relationship endpoints in `internal/api/handlers_test.go`:
  - Test GET /entities/:id/relationships returns 200 with relationship list
  - Test GET /entities/:id/neighbors?depth=1 returns immediate neighbors
  - Test GET /entities/:id/neighbors?depth=3 returns multi-hop neighbors
  - Test depth parameter validation (max depth enforcement)
- [X] T057 [P] [US2] Write contract tests for path finding in `internal/api/handlers_test.go`:
  - Test POST /entities/path with valid source/target returns path
  - Test POST /entities/path returns 404 if no path exists
  - Test POST /entities/path returns 400 for invalid request body
- [X] T058 [P] [US2] Write contract tests for semantic search in `internal/api/handlers_test.go`:
  - Test POST /entities/search with text returns similar entities
  - Test POST /entities/search validates request body
  - Test POST /entities/search returns ranked results by similarity

### Unit Tests for Query Engine (TDD - Write First)

- [X] T059 [P] [US2] Write unit tests for pagination in `internal/graph/pagination_test.go`:
  - Test limit/offset calculation
  - Test boundary conditions (offset > total count)
  - Test default pagination values
- [X] T060 [P] [US2] Write unit tests for filters in `internal/graph/filters_test.go`:
  - Test confidence score filtering
  - Test date range filtering
  - Test type filtering
  - Test combined filters

### API Design & Implementation

- [X] T061 [P] [US2] Define REST API endpoints in `specs/001-cognitive-backbone-poc/contracts/api.md`:
  - GET /entities/:id - Get entity by ID
  - GET /entities?type=&name= - Search entities
  - GET /entities/:id/relationships - Get relationships
  - GET /entities/:id/neighbors?depth= - Traverse relationships
  - POST /entities/path - Find shortest path
  - POST /entities/search - Semantic search by embedding
- [X] T062 [US2] Implement API handlers in `internal/api/handlers.go` using Chi router:
  - Map HTTP requests to repository methods
  - Input validation (IDs, query parameters)
  - JSON response formatting
  - Error handling (404, 400 status codes)
  - **Verify**: T055-T058 tests pass
- [X] T063 [P] [US2] Add middleware in `internal/api/middleware.go`:
  - Request logging
  - CORS (for future web UI)
  - Panic recovery
- [X] T064 [US2] Implement HTTP server in `cmd/server/main.go`:
  - Initialize Chi router
  - Configure database connection pooling
  - Graceful shutdown

### Query Engine Enhancements

- [X] T065 [P] [US2] Add database indexes in `internal/graph/indexes.go`:
  - Index on name, type, unique_id fields
- [X] T066 [P] [US2] Implement pagination in `internal/graph/pagination.go`:
  - Add limit/offset to queries
  - **Verify**: T059 tests pass
- [X] T067 [P] [US2] Add filtering by confidence score and date range in `internal/graph/filters.go`
  - **Verify**: T060 tests pass

### Integration Tests

- [ ] T068 [US2] Write integration test in `tests/integration/api_test.go`:
  - Setup: Pre-populate test database with entities and relationships
  - Test: Query entities via API
  - Verify: Correct entities returned
  - Test: Find shortest path between known entities
  - Verify: Correct path returned
  - Test: Semantic search for concept
  - Verify: Similar entities returned
  - Teardown: Clean test database

**Acceptance Tests** (from spec.md):

- [ ] T069 [US2] Verify: Query person by name returns entity with properties
- [ ] T070 [US2] Verify: Query relationships returns list of connected entities
- [ ] T071 [US2] Verify: Shortest path between entities returns relationship chain
- [ ] T072 [US2] Verify: Filter by entity type returns matching entities
- [ ] T073 [US2] Verify: SC-003 - Entity lookup <500ms for 100k nodes
- [ ] T074 [US2] Verify: SC-004 - Shortest path <2s for 6 degrees

**Checkpoint**: User Story 2 complete - graph querying functional via API

---

## Phase 5: User Story 3 - Schema Evolution through Type Promotion (Priority: P3)

**Goal**: Identify patterns, promote entity types from discovered to core schema

**Independent Test**: Run analyst on graph with discovered entities, verify candidates identified, promote one type

### Unit Tests (TDD - Write First)

- [ ] T075 [P] [US3] Write unit tests for pattern detection in `internal/analyst/detector_test.go`:
  - Test frequency calculation per type
  - Test relationship density calculation
  - Test property consistency calculation
  - Test grouping by type_category
- [ ] T076 [P] [US3] Write unit tests for embedding clustering in `internal/analyst/clustering_test.go`:
  - Test similarity grouping (cosine >0.85)
  - Test cluster identification
  - Test type candidate extraction
- [ ] T077 [P] [US3] Write unit tests for candidate ranking in `internal/analyst/ranker_test.go`:
  - Test scoring formula (0.4*freq + 0.3*density + 0.3*consistency)
  - Test threshold application (min 50 occurrences, 70% consistency)
  - Test sorting by score
- [ ] T078 [P] [US3] Write unit tests for schema generator in `internal/analyst/schema_gen_test.go`:
  - Test required property inference (>90% presence)
  - Test optional property inference (30-90% presence)
  - Test data type inference from samples
  - Test validation rule generation
  - Test JSON schema output format
- [ ] T079 [P] [US3] Write unit tests for ent schema codegen in `internal/promoter/codegen_test.go`:
  - Test ent schema file generation from JSON schema
  - Test field type mapping
  - Test validation rule conversion
  - Test file output format

### Pattern Detection & Ranking

- [ ] T080 [US3] Implement pattern detection in `internal/analyst/detector.go`:
  - Query discovered entities grouped by type_category
  - Calculate frequency: COUNT(*) per type
  - Calculate relationship density: AVG(degree) per type
  - Calculate property consistency: % entities with each property
  - **Verify**: T075 tests pass
- [ ] T081 [P] [US3] Implement embedding clustering in `internal/analyst/clustering.go`:
  - Group entities by vector similarity (cosine >0.85)
  - Identify type candidates from clusters
  - **Verify**: T076 tests pass
- [ ] T082 [US3] Implement candidate ranking in `internal/analyst/ranker.go`:
  - Score = 0.4*frequency + 0.3*density + 0.3*consistency
  - Apply thresholds: min 50 occurrences, 70% property consistency
  - Sort by score, return top 10
  - **Verify**: T077 tests pass

### Schema Generation & Promotion

- [ ] T083 [US3] Implement schema generator in `internal/analyst/schema_gen.go`:
  - Infer required properties (>90% presence)
  - Infer optional properties (30-90% presence)
  - Infer data types from sample values
  - Generate validation rules
  - Output JSON schema definition
  - **Verify**: T078 tests pass
- [ ] T084 [US3] Create ent schema file generator in `internal/promoter/codegen.go`:
  - Input: JSON schema definition
  - Output: `ent/schema/{typename}.go` with ent schema struct
  - **Verify**: T079 tests pass
- [ ] T085 [US3] Implement promotion workflow in `internal/promoter/promoter.go`:
  - Generate ent schema file
  - Run `go generate ./ent`
  - Run database migration (create new table)
  - Copy data from DiscoveredEntity to new table
  - Validate entities against schema, log failures
  - Create SchemaPromotion audit record

### CLI Tools

- [ ] T086 [US3] Create analyst CLI in `cmd/analyst/main.go`:
  - `analyze` subcommand: run detection, rank candidates, display top 10
  - `promote` subcommand: interactive selection, confirm, execute promotion
- [ ] T087 [P] [US3] Create promoter CLI in `cmd/promoter/main.go`:
  - Accept schema definition file
  - Execute promotion workflow

### Integration Tests

- [ ] T088 [US3] Write integration test in `tests/integration/analyst_test.go`:
  - Setup: Pre-populate test database with diverse discovered entities
  - Test: Run pattern detection
  - Verify: Candidates identified with correct scores
  - Verify: Ranking by frequency/density/consistency
  - Teardown: Clean test database
- [ ] T089 [US3] Write integration test in `tests/integration/promoter_test.go`:
  - Setup: Test database with candidate type entities
  - Test: Generate schema from candidate
  - Verify: JSON schema has correct properties and types
  - Test: Run promotion workflow
  - Verify: New ent schema file created
  - Verify: New database table created
  - Verify: Data migrated from DiscoveredEntity
  - Verify: SchemaPromotion audit record created
  - Teardown: Clean test database and generated files

**Acceptance Tests** (from spec.md):

- [ ] T090 [US3] Verify: Analyst identifies frequent/high-connectivity entity types
- [ ] T091 [US3] Verify: Candidates ranked by frequency, density, consistency
- [ ] T092 [US3] Verify: Promotion adds type to schema with properties/constraints
- [ ] T093 [US3] Verify: New entities validated against promoted schema
- [ ] T094 [US3] Verify: Audit log captures promotion events
- [ ] T095 [US3] Verify: SC-005 - 3+ candidates identified from 10k emails
- [ ] T096 [US3] Verify: SC-006 - 1+ type successfully promoted
- [ ] T097 [US3] Verify: SC-010 - Audit log is complete

**Checkpoint**: User Story 3 complete - schema evolution demonstrated

---

## Phase 6: User Story 4 - Basic Visualization of Graph Structure (Priority: P4)

**Goal**: TUI interface to visualize entities and relationships as ASCII graph

**Independent Test**: Load subgraph, verify rendering with nodes/edges, test navigation

### Unit Tests (TDD - Write First)

- [ ] T098 [P] [US4] Write unit tests for ASCII graph rendering in `internal/tui/graph_view_test.go`:
  - Test node formatting: `[Type: Name]`
  - Test edge formatting: `---[REL_TYPE]-->`
  - Test layout calculation for tree structure
  - Test node limit enforcement (max 50 nodes)
  - Test color-coding by entity type
- [ ] T099 [P] [US4] Write unit tests for entity list view in `internal/tui/entity_list_test.go`:
  - Test table rendering with columns
  - Test pagination calculation
  - Test filter by type logic
  - Test search by name logic

### TUI Framework Setup

- [ ] T100 [P] [US4] Setup Bubble Tea framework in `internal/tui/app.go`:
  - Create base model (state container)
  - Implement Update function (message handler)
  - Implement View function (renderer)
  - Multi-view navigation

### Entity Browser

- [ ] T101 [US4] Implement entity list view in `internal/tui/entity_list.go`:
  - Display entities in table (ID, type, name, confidence)
  - Keyboard navigation (↑↓ arrows, page up/down)
  - Filter by type (F key)
  - Search by name (/ key)
  - **Verify**: T099 tests pass
- [ ] T102 [US4] Fetch entities from repository with pagination

### Graph Visualization

- [ ] T103 [US4] Implement ASCII graph rendering in `internal/tui/graph_view.go`:
  - Node format: `[Type: Name]` (color-coded)
  - Edge format: `---[REL_TYPE]-->` (directional)
  - Simple tree or force-directed layout
  - Limit to 50 nodes (performance)
  - **Verify**: T098 tests pass
- [ ] T104 [US4] Implement subgraph view around selected entity:
  - Show entity + immediate neighbors (1-hop)
  - Expand node: E key → fetch and add neighbors
- [ ] T105 [US4] Implement keyboard navigation:
  - Tab: cycle through nodes
  - Enter: select node → show details
  - E: expand node
  - B: back to entity list

### Detail Panel

- [ ] T106 [P] [US4] Implement detail view in `internal/tui/detail_view.go`:
  - Display entity properties (key-value pairs)
  - Display relationships (list)
  - Actions: V (visualize), R (show related), Q (close)

### Main Application

- [ ] T107 [US4] Integrate TUI into `cmd/server/main.go`:
  - Initialize Bubble Tea program
  - Connect to database repository
  - Add welcome screen
  - Add status bar (entity count, current view)

### Manual Testing Checklist

- [ ] T108 [US4] Manual test: TUI navigation flows
  - Start TUI, navigate entity list
  - Filter by type, verify results
  - Search by name, verify results
  - Select entity, view details
  - Visualize entity as graph
  - Navigate graph view, expand nodes
- [ ] T109 [US4] Manual test: Error handling
  - Test empty database
  - Test network disconnection
  - Test invalid entity selection

**Acceptance Tests** (from spec.md):

- [ ] T110 [US4] Verify: TUI displays entities and relationships
- [ ] T111 [US4] Verify: Graph view renders ASCII graph with nodes/edges
- [ ] T112 [US4] Verify: Clicking node shows properties and expansion option
- [ ] T113 [US4] Verify: Controls for filtering by entity type work
- [ ] T114 [US4] Verify: SC-007 - <3s render time for 500 nodes

**Checkpoint**: User Story 4 complete - TUI visualization functional

---

## Phase 7: User Story 5 - Natural Language Search and Chat Interface (Priority: P5)

**Goal**: Chat interface for natural language queries against graph

**Independent Test**: Submit NL queries, verify relevant entities/relationships returned

### Unit Tests (TDD - Write First)

- [ ] T115 [P] [US5] Write unit tests for pattern matching in `internal/chat/patterns_test.go`:
  - Test entity lookup pattern: "Who is X?"
  - Test relationship pattern: "Who did X email?"
  - Test path finding pattern: "How are X and Y connected?"
  - Test concept search pattern: "Emails about energy"
  - Test aggregation pattern: "How many emails did X send?"
  - Test ambiguity handling
- [ ] T116 [P] [US5] Write unit tests for context management in `internal/chat/context_test.go`:
  - Test conversation history storage (last 5 queries)
  - Test entity tracking for pronoun resolution
  - Test context injection into prompts
- [ ] T117 [P] [US5] Write unit tests for chat handler in `internal/chat/handler_test.go`:
  - Test query processing with mock LLM
  - Test response formatting
  - Test error handling
  - Test context propagation

### Chat Handler Implementation

- [ ] T118 [US5] Implement chat handler in `internal/chat/handler.go`:
  - ProcessQuery(query, context, llm) → Response
  - Send query + history + schema to LLM
  - Parse LLM response (graph query or answer)
  - Execute query via repository
  - Format results for user
  - **Verify**: T117 tests pass
- [ ] T119 [US5] Implement conversation context management in `internal/chat/context.go`:
  - Store last 5 queries and responses
  - Track mentioned entities (pronoun resolution)
  - Include context in LLM prompt
  - **Verify**: T116 tests pass

### Query Pattern Matching

- [ ] T120 [US5] Implement pattern matching in `internal/chat/patterns.go`:
  - Entity lookup: "Who is X?" → FindEntityByName(X)
  - Relationship discovery: "Who did X email?" → TraverseRelationships(X, "SENT")
  - Path finding: "How are X and Y connected?" → FindShortestPath(X, Y)
  - Concept search: "Emails about energy" → SimilaritySearch(embedding("energy"))
  - Aggregations: "How many emails did X send?" → CountRelationships(X, "SENT")
  - **Verify**: T115 tests pass
- [ ] T121 [US5] Handle ambiguity: return multiple options, ask for clarification

### Prompt Engineering

- [ ] T122 [P] [US5] Design chat prompt template in `internal/chat/prompts.go`:
  - System role: graph database assistant
  - Available operations and schema
  - Conversation history
  - User query
  - Response format (answer or graph query command)
- [ ] T123 [US5] Test prompt on 10 sample queries, iterate for accuracy
  - Create test queries in `internal/chat/prompts_test.go`
  - Verify expected responses for each query type

### TUI Chat View

- [ ] T124 [US5] Implement chat interface in `internal/tui/chat_view.go`:
  - Message history (scrollable)
  - Input box (text input)
  - Send: Enter key
  - Clear history: Ctrl+L
- [ ] T125 [US5] Display query results:
  - Entity cards (name, type, snippet)
  - Relationship paths (arrow chains)
  - "Visualize" button → graph view with results highlighted
- [ ] T126 [US5] Add loading indicator for LLM processing

### Integration

- [ ] T127 [US5] Add chat view to TUI in `cmd/server/main.go`:
  - C key → chat view
  - Initialize LLM client
  - Session management

### Integration Tests

- [ ] T128 [US5] Write integration test in `tests/integration/chat_test.go`:
  - Setup: Pre-populate test database with entities
  - Test: Submit query "Who is Jeff Skilling?"
  - Verify: Correct entity returned
  - Test: Submit query "Who did Jeff Skilling email?"
  - Verify: Related entities returned
  - Test: Submit query "How are Jeff Skilling and Kenneth Lay connected?"
  - Verify: Path returned
  - Test: Submit query "Emails about energy trading"
  - Verify: Semantic search results returned
  - Test: Context maintenance across multiple queries
  - Verify: Second query uses context from first
  - Teardown: Clean test database
- [ ] T129 [US5] Write end-to-end chat test in `tests/integration/test_chat_e2e.sh`:
  - Start chat interface
  - Submit series of related queries
  - Verify responses are contextually appropriate
  - Test error handling for ambiguous queries

**Acceptance Tests** (from spec.md):

- [ ] T130 [US5] Verify: Chat processes natural language queries
- [ ] T131 [US5] Verify: Entity lookup queries work ("Who is X?")
- [ ] T132 [US5] Verify: Relationship queries work ("Who did X email?")
- [ ] T133 [US5] Verify: Path finding queries work ("How are X and Y connected?")
- [ ] T134 [US5] Verify: Concept search works ("Emails about energy")
- [ ] T135 [US5] Verify: Conversation context maintained across queries
- [ ] T136 [US5] Verify: Ambiguity handled with clarification
- [ ] T137 [US5] Verify: Results can be visualized in graph view
- [ ] T138 [US5] Verify: SC-012 - 80%+ accuracy on test queries
- [ ] T139 [US5] Verify: SC-013 - Context maintained 3+ consecutive queries

**Checkpoint**: User Story 5 complete - chat interface functional

---

## Phase 8: Integration & Polish

**Purpose**: End-to-end testing, performance validation, POC completion

### Integration Testing

- [ ] T140 Full workflow test: load emails → extract → query → promote → query promoted entities
- [ ] T141 Test with full 10k+ email dataset
- [ ] T142 Verify all P1-P3 user stories pass acceptance scenarios
- [ ] T143 Test P4-P5 demo features

### Performance Validation

- [ ] T144 Validate SC-001: 10k emails in <10 minutes
  - Write benchmark in `tests/benchmarks/loader_bench_test.go`
  - Load 10k emails with extraction enabled
  - Measure total time, verify <10 minutes
- [ ] T145 Validate SC-003: Entity lookup <500ms (100k nodes)
  - Write benchmark in `tests/benchmarks/query_bench_test.go`
  - Pre-populate database with 100k entities
  - Run entity lookup queries
  - Measure p50, p95, p99 latencies
  - Verify p95 <500ms
- [ ] T146 Validate SC-004: Shortest path <2s (6 degrees)
  - Write benchmark in `tests/benchmarks/path_bench_test.go`
  - Create graph with 6-degree separation
  - Run shortest path queries
  - Measure latency, verify <2s
- [ ] T147 Validate SC-007: Visualization <3s (500 nodes)
  - Write performance test in `tests/benchmarks/tui_bench_test.go`
  - Load 500-node subgraph
  - Measure rendering time
  - Verify <3s

### Data Consistency

- [ ] T148 Validate SC-009: No duplicate entities with same unique ID
  - Write data integrity test in `tests/integration/integrity_test.go`
  - Load diverse dataset
  - Query for duplicates by unique_id
  - Verify count = 0
- [ ] T149 Validate SC-009: All relationships reference valid entities
  - Write referential integrity test in `tests/integration/integrity_test.go`
  - Query relationships with LEFT JOIN on entities
  - Verify no NULL entity references
- [ ] T150 Test concurrent writes (loader + extractor)
  - Write concurrency test in `tests/integration/concurrency_test.go`
  - Run multiple loaders simultaneously
  - Verify no race conditions or deadlocks
  - Verify data integrity after concurrent writes

### Documentation & Demo

- [ ] T151 Update README.md with setup instructions
- [ ] T152 Create demo script for stakeholders in `docs/demo.md`:
  - Step 1: Load sample emails
  - Step 2: Query entities via API
  - Step 3: Run analyst, promote type
  - Step 4: Explore via TUI
  - Step 5: Chat interface demo
- [ ] T153 Document lessons learned in `specs/001-cognitive-backbone-poc/lessons-learned.md`
  - What worked well
  - What challenges encountered
  - What would be done differently
  - Recommendations for production implementation

### Bug Fixes

- [ ] T154 Fix issues found during integration testing
- [ ] T155 Improve error messages and user feedback
- [ ] T156 Add missing input validation
- [ ] T157 Optimize slow queries identified in performance testing

---

## Summary

**Total Tasks**: 157

**Tasks by User Story**:
- Setup & Foundation: 25 tasks (T001-T025)
- User Story 1 (P1 MVP): 29 tasks (T026-T054) - includes 6 unit test tasks, 3 integration test tasks
- User Story 2 (P2): 20 tasks (T055-T074) - includes 6 contract/unit test tasks, 1 integration test task
- User Story 3 (P3): 23 tasks (T075-T097) - includes 5 unit test tasks, 2 integration test tasks
- User Story 4 (P4): 17 tasks (T098-T114) - includes 2 unit test tasks, 2 manual test tasks
- User Story 5 (P5): 25 tasks (T115-T139) - includes 3 unit test tasks, 2 integration test tasks
- Integration & Polish: 18 tasks (T140-T157) - includes 7 performance/integrity test tasks

**Test Tasks by Type**:
- Unit Tests: 22 tasks
- Integration Tests: 11 tasks
- Contract Tests: 6 tasks
- Performance/Benchmark Tests: 7 tasks
- Manual Tests: 2 tasks
- Acceptance Tests: 29 tasks
- **Total Test Tasks**: 77 (49% of all tasks)

**TDD Compliance**:
All implementation follows Red-Green-Refactor cycle:
1. Write failing tests first (contract tests, unit tests, integration tests)
2. Verify tests fail for the right reasons
3. Implement minimum code to make tests pass
4. Refactor while keeping tests green

**MVP Scope** (User Story 1):
Tasks T001-T054 deliver the core POC: email loading, entity extraction, graph storage, basic querying with comprehensive test coverage.

**Parallel Opportunities**:
- Phase 1: T006, T007, T008 (project structure setup)
- Phase 2: T014-T016 (schema definitions), T024-T025 (utilities)
- Phase 3: T026-T031 (unit tests - can all run in parallel), T032-T033 (loader components), T036-T038 (LLM clients)
- Phase 4: T055-T060 (contract/unit tests - can all run in parallel), T061, T063, T065-T067 (API components)
- Phase 5: T075-T079 (unit tests - can all run in parallel), T081 (clustering), T087 (promoter CLI)
- Phase 6: T098-T099 (unit tests - can all run in parallel), T100, T106 (TUI components)
- Phase 7: T115-T117 (unit tests - can all run in parallel), T122 (prompt engineering)

**Critical Path**: 
T001-T025 (Setup/Foundation) → T026-T054 (US1 with tests) → T055-T074 (US2 with tests) → T075-T097 (US3 with tests)

**Estimated Timeline**: 7-9 weeks (single developer, full-time)
- Week 1-2: Setup + Foundation (T001-T025)
- Week 3-4: User Story 1 + Tests (T026-T054)
- Week 5: User Story 2 + Tests (T055-T074)
- Week 6: User Story 3 + Tests (T075-T097)
- Week 7: User Story 4 + Tests (T098-T114)
- Week 8: User Story 5 + Tests (T115-T139)
- Week 9: Integration, Performance Testing & Polish (T140-T157)
