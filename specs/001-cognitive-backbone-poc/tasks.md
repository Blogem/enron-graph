# Tasks: Cognitive Backbone POC

**Feature**: 001-cognitive-backbone-poc  
**Generated**: 2026-01-24  
**Input**: spec.md, plan.md, analysis.md, clarifications.md  
**Tech Stack**: Go 1.21+, PostgreSQL + ent + pgvector, Ollama, Bubble Tea

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
- [ ] T021 Implement pgvector similarity search using raw SQL: `ORDER BY embedding <-> $1 LIMIT $2`
- [ ] T022 Add relationship traversal logic: 1-hop and n-hop queries
- [ ] T023 Implement shortest path algorithm (BFS) in `internal/graph/path.go`

### Shared Utilities

- [x] T024 [P] Create structured logging setup using `log/slog` in `pkg/utils/logger.go`
- [x] T025 [P] Create configuration management for database connection in `pkg/utils/config.go`

**Checkpoint**: Foundation ready - user story implementation can begin (pending T021-T023)

---

## Phase 3: User Story 1 - Email Data Ingestion and Entity Extraction (Priority: P1) 🎯 MVP

**Goal**: Load Enron emails, extract entities (people, orgs, concepts) with relationships

**Independent Test**: Load 100 sample emails, verify entities extracted and stored, query back from database

### Email Loader Implementation

- [ ] T026 [P] [US1] Implement CSV parser in `internal/loader/parser.go`:
  - Stream CSV rows (don't load entire file)
  - Parse `file` and `message` columns
- [ ] T027 [P] [US1] Implement email header parser in `internal/loader/headers.go`:
  - Extract Message-ID, Date, From, To, CC, BCC, Subject
  - Use `net/mail` package for parsing
  - Handle encoding issues (UTF-8, Latin-1)
- [ ] T028 [US1] Implement batch processor in `internal/loader/processor.go`:
  - Check for duplicates via message-id
  - Insert emails into database via repository
  - Concurrent processing with 10-100 goroutines
  - Progress logging every 100 emails
- [ ] T029 [US1] Create loader CLI tool in `cmd/loader/main.go`:
  - Accept flags: `--csv-path`, `--batch-size`, `--workers`, `--extract`
  - Initialize database connection and repository
  - Report summary: processed count, failures, duration

### LLM Client Implementation

- [ ] T030 [P] [US1] Implement Ollama client in `pkg/llm/ollama.go`:
  - GenerateCompletion(prompt) using LangChainGo
  - GenerateEmbedding(text) for mxbai-embed-large
  - Retry logic (3 retries, exponential backoff)
  - Timeouts (30s completion, 10s embedding)
- [ ] T031 [P] [US1] Create LLM client interface in `pkg/llm/client.go`
- [ ] T032 [P] [US1] (Optional) Implement LiteLLM fallback client in `pkg/llm/litelim.go`

### Entity Extractor Implementation

- [ ] T033 [US1] Design entity extraction prompt template in `internal/extractor/prompts.go`:
  - Extract persons, organizations, concepts with confidence scores
  - Return structured JSON output
  - Test on 10 sample emails, iterate for 70%+ precision
- [ ] T034 [US1] Implement `ExtractFromEmail` in `internal/extractor/extractor.go`:
  - Parse email headers → high-confidence person entities
  - Call LLM with prompt → parse JSON response
  - Generate embeddings for each entity
  - Assign confidence scores, filter below 0.7
- [ ] T035 [US1] Implement deduplication logic in `internal/extractor/dedup.go`:
  - Person: email address as unique key
  - Organization: normalize name (lowercase, trim)
  - Concept: embedding similarity (cosine >0.85)
- [ ] T036 [US1] Implement relationship creation in `internal/extractor/relationships.go`:
  - SENT: person → email
  - RECEIVED: email → person
  - MENTIONS: email → organization/concept
  - COMMUNICATES_WITH: person ↔ person (inferred)
- [ ] T037 [US1] Implement batch extraction in `internal/extractor/batch.go`:
  - Process 10-100 emails concurrently
  - Generate embeddings in batches
  - Handle LLM errors gracefully
  - Track metrics: success rate, avg confidence

### Integration

- [ ] T038 [US1] Integrate extractor with loader in `cmd/loader/main.go`:
  - Call extractor when `--extract` flag is set
  - Report extraction summary: entities created, relationships created

**Acceptance Tests** (from spec.md):

- [ ] T039 [US1] Verify: CSV parsing extracts metadata (sender, recipients, date, subject)
- [ ] T040 [US1] Verify: Extractor identifies entities and relationships with structure
- [ ] T041 [US1] Verify: Entities stored in graph and can be queried back
- [ ] T042 [US1] Verify: Duplicate entities are merged, relationships aggregated
- [ ] T043 [US1] Verify: SC-001 - 10k emails processed in <10 minutes
- [ ] T044 [US1] Verify: SC-002 - 90%+ precision for persons, 70%+ for orgs
- [ ] T045 [US1] Verify: SC-011 - 5+ loose entity types discovered

**Checkpoint**: User Story 1 complete - data loading and extraction functional

---

## Phase 4: User Story 2 - Graph Query and Exploration (Priority: P2)

**Goal**: Query entities, traverse relationships, explore organizational network via REST API

**Independent Test**: Query pre-populated graph for entities, find paths, filter by type

### API Design & Implementation

- [ ] T046 [P] [US2] Define REST API endpoints in `specs/001-cognitive-backbone-poc/contracts/api.md`:
  - GET /entities/:id - Get entity by ID
  - GET /entities?type=&name= - Search entities
  - GET /entities/:id/relationships - Get relationships
  - GET /entities/:id/neighbors?depth= - Traverse relationships
  - POST /entities/path - Find shortest path
  - POST /entities/search - Semantic search by embedding
- [ ] T047 [US2] Implement API handlers in `internal/api/handlers.go` using Chi router:
  - Map HTTP requests to repository methods
  - Input validation (IDs, query parameters)
  - JSON response formatting
  - Error handling (404, 400 status codes)
- [ ] T048 [P] [US2] Add middleware in `internal/api/middleware.go`:
  - Request logging
  - CORS (for future web UI)
  - Panic recovery
- [ ] T049 [US2] Implement HTTP server in `cmd/server/main.go`:
  - Initialize Chi router
  - Configure database connection pooling
  - Graceful shutdown

### Query Engine Enhancements

- [ ] T050 [P] [US2] Add database indexes in `internal/graph/indexes.go`:
  - Index on name, type, unique_id fields
- [ ] T051 [P] [US2] Implement pagination in `internal/graph/pagination.go`:
  - Add limit/offset to queries
- [ ] T052 [P] [US2] Add filtering by confidence score and date range in `internal/graph/filters.go`

**Acceptance Tests** (from spec.md):

- [ ] T053 [US2] Verify: Query person by name returns entity with properties
- [ ] T054 [US2] Verify: Query relationships returns list of connected entities
- [ ] T055 [US2] Verify: Shortest path between entities returns relationship chain
- [ ] T056 [US2] Verify: Filter by entity type returns matching entities
- [ ] T057 [US2] Verify: SC-003 - Entity lookup <500ms for 100k nodes
- [ ] T058 [US2] Verify: SC-004 - Shortest path <2s for 6 degrees

**Checkpoint**: User Story 2 complete - graph querying functional via API

---

## Phase 5: User Story 3 - Schema Evolution through Type Promotion (Priority: P3)

**Goal**: Identify patterns, promote entity types from discovered to core schema

**Independent Test**: Run analyst on graph with discovered entities, verify candidates identified, promote one type

### Pattern Detection & Ranking

- [ ] T059 [US3] Implement pattern detection in `internal/analyst/detector.go`:
  - Query discovered entities grouped by type_category
  - Calculate frequency: COUNT(*) per type
  - Calculate relationship density: AVG(degree) per type
  - Calculate property consistency: % entities with each property
- [ ] T060 [P] [US3] Implement embedding clustering in `internal/analyst/clustering.go`:
  - Group entities by vector similarity (cosine >0.85)
  - Identify type candidates from clusters
- [ ] T061 [US3] Implement candidate ranking in `internal/analyst/ranker.go`:
  - Score = 0.4*frequency + 0.3*density + 0.3*consistency
  - Apply thresholds: min 50 occurrences, 70% property consistency
  - Sort by score, return top 10

### Schema Generation & Promotion

- [ ] T062 [US3] Implement schema generator in `internal/analyst/schema_gen.go`:
  - Infer required properties (>90% presence)
  - Infer optional properties (30-90% presence)
  - Infer data types from sample values
  - Generate validation rules
  - Output JSON schema definition
- [ ] T063 [US3] Create ent schema file generator in `internal/promoter/codegen.go`:
  - Input: JSON schema definition
  - Output: `ent/schema/{typename}.go` with ent schema struct
- [ ] T064 [US3] Implement promotion workflow in `internal/promoter/promoter.go`:
  - Generate ent schema file
  - Run `go generate ./ent`
  - Run database migration (create new table)
  - Copy data from DiscoveredEntity to new table
  - Validate entities against schema, log failures
  - Create SchemaPromotion audit record

### CLI Tools

- [ ] T065 [US3] Create analyst CLI in `cmd/analyst/main.go`:
  - `analyze` subcommand: run detection, rank candidates, display top 10
  - `promote` subcommand: interactive selection, confirm, execute promotion
- [ ] T066 [P] [US3] Create promoter CLI in `cmd/promoter/main.go`:
  - Accept schema definition file
  - Execute promotion workflow

**Acceptance Tests** (from spec.md):

- [ ] T067 [US3] Verify: Analyst identifies frequent/high-connectivity entity types
- [ ] T068 [US3] Verify: Candidates ranked by frequency, density, consistency
- [ ] T069 [US3] Verify: Promotion adds type to schema with properties/constraints
- [ ] T070 [US3] Verify: New entities validated against promoted schema
- [ ] T071 [US3] Verify: Audit log captures promotion events
- [ ] T072 [US3] Verify: SC-005 - 3+ candidates identified from 10k emails
- [ ] T073 [US3] Verify: SC-006 - 1+ type successfully promoted
- [ ] T074 [US3] Verify: SC-010 - Audit log is complete

**Checkpoint**: User Story 3 complete - schema evolution demonstrated

---

## Phase 6: User Story 4 - Basic Visualization of Graph Structure (Priority: P4)

**Goal**: TUI interface to visualize entities and relationships as ASCII graph

**Independent Test**: Load subgraph, verify rendering with nodes/edges, test navigation

### TUI Framework Setup

- [ ] T075 [P] [US4] Setup Bubble Tea framework in `internal/tui/app.go`:
  - Create base model (state container)
  - Implement Update function (message handler)
  - Implement View function (renderer)
  - Multi-view navigation

### Entity Browser

- [ ] T076 [US4] Implement entity list view in `internal/tui/entity_list.go`:
  - Display entities in table (ID, type, name, confidence)
  - Keyboard navigation (↑↓ arrows, page up/down)
  - Filter by type (F key)
  - Search by name (/ key)
- [ ] T077 [US4] Fetch entities from repository with pagination

### Graph Visualization

- [ ] T078 [US4] Implement ASCII graph rendering in `internal/tui/graph_view.go`:
  - Node format: `[Type: Name]` (color-coded)
  - Edge format: `---[REL_TYPE]-->` (directional)
  - Simple tree or force-directed layout
  - Limit to 50 nodes (performance)
- [ ] T079 [US4] Implement subgraph view around selected entity:
  - Show entity + immediate neighbors (1-hop)
  - Expand node: E key → fetch and add neighbors
- [ ] T080 [US4] Implement keyboard navigation:
  - Tab: cycle through nodes
  - Enter: select node → show details
  - E: expand node
  - B: back to entity list

### Detail Panel

- [ ] T081 [P] [US4] Implement detail view in `internal/tui/detail_view.go`:
  - Display entity properties (key-value pairs)
  - Display relationships (list)
  - Actions: V (visualize), R (show related), Q (close)

### Main Application

- [ ] T082 [US4] Integrate TUI into `cmd/server/main.go`:
  - Initialize Bubble Tea program
  - Connect to database repository
  - Add welcome screen
  - Add status bar (entity count, current view)

**Acceptance Tests** (from spec.md):

- [ ] T083 [US4] Verify: TUI displays entities and relationships
- [ ] T084 [US4] Verify: Graph view renders ASCII graph with nodes/edges
- [ ] T085 [US4] Verify: Clicking node shows properties and expansion option
- [ ] T086 [US4] Verify: Controls for filtering by entity type work
- [ ] T087 [US4] Verify: SC-007 - <3s render time for 500 nodes

**Checkpoint**: User Story 4 complete - TUI visualization functional

---

## Phase 7: User Story 5 - Natural Language Search and Chat Interface (Priority: P5)

**Goal**: Chat interface for natural language queries against graph

**Independent Test**: Submit NL queries, verify relevant entities/relationships returned

### Chat Handler Implementation

- [ ] T088 [US5] Implement chat handler in `internal/chat/handler.go`:
  - ProcessQuery(query, context, llm) → Response
  - Send query + history + schema to LLM
  - Parse LLM response (graph query or answer)
  - Execute query via repository
  - Format results for user
- [ ] T089 [US5] Implement conversation context management in `internal/chat/context.go`:
  - Store last 5 queries and responses
  - Track mentioned entities (pronoun resolution)
  - Include context in LLM prompt

### Query Pattern Matching

- [ ] T090 [US5] Implement pattern matching in `internal/chat/patterns.go`:
  - Entity lookup: "Who is X?" → FindEntityByName(X)
  - Relationship discovery: "Who did X email?" → TraverseRelationships(X, "SENT")
  - Path finding: "How are X and Y connected?" → FindShortestPath(X, Y)
  - Concept search: "Emails about energy" → SimilaritySearch(embedding("energy"))
  - Aggregations: "How many emails did X send?" → CountRelationships(X, "SENT")
- [ ] T091 [US5] Handle ambiguity: return multiple options, ask for clarification

### Prompt Engineering

- [ ] T092 [P] [US5] Design chat prompt template in `internal/chat/prompts.go`:
  - System role: graph database assistant
  - Available operations and schema
  - Conversation history
  - User query
  - Response format (answer or graph query command)
- [ ] T093 [US5] Test prompt on 10 sample queries, iterate for accuracy

### TUI Chat View

- [ ] T094 [US5] Implement chat interface in `internal/tui/chat_view.go`:
  - Message history (scrollable)
  - Input box (text input)
  - Send: Enter key
  - Clear history: Ctrl+L
- [ ] T095 [US5] Display query results:
  - Entity cards (name, type, snippet)
  - Relationship paths (arrow chains)
  - "Visualize" button → graph view with results highlighted
- [ ] T096 [US5] Add loading indicator for LLM processing

### Integration

- [ ] T097 [US5] Add chat view to TUI in `cmd/server/main.go`:
  - C key → chat view
  - Initialize LLM client
  - Session management

**Acceptance Tests** (from spec.md):

- [ ] T098 [US5] Verify: Chat processes natural language queries
- [ ] T099 [US5] Verify: Entity lookup queries work ("Who is X?")
- [ ] T100 [US5] Verify: Relationship queries work ("Who did X email?")
- [ ] T101 [US5] Verify: Path finding queries work ("How are X and Y connected?")
- [ ] T102 [US5] Verify: Concept search works ("Emails about energy")
- [ ] T103 [US5] Verify: Conversation context maintained across queries
- [ ] T104 [US5] Verify: Ambiguity handled with clarification
- [ ] T105 [US5] Verify: Results can be visualized in graph view
- [ ] T106 [US5] Verify: SC-012 - 80%+ accuracy on test queries
- [ ] T107 [US5] Verify: SC-013 - Context maintained 3+ consecutive queries

**Checkpoint**: User Story 5 complete - chat interface functional

---

## Phase 8: Integration & Polish

**Purpose**: End-to-end testing, performance validation, POC completion

### Integration Testing

- [ ] T108 Full workflow test: load emails → extract → query → promote → query promoted entities
- [ ] T109 Test with full 10k+ email dataset
- [ ] T110 Verify all P1-P3 user stories pass acceptance scenarios
- [ ] T111 Test P4-P5 demo features

### Performance Validation

- [ ] T112 Validate SC-001: 10k emails in <10 minutes
- [ ] T113 Validate SC-003: Entity lookup <500ms (100k nodes)
- [ ] T114 Validate SC-004: Shortest path <2s (6 degrees)
- [ ] T115 Validate SC-007: Visualization <3s (500 nodes)

### Data Consistency

- [ ] T116 Validate SC-009: No duplicate entities with same unique ID
- [ ] T117 Validate SC-009: All relationships reference valid entities
- [ ] T118 Test concurrent writes (loader + extractor)

### Documentation & Demo

- [ ] T119 Update README.md with setup instructions
- [ ] T120 Create demo script for stakeholders
- [ ] T121 Document lessons learned in `specs/001-cognitive-backbone-poc/lessons-learned.md`

### Bug Fixes

- [ ] T122 Fix issues found during integration testing
- [ ] T123 Improve error messages and user feedback

---

## Summary

**Total Tasks**: 123

**Tasks by User Story**:
- Setup & Foundation: 25 tasks (T001-T025)
- User Story 1 (P1 MVP): 20 tasks (T026-T045)
- User Story 2 (P2): 13 tasks (T046-T058)
- User Story 3 (P3): 16 tasks (T059-T074)
- User Story 4 (P4): 13 tasks (T075-T087)
- User Story 5 (P5): 20 tasks (T088-T107)
- Integration & Polish: 16 tasks (T108-T123)

**MVP Scope** (User Story 1):
Tasks T001-T045 deliver the core POC: email loading, entity extraction, graph storage, basic querying.

**Parallel Opportunities**:
- Phase 1: T006, T007, T008 (project structure setup)
- Phase 2: T014-T016 (schema definitions), T024-T025 (utilities)
- Phase 3: T026-T027 (loader components), T030-T032 (LLM clients)
- Phase 4: T046, T048, T050-T052 (API components)
- Phase 6: T075, T081 (TUI components)
- Phase 7: T092 (prompt engineering)

**Critical Path**: T001-T025 (Setup/Foundation) → T026-T045 (US1) → T046-T058 (US2) → T059-T074 (US3)

**Estimated Timeline**: 5-7 weeks (single developer, full-time)
