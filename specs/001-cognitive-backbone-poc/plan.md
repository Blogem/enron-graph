# Implementation Plan: Cognitive Backbone POC

**Feature**: 001-cognitive-backbone-poc  
**Plan Date**: 2026-01-24  
**Planner**: GitHub Copilot  
**Status**: Approved - Ready for Implementation

## Executive Summary

This plan outlines the implementation approach for building a self-evolving organizational knowledge graph POC using the Enron email corpus. The system will be built in Go using PostgreSQL + ent for graph storage, Ollama for local LLM inference, and Bubble Tea for TUI visualization. Implementation follows a phased approach prioritizing core functionality (P1-P3) before demo enhancements (P4-P5).

**Key Technical Decisions**:
- **Language**: Go 1.21+ (performance, concurrency, type safety)
- **Storage**: PostgreSQL + ent ORM + pgvector extension
- **LLM**: Ollama (Llama 3.1 8B local) + LangChainGo + optional LiteLLM for API access
- **Embeddings**: mxbai-embed-large (Ollama) with fallback to OpenAI via LiteLLM
- **UI**: Bubble Tea TUI (terminal-first, web fallback if needed)
- **Migrations**: ent built-in migrations with optional Atlas Go for complex scenarios

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI/TUI Application                      │
│  (Bubble Tea - Chat, Visualization, Admin Commands)         │
└───────────────┬─────────────────────────────────────────────┘
                │
┌───────────────▼─────────────────────────────────────────────┐
│                    Application Layer (Go)                    │
├──────────────────────────────────────────────────────────────┤
│  Email Loader  │  Entity Extractor  │  Schema Analyst       │
│  Query Engine  │  Chat Handler      │  API Server           │
└───────────────┬──────────────┬──────────────────────────────┘
                │              │
        ┌───────▼──────┐  ┌───▼──────────┐
        │  Graph Core  │  │ LLM Services │
        │ (ent + PG)   │  │   (Ollama)   │
        └───────┬──────┘  └──────────────┘
                │
        ┌───────▼──────────────────┐
        │  PostgreSQL + pgvector   │
        │  (Graph + Vector Store)  │
        └──────────────────────────┘
```

### Data Model (ent Schema)

**Core Entities** (defined upfront):

```go
// Email - Communication event
type Email struct {
    ent.Schema
    ID         int       `json:"id"`
    MessageID  string    `json:"message_id" unique:"true"`
    From       string    `json:"from"`
    To         []string  `json:"to"`
    CC         []string  `json:"cc"`
    BCC        []string  `json:"bcc"`
    Subject    string    `json:"subject"`
    Date       time.Time `json:"date"`
    Body       string    `json:"body"`
    FilePath   string    `json:"file_path"`
    
    // Edges
    Edges EmailEdges `json:"edges"`
}

// DiscoveredEntity - Flexible schema for extracted entities before promotion
type DiscoveredEntity struct {
    ent.Schema
    ID              int               `json:"id"`
    UniqueID        string            `json:"unique_id" unique:"true"` // email, domain, etc.
    TypeCategory    string            `json:"type_category"`           // "person", "organization", "concept", etc.
    Name            string            `json:"name"`
    Properties      map[string]any    `json:"properties"`              // JSONB in PostgreSQL
    Embedding       []float32         `json:"embedding"`               // pgvector
    ConfidenceScore float64           `json:"confidence_score"`
    CreatedAt       time.Time         `json:"created_at"`
    
    // Edges
    Edges DiscoveredEntityEdges `json:"edges"`
}

// Relationship - Connections between entities
type Relationship struct {
    ent.Schema
    ID              int       `json:"id"`
    Type            string    `json:"type"`           // SENT, RECEIVED, MENTIONS, COMMUNICATES_WITH
    FromType        string    `json:"from_type"`      // "email", "discovered_entity", "person", etc.
    FromID          int       `json:"from_id"`
    ToType          string    `json:"to_type"`
    ToID            int       `json:"to_id"`
    Timestamp       time.Time `json:"timestamp"`
    ConfidenceScore float64   `json:"confidence_score"`
    Properties      map[string]any `json:"properties"` // Additional metadata
    
    CreatedAt       time.Time `json:"created_at"`
}

// SchemaPromotion - Audit log for type promotions
type SchemaPromotion struct {
    ent.Schema
    ID                  int            `json:"id"`
    TypeName            string         `json:"type_name"`
    PromotedAt          time.Time      `json:"promoted_at"`
    PromotionCriteria   map[string]any `json:"promotion_criteria"` // frequency, density, consistency
    EntitiesAffected    int            `json:"entities_affected"`
    ValidationFailures  int            `json:"validation_failures"`
    SchemaDefinition    map[string]any `json:"schema_definition"`  // Generated schema rules
}
```

**Promoted Entities** (added dynamically via code generation):

```go
// Example: After promotion, these are generated via `ent generate`
type Person struct {
    ent.Schema
    ID          int      `json:"id"`
    Email       string   `json:"email" unique:"true"`
    Name        string   `json:"name"`
    Department  string   `json:"department,omitempty"`
    Embedding   []float32 `json:"embedding"`
    
    // Original discovered entity reference
    SourceEntityID int `json:"source_entity_id"`
    
    Edges PersonEdges `json:"edges"`
}

type Organization struct {
    ent.Schema
    ID          int      `json:"id"`
    Name        string   `json:"name" unique:"true"`
    Domain      string   `json:"domain"`
    Embedding   []float32 `json:"embedding"`
    
    SourceEntityID int `json:"source_entity_id"`
    
    Edges OrganizationEdges `json:"edges"`
}
```

### Promotion Workflow

1. **Discovery Phase**: Entities extracted → stored in `DiscoveredEntity` table
2. **Analysis Phase**: Schema Analyst identifies patterns → ranks candidates
3. **User Approval**: TUI presents candidates → user confirms promotion
4. **Code Generation**:
   ```bash
   # Generate new ent schema file
   go run cmd/schema-promoter/main.go --type "Person" --schema schema-def.json
   # This creates: ent/schema/person.go
   
   # Generate ent code
   go generate ./ent
   
   # Run migrations
   go run cmd/migrate/main.go
   ```
5. **Data Migration**: Copy data from `DiscoveredEntity` to new `Person` table
6. **Validation**: Check constraints, log failures in `SchemaPromotion`

### Technology Stack

| Layer | Technology | Version | Purpose |
|-------|-----------|---------|---------|
| **Language** | Go | 1.21+ | Application logic |
| **Database** | PostgreSQL | 15+ | Graph and vector storage |
| **ORM** | ent | v0.12+ | Schema definition, queries, migrations |
| **Vector Extension** | pgvector | 0.5+ | Embedding storage and similarity search |
| **Migrations** | ent migrations + Atlas Go (optional) | Latest | Schema evolution |
| **LLM Provider** | Ollama | Latest | Local model inference |
| **LLM Model** | Llama 3.1 8B | - | Entity extraction, NL queries |
| **Embedding Model** | mxbai-embed-large | - | Vector embeddings (1024-dim) |
| **LLM Abstraction** | LangChainGo | Latest | Prompt management, LLM orchestration |
| **API Fallback** | LiteLLM | Latest | Optional API access (OpenAI, etc.) |
| **TUI Framework** | Bubble Tea | v0.25+ | Terminal interface |
| **HTTP Router** | Chi | v5+ | REST API (lightweight) |
| **CSV Parser** | encoding/csv | stdlib | Enron dataset parsing |
| **Logging** | log/slog | stdlib | Structured logging |
| **Testing** | testing + testify | stdlib + v1.8+ | Unit and integration tests |

### Project Structure

```
enron-graph/
├── cmd/
│   ├── server/          # Main TUI application + API server
│   ├── loader/          # Email loading CLI
│   ├── analyst/         # Schema analysis CLI
│   ├── promoter/        # Schema promotion tool
│   └── migrate/         # Database migration runner
├── internal/
│   ├── loader/          # Email parsing and loading logic
│   ├── extractor/       # Entity extraction with LLM
│   ├── graph/           # Graph operations (queries, traversal)
│   ├── analyst/         # Pattern detection and ranking
│   ├── promoter/        # Schema promotion logic
│   ├── chat/            # Natural language query handler
│   ├── api/             # REST API handlers
│   └── tui/             # Bubble Tea UI components
├── ent/
│   ├── schema/          # ent schema definitions
│   └── generated/       # Generated ent code
├── pkg/
│   ├── llm/             # LLM client (Ollama + LiteLLM)
│   ├── embeddings/      # Embedding generation
│   └── utils/           # Shared utilities
├── assets/
│   └── enron-emails/    # Dataset
├── specs/               # Specifications
├── migrations/          # SQL migrations (if using Atlas)
├── tests/
│   ├── integration/     # Integration tests
│   └── fixtures/        # Test data
├── scripts/             # Helper scripts
├── go.mod
├── go.sum
└── README.md
```

## Implementation Phases

### Phase 0: Setup & Research (1-2 days)

**Objective**: Validate technical assumptions and set up development environment

**Tasks**:
1. **Environment Setup**
   - [ ] Install Go 1.21+
   - [ ] Install PostgreSQL 15+ with pgvector extension (Docker)
   - [ ] Install Ollama and pull models: `llama3.1:8b`, `mxbai-embed-large`
   - [ ] Initialize Go module: `go mod init github.com/Blogem/enron-graph`
   - [ ] Install ent: `go get entgo.io/ent/cmd/ent`

2. **Dataset Inspection**
   - [ ] Load `assets/enron-emails/emails.csv` and count records
   - [ ] Inspect CSV structure (verify 2 columns: `file`, `message`)
   - [ ] Sample 10 emails and parse headers manually
   - [ ] Document data quality issues (encoding, missing fields, duplicates)
   - [ ] Estimate dataset size (rows, storage)

3. **Technology Validation**
   - [ ] Create minimal ent schema with JSONB and vector fields
   - [ ] Test pgvector setup: create table with vector column, insert sample data
   - [ ] Test HNSW index creation and similarity search performance
   - [ ] Prototype Ollama integration: call Llama 3.1 8B with sample prompt
   - [ ] Prototype embedding generation: call mxbai-embed-large via Ollama
   - [ ] Test LangChainGo with Ollama provider
   - [ ] Evaluate Bubble Tea: build simple TUI with list navigation

4. **Performance Baselines**
   - [ ] Benchmark pgvector similarity search: 1k, 10k, 100k vectors (HNSW)
   - [ ] Measure Ollama inference speed: tokens/sec for Llama 3.1 8B on M4
   - [ ] Measure embedding generation speed: embeddings/sec for mxbai-embed-large
   - [ ] Estimate total processing time for 1k emails

**Deliverables**:
- Research findings document (`specs/001-cognitive-backbone-poc/research.md`)
- Working Docker Compose setup (PostgreSQL + pgvector)
- Proof-of-concept scripts for Ollama and pgvector

---

### Phase 1: Core Graph Engine (3-4 days) - P1

**Objective**: Implement graph storage with ent + PostgreSQL + pgvector

**User Story**: Foundation for User Story 1 (Email Data Ingestion)

**Tasks**:

1. **Database Setup**
   - [ ] Create Docker Compose file: PostgreSQL 15 + pgvector extension
   - [ ] Write initialization script: enable pgvector, create database
   - [ ] Document connection string and credentials

2. **Ent Schema Definition**
   - [ ] Define `Email` schema (FR-007, FR-008)
   - [ ] Define `DiscoveredEntity` schema with JSONB properties and vector field (FR-006, FR-014a, FR-017b)
   - [ ] Define `Relationship` schema (FR-015, FR-015a)
   - [ ] Define `SchemaPromotion` schema for audit log (FR-023)
   - [ ] Generate ent code: `go generate ./ent`
   - [ ] Run initial migration

3. **Graph Core Implementation** (`internal/graph/`)
   - [ ] Create `Repository` interface with methods:
     - `CreateEmail(email *ent.Email) error`
     - `CreateDiscoveredEntity(entity *ent.DiscoveredEntity) error`
     - `CreateRelationship(rel *ent.Relationship) error`
     - `FindEntityByID(id int) (*ent.DiscoveredEntity, error)`
     - `FindEntityByUniqueID(uniqueID string) (*ent.DiscoveredEntity, error)`
     - `FindEntitiesByType(typeCategory string) ([]*ent.DiscoveredEntity, error)`
     - `TraverseRelationships(fromID int, relType string, depth int) ([]*ent.DiscoveredEntity, error)`
     - `FindShortestPath(fromID, toID int) ([]*ent.Relationship, error)`
     - `SimilaritySearch(embedding []float32, topK int, threshold float64) ([]*ent.DiscoveredEntity, error)`
   - [ ] Implement repository with ent client
   - [ ] Add pgvector similarity search using SQL: `SELECT * FROM discovered_entities ORDER BY embedding <-> $1 LIMIT $2`

4. **Testing**
   - [ ] Unit tests for repository methods (mock data)
   - [ ] Integration tests: insert entities, query back (FR-001, FR-002)
   - [ ] Test relationship traversal (1-hop, n-hop) (FR-003)
   - [ ] Test vector similarity search with sample embeddings
   - [ ] Validate performance: <500ms for entity lookup on 1k entities (SC-003 baseline)

**Acceptance Criteria**:
- ✅ Entities and relationships can be stored and queried
- ✅ Vector similarity search returns relevant results
- ✅ All repository methods have passing tests
- ✅ Database schema is versioned and reproducible

---

### Phase 2: Email Loader (2-3 days) - P1

**Objective**: Parse Enron CSV and load emails into graph

**User Story**: User Story 1, Acceptance Scenario 1

**Tasks**:

1. **CSV Parser** (`internal/loader/parser.go`)
   - [ ] Implement `ParseCSV(filePath string) chan Email` 
     - Stream CSV rows (don't load entire file into memory)
     - Parse `file` and `message` columns
   - [ ] Implement `ParseEmailHeaders(message string) (EmailMetadata, error)`
     - Extract: Message-ID, Date (RFC 2822), From, To, CC, BCC, Subject
     - Use Go `net/mail` package for header parsing
   - [ ] Handle encoding issues: detect UTF-8, Latin-1, convert to UTF-8 (FR-011)
   - [ ] Return errors for malformed emails (log, don't crash) (FR-011a)

2. **Batch Processor** (`internal/loader/processor.go`)
   - [ ] Implement `ProcessBatch(emails []Email, repo graph.Repository) error`
     - Check for duplicates via message-id (FR-010)
     - Insert emails into database
     - Track processed count and failure rate
   - [ ] Implement concurrent processing: 10-100 goroutines (FR-009a)
   - [ ] Add progress logging: every 100 emails processed
   - [ ] Halt on critical errors (database connection failure) (FR-011b)

3. **CLI Tool** (`cmd/loader/main.go`)
   - [ ] Accept arguments: `--csv-path`, `--batch-size`, `--workers`
   - [ ] Initialize database connection and repository
   - [ ] Call parser and processor
   - [ ] Report summary: total processed, failures, duration
   - [ ] Exit with error code if failure rate >2% (FR-011a)

4. **Testing**
   - [ ] Unit tests for header parsing (sample emails with various formats)
   - [ ] Integration test: load 100 sample emails, verify in database
   - [ ] Test duplicate handling: load same email twice, verify single record
   - [ ] Test error handling: malformed CSV, missing headers, corrupt data
   - [ ] Benchmark: process 1k emails in <2 minutes (partial SC-001)

**Acceptance Criteria**:
- ✅ CLI tool successfully loads 1k+ Enron emails
- ✅ Emails are deduplicated by message-id
- ✅ Failure rate <2%
- ✅ Processing completes in <2 minutes for 1k emails

---

### Phase 3: Entity Extractor (5-7 days) - P1

**Objective**: Extract entities and relationships from emails using LLM

**User Story**: User Story 1, Acceptance Scenarios 2-4

**Tasks**:

1. **LLM Client** (`pkg/llm/client.go`)
   - [ ] Implement Ollama client using LangChainGo
     - `GenerateCompletion(prompt string) (string, error)`
     - `GenerateEmbedding(text string) ([]float32, error)`
   - [ ] Add LiteLLM fallback client (optional, for API access)
   - [ ] Add retry logic (3 retries with exponential backoff)
   - [ ] Add timeout (30s for completion, 10s for embedding)

2. **Prompt Engineering** (`internal/extractor/prompts.go`)
   - [ ] Design extraction prompt template:
     ```
     Extract entities from this email. Return JSON with:
     - persons: [{name, email, confidence}]
     - organizations: [{name, domain, confidence}]
     - concepts: [{name, keywords, confidence}]
     - other: [{type, name, properties, confidence}]
     
     Email:
     From: {{.From}}
     To: {{.To}}
     Subject: {{.Subject}}
     Body: {{.Body}}
     
     Output JSON only, no explanation.
     ```
   - [ ] Test prompt on 10 sample emails, validate JSON output
   - [ ] Iterate on prompt to improve accuracy (target 70%+ precision)

3. **Entity Extractor** (`internal/extractor/extractor.go`)
   - [ ] Implement `ExtractFromEmail(email *ent.Email, llm LLMClient) (ExtractionResult, error)`
     - Parse email headers → high-confidence person entities (FR-012)
     - Call LLM with prompt → parse JSON response (FR-013, FR-014)
     - Generate embeddings for each entity (FR-017b)
     - Assign confidence scores based on extraction method (FR-017)
     - Filter entities below 0.7 confidence threshold (FR-017a)
   - [ ] Implement deduplication logic:
     - Person: email address as unique key
     - Organization: normalize name (lowercase, trim), check existing
     - Concept: use embedding similarity (cosine >0.85 threshold)
   - [ ] Create relationships (FR-015):
     - SENT: person (from email) → email
     - RECEIVED: email → person (to/cc/bcc)
     - MENTIONS: email → organization/concept
     - COMMUNICATES_WITH: person ↔ person (inferred from SENT/RECEIVED)

4. **Batch Extraction** (`internal/extractor/batch.go`)
   - [ ] Implement `ProcessEmailBatch(emails []*ent.Email, repo graph.Repository) error`
     - Process 10-100 emails concurrently (FR-009a)
     - Generate embeddings in batches (10-50 at a time)
     - Handle LLM rate limits and errors gracefully
     - Progress logging: every 50 emails
   - [ ] Track extraction metrics: success rate, average confidence, entities per email

5. **Integration with Loader** (`cmd/loader/main.go`)
   - [ ] Add `--extract` flag to enable entity extraction during loading
   - [ ] Initialize LLM client and extractor
   - [ ] Call extractor after email insertion
   - [ ] Report extraction summary: entities created, relationships created

6. **Testing**
   - [ ] Unit tests for prompt parsing (mock LLM responses)
   - [ ] Integration test: extract from 10 sample emails, verify entities in DB
   - [ ] Test deduplication: multiple emails mentioning same person/org
   - [ ] Test confidence filtering: verify entities below 0.7 are excluded
   - [ ] Test embedding generation: verify vectors are stored correctly
   - [ ] Benchmark: extract from 100 emails, measure time and accuracy
   - [ ] Validate SC-002: 90%+ precision for persons, 70%+ for organizations

**Acceptance Criteria**:
- ✅ Entities are extracted with confidence scores
- ✅ Embeddings are generated and stored in pgvector
- ✅ Relationships are created correctly (SENT, RECEIVED, MENTIONS, COMMUNICATES_WITH)
- ✅ Deduplication works for persons and organizations
- ✅ Extraction achieves precision targets (SC-002)
- ✅ 1k emails processed with extraction in <10 minutes (SC-001)

---

### Phase 4: Query API (3-4 days) - P2

**Objective**: Expose graph query capabilities via REST API

**User Story**: User Story 2, Acceptance Scenarios 1-4

**Tasks**:

1. **API Design** (`specs/001-cognitive-backbone-poc/contracts/api.md`)
   - [ ] Define REST endpoints:
     - `GET /entities/:id` - Get entity by ID (FR-024, FR-025)
     - `GET /entities?type=:type&name=:name` - Search entities (FR-025)
     - `GET /entities/:id/relationships` - Get entity relationships (FR-025)
     - `GET /entities/:id/neighbors?depth=:n` - Traverse relationships (FR-003)
     - `POST /entities/path` - Find shortest path between entities (FR-003)
     - `POST /entities/search` - Semantic search by embedding (FR-017b)
   - [ ] Define request/response schemas (JSON)
   - [ ] Document error codes and status codes

2. **API Implementation** (`internal/api/handlers.go`)
   - [ ] Implement handlers using Chi router
   - [ ] Map HTTP requests to graph repository methods
   - [ ] Add input validation (entity IDs, query parameters)
   - [ ] Format responses as JSON (FR-026)
   - [ ] Add error handling (404 for missing entities, 400 for invalid input)
   - [ ] Add performance logging (request duration)

3. **Query Engine Enhancements** (`internal/graph/queries.go`)
   - [ ] Optimize entity lookup with indexes (name, type, unique_id)
   - [ ] Implement efficient shortest path algorithm (BFS or Dijkstra)
   - [ ] Add pagination for large result sets (limit, offset)
   - [ ] Add filtering by confidence score, date range

4. **Server Setup** (`cmd/server/main.go`)
   - [ ] Initialize HTTP server with Chi router
   - [ ] Add middleware: logging, CORS (for future web UI), panic recovery
   - [ ] Configure database connection pooling
   - [ ] Add graceful shutdown

5. **Testing**
   - [ ] Contract tests for each endpoint (testify/assert)
   - [ ] Integration tests with test database (insert fixtures, query API)
   - [ ] Test pagination and filtering
   - [ ] Test error cases (missing entities, invalid IDs)
   - [ ] Performance tests:
     - Entity lookup: <500ms for 100k nodes (SC-003)
     - Shortest path: <2s for 6 degrees (SC-004)
   - [ ] Load test: 100 concurrent requests

**Acceptance Criteria**:
- ✅ All API endpoints return correct data
- ✅ Performance targets met (SC-003, SC-004)
- ✅ Error handling is robust
- ✅ API is documented with examples

---

### Phase 5: Schema Analyst (4-5 days) - P3

**Objective**: Detect patterns and promote entity types to core schema

**User Story**: User Story 3, Acceptance Scenarios 1-5

**Tasks**:

1. **Pattern Detection** (`internal/analyst/detector.go`)
   - [ ] Implement `AnalyzeDiscoveredEntities(repo graph.Repository) ([]Candidate, error)`
     - Query all discovered entities grouped by type_category (FR-018)
     - Calculate frequency: COUNT(*) per type
     - Calculate relationship density: AVG(degree) per type
     - Calculate property consistency: % of entities with each property
   - [ ] Implement embedding-based clustering:
     - Group entities by vector similarity (cosine >0.85)
     - Identify type candidates from clusters
     - Combine with type_category tags for better accuracy

2. **Candidate Ranking** (`internal/analyst/ranker.go`)
   - [ ] Implement `RankCandidates(candidates []Candidate) []RankedCandidate`
     - Score = 0.4*frequency + 0.3*density + 0.3*consistency (FR-019)
     - Apply thresholds: min 50 occurrences, min 70% property consistency
     - Sort by score descending
     - Return top 10 candidates

3. **Schema Generator** (`internal/analyst/schema_gen.go`)
   - [ ] Implement `GenerateSchema(candidate RankedCandidate) (SchemaDefinition, error)` (FR-020)
     - Infer required properties (>90% presence)
     - Infer optional properties (30-90% presence)
     - Infer data types from sample values (string, int, bool, timestamp)
     - Generate validation rules (min/max length, regex patterns)
     - Output JSON schema definition

4. **Promotion Tool** (`cmd/promoter/main.go`)
   - [ ] Create ent schema file generator:
     - Input: JSON schema definition
     - Output: `ent/schema/{typename}.go` file with ent schema struct
   - [ ] Implement `PromoteType(schemaPath string)`:
     - Generate ent schema file
     - Run `go generate ./ent`
     - Run database migration (create new table)
     - Copy data from `DiscoveredEntity` to new table (FR-022)
     - Validate entities against schema, log failures
     - Create `SchemaPromotion` audit record (FR-023)

5. **CLI Tool** (`cmd/analyst/main.go`)
   - [ ] Implement `analyze` subcommand:
     - Run pattern detection
     - Rank candidates
     - Display top 10 with metrics (FR-021a)
   - [ ] Implement `promote` subcommand:
     - Prompt user to select candidate (interactive) (FR-021b)
     - Generate schema definition
     - Confirm promotion with user
     - Call promotion tool

6. **Testing**
   - [ ] Unit tests for pattern detection (mock data with known patterns)
   - [ ] Unit tests for ranking algorithm
   - [ ] Unit tests for schema generation (verify property inference)
   - [ ] Integration test: full promotion workflow
     - Insert 100 discovered "person" entities
     - Run analyst, verify candidate identified
     - Promote type, verify new table created
     - Verify entities migrated correctly
   - [ ] Test validation failures (entities missing required properties)
   - [ ] Validate SC-005, SC-006: identify 3+ candidates, promote 1 successfully

**Acceptance Criteria**:
- ✅ Analyst identifies at least 3 candidate types from 1k email dataset (SC-005)
- ✅ Schema generation produces valid ent schema files
- ✅ Promotion creates new database table and migrates data (SC-006)
- ✅ Audit log records all promotions (SC-010)
- ✅ Validation failures are logged and handled gracefully

---

### Phase 6: TUI Visualization (3-4 days) - P4

**Objective**: Build terminal interface for graph exploration

**User Story**: User Story 4, Acceptance Scenarios 1-4

**Tasks**:

1. **TUI Framework Setup** (`internal/tui/`)
   - [ ] Choose framework: Bubble Tea (based on research)
   - [ ] Create base application structure:
     - Main model (state container)
     - Update function (handle messages)
     - View function (render UI)
   - [ ] Set up navigation: multiple views (entity list, graph view, details)

2. **Entity Browser** (`internal/tui/entity_list.go`)
   - [ ] Implement entity list view:
     - Display entities in table (ID, type, name, confidence)
     - Keyboard navigation (↑↓ arrows, page up/down)
     - Filter by type (F key → type selector)
     - Search by name (/ key → search input)
   - [ ] Fetch entities from repository (paginated)
   - [ ] Highlight selected entity

3. **Graph Visualization** (`internal/tui/graph_view.go`)
   - [ ] Implement ASCII graph rendering:
     - Node: `[Type: Name]` (color-coded by type)
     - Edge: `---[REL_TYPE]-->` (directional arrows)
     - Layout: Tree or force-directed (simple algorithm)
   - [ ] Limit to 50 nodes max (performance) (FR-027a)
   - [ ] Implement subgraph view around selected entity:
     - Show entity + immediate neighbors (1-hop)
     - Expand node: E key → fetch and add neighbors
   - [ ] Keyboard navigation:
     - Tab: cycle through nodes
     - Enter: select node → show details
     - E: expand node
     - B: go back to entity list

4. **Detail Panel** (`internal/tui/detail_view.go`)
   - [ ] Display entity properties (key-value pairs)
   - [ ] Display relationships (list of connected entities)
   - [ ] Actions:
     - V: visualize entity in graph view
     - R: show related entities
     - Q: close panel

5. **Main Application** (`cmd/server/main.go`)
   - [ ] Initialize Bubble Tea program
   - [ ] Connect to database repository
   - [ ] Add welcome screen with instructions
   - [ ] Add status bar (entity count, current view)

6. **Testing**
   - [ ] Manual testing with sample dataset (100 entities, 300 relationships)
   - [ ] Test navigation flows (list → graph → details → back)
   - [ ] Test graph rendering with 50+ nodes (verify performance) (SC-007)
   - [ ] Test keyboard shortcuts
   - [ ] Validate SC-007: <3s render time for 500 nodes (if performance issues, reduce to 50 nodes)

**Acceptance Criteria**:
- ✅ TUI displays entities and relationships
- ✅ Graph view renders ASCII graph with nodes and edges (FR-027, FR-027a)
- ✅ Keyboard navigation works smoothly (FR-028)
- ✅ Performance is acceptable for 50-500 nodes (SC-007)
- ✅ User can explore graph without technical knowledge

---

### Phase 7: Chat Interface (4-5 days) - P5

**Objective**: Natural language query interface for graph exploration

**User Story**: User Story 5, Acceptance Scenarios 1-6

**Tasks**:

1. **Chat Handler** (`internal/chat/handler.go`)
   - [ ] Implement `ProcessQuery(query string, context ConversationContext, llm LLMClient) (Response, error)` (FR-029, FR-030)
     - Send query + conversation history + graph schema to LLM
     - Parse LLM response (graph query or direct answer)
     - Execute query via repository
     - Format results for user
   - [ ] Implement conversation context management (FR-031):
     - Store last 5 queries and responses
     - Track entities mentioned (for pronoun resolution)
     - Include context in LLM prompt

2. **Query Patterns** (`internal/chat/patterns.go`)
   - [ ] Implement pattern matching for common queries (FR-032):
     - Entity lookup: "Who is X?" → `FindEntityByName(X)`
     - Relationship discovery: "Who did X email?" → `TraverseRelationships(X, "SENT")`
     - Path finding: "How are X and Y connected?" → `FindShortestPath(X, Y)`
     - Concept search: "Emails about energy" → `SimilaritySearch(embedding("energy"))`
     - Aggregations: "How many emails did X send?" → `CountRelationships(X, "SENT")`
   - [ ] Fallback to LLM for complex queries
   - [ ] Handle ambiguity: return multiple options, ask user to clarify (FR-034)

3. **Prompt Engineering** (`internal/chat/prompts.go`)
   - [ ] Design chat prompt template:
     ```
     You are a graph database assistant. Help the user query an organizational knowledge graph.
     
     Schema:
     - Entities: Person, Organization, Concept, Email
     - Relationships: SENT, RECEIVED, MENTIONS, COMMUNICATES_WITH
     
     Available operations:
     - Find entity by name
     - Get entity relationships
     - Find path between entities
     - Semantic search by concept
     
     Conversation history:
     {{.History}}
     
     User query: {{.Query}}
     
     Respond with either:
     1. A direct answer if information is in context
     2. A graph query command: {"operation": "...", "params": {...}}
     ```
   - [ ] Test prompt on 10 sample queries from FR-032

4. **TUI Chat View** (`internal/tui/chat_view.go`)
   - [ ] Implement chat interface:
     - Message history (scrollable)
     - Input box (text input)
     - Send message: Enter key
     - Clear history: Ctrl+L
   - [ ] Display query results:
     - Entity cards (name, type, snippet)
     - Relationship paths (visual arrow chains)
     - "Visualize" button → switch to graph view with results highlighted (FR-033)
   - [ ] Handle loading states (typing indicator while LLM processes)

5. **Integration with Main App** (`cmd/server/main.go`)
   - [ ] Add chat view to TUI navigation (C key → chat)
   - [ ] Initialize LLM client (Ollama)
   - [ ] Add session management (store conversation context)

6. **Testing**
   - [ ] Unit tests for pattern matching (verify correct query generation)
   - [ ] Integration tests with mock LLM (pre-defined responses)
   - [ ] Manual testing with 10 sample queries from FR-032:
     1. "Who is Jeff Skilling?"
     2. "Who did Jeff Skilling email?"
     3. "How are Jeff Skilling and Kenneth Lay connected?"
     4. "Find emails about energy trading"
     5. "What topics did Jeff Skilling discuss?"
     6. "How many emails did Jeff Skilling send?"
     7. "Show me information about Enron"
     8. "Tell me more about him" (follow-up)
     9. "Who else did he communicate with?" (follow-up)
     10. "Find people in the legal department"
   - [ ] Validate SC-012: 80%+ accuracy (8/10 queries successful)
   - [ ] Validate SC-013: 3 consecutive follow-up queries maintain context

**Acceptance Criteria**:
- ✅ Chat interface processes natural language queries (FR-029)
- ✅ Common query patterns work correctly (FR-032)
- ✅ Conversation context is maintained across queries (FR-031, SC-013)
- ✅ Ambiguity is handled with clarification prompts (FR-034)
- ✅ Query results can be visualized in graph view (FR-033)
- ✅ 80%+ accuracy on test queries (SC-012)

---

### Phase 8: Integration & Testing (2-3 days)

**Objective**: End-to-end testing and POC validation

**Tasks**:

1. **Integration Testing**
   - [ ] Test full workflow: load emails → extract entities → query → promote type → query promoted entities (SC-008)
   - [ ] Test with full 10k+ email dataset (SC-001)
   - [ ] Verify all user stories pass acceptance scenarios (P1, P2, P3)
   - [ ] Test P4, P5 demo features (optional but valuable)

2. **Performance Testing**
   - [ ] Validate SC-001: 10k emails processed in <10 minutes
   - [ ] Validate SC-002: Extraction precision (90% persons, 70% orgs)
   - [ ] Validate SC-003: Entity lookup <500ms (100k nodes)
   - [ ] Validate SC-004: Shortest path <2s (6 degrees)
   - [ ] Validate SC-007: Visualization <3s (500 nodes)
   - [ ] Validate SC-011: 5+ loose entity types discovered
   - [ ] Validate SC-012: Chat accuracy 80%+
   - [ ] Validate SC-013: Context maintained 3+ queries

3. **Data Consistency**
   - [ ] Validate SC-009: No duplicate entities with same unique ID
   - [ ] Validate SC-009: All relationships reference valid entities
   - [ ] Test concurrent writes (loader + extractor running simultaneously)

4. **Audit & Documentation**
   - [ ] Validate SC-010: Schema promotion audit log is complete
   - [ ] Document promotion workflow with screenshots
   - [ ] Create demo script for stakeholders

5. **Bug Fixes & Polish**
   - [ ] Fix issues found during integration testing
   - [ ] Improve error messages and user feedback
   - [ ] Add usage documentation (README updates)

**Acceptance Criteria**:
- ✅ All success criteria (SC-001 to SC-013) are met
- ✅ All P1-P3 user stories pass acceptance tests
- ✅ POC is ready for demo and stakeholder review
- ✅ Documentation is complete

---

## Testing Strategy

### Unit Tests
- **Coverage Target**: 70%+ for core logic (extractor, analyst, graph queries)
- **Framework**: Go `testing` package + `testify/assert`
- **Approach**: Test pure functions, mock external dependencies (LLM, database)
- **Examples**:
  - Email header parsing
  - Confidence score calculation
  - Schema property inference
  - Shortest path algorithm

### Integration Tests
- **Scope**: Multi-component workflows
- **Setup**: Test database (PostgreSQL in Docker with test schema)
- **Fixtures**: Sample emails, pre-populated entities
- **Examples**:
  - Load emails → verify in database
  - Extract entities → verify relationships created
  - Promote type → verify table created and data migrated
  - Query API → verify correct results

### Contract Tests
- **Scope**: API endpoints
- **Framework**: `httptest` package
- **Approach**: Test request/response schemas, status codes, error handling
- **Examples**:
  - `GET /entities/:id` returns 200 with valid JSON
  - `POST /entities/path` returns 404 if entities don't exist

### Performance Tests
- **Scope**: Critical path operations
- **Framework**: Go `testing/benchmark`
- **Metrics**: Latency (p50, p95, p99), throughput
- **Examples**:
  - Entity lookup: <500ms for 100k nodes
  - Shortest path: <2s for 6 degrees
  - Vector similarity: <500ms for top-50 results

### Manual Testing
- **Scope**: TUI, chat interface, end-to-end workflows
- **Approach**: Test with real Enron dataset, exploratory testing
- **Checklist**:
  - TUI navigation flows
  - Chat query patterns
  - Graph visualization rendering
  - Error handling edge cases

---

## Risk Mitigation

| Risk | Impact | Mitigation | Owner | Status |
|------|--------|------------|-------|--------|
| **LLM extraction accuracy <70%** | High | Iterate on prompts, test on sample data early, consider hybrid approach (rules + LLM) | Phase 3 | Open |
| **pgvector performance insufficient** | Medium | Benchmark early (Phase 0), optimize indexes (HNSW), limit vector dimensions if needed | Phase 0 | Open |
| **Ollama inference too slow** | Medium | Use smaller model (Llama 3.2 3B), batch processing, consider API fallback via LiteLLM | Phase 0 | Open |
| **Schema promotion complexity** | Medium | Start with simple promotion (manual code generation), automate incrementally | Phase 5 | Open |
| **TUI graph rendering limitations** | Low | Limit to 50 nodes, use tree layout, fallback to web UI if needed | Phase 6 | Open |
| **Chat query translation unreliable** | Low | Implement pattern matching fallback, test on known queries, set user expectations | Phase 7 | Open |
| **Enron dataset quality issues** | Low | Inspect early (Phase 0), document issues, handle gracefully in parser | Phase 0 | Open |

---

## Success Metrics Dashboard

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| **SC-001**: 10k emails processed in 10 min | <10 min | TBD | ⏳ |
| **SC-002**: Person extraction precision | 90%+ | TBD | ⏳ |
| **SC-002**: Org extraction precision | 70%+ | TBD | ⏳ |
| **SC-003**: Entity lookup latency | <500ms | TBD | ⏳ |
| **SC-004**: Shortest path latency | <2s | TBD | ⏳ |
| **SC-005**: Candidate types identified | 3+ | TBD | ⏳ |
| **SC-006**: Types promoted | 1+ | TBD | ⏳ |
| **SC-007**: Visualization render time | <3s | TBD | ⏳ |
| **SC-008**: End-to-end workflow | Pass | TBD | ⏳ |
| **SC-009**: Data consistency | Pass | TBD | ⏳ |
| **SC-010**: Audit log completeness | Pass | TBD | ⏳ |
| **SC-011**: Loose types discovered | 5+ | TBD | ⏳ |
| **SC-012**: Chat query accuracy | 80%+ | TBD | ⏳ |
| **SC-013**: Context maintenance | 3+ queries | TBD | ⏳ |

---

## Timeline Estimate

| Phase | Duration | Dependencies | Deliverables |
|-------|----------|--------------|--------------|
| **Phase 0**: Setup & Research | 1-2 days | None | Environment ready, research doc, POC scripts |
| **Phase 1**: Graph Engine | 3-4 days | Phase 0 | ent schemas, repository, tests |
| **Phase 2**: Email Loader | 2-3 days | Phase 1 | CLI tool, CSV parser, tests |
| **Phase 3**: Entity Extractor | 5-7 days | Phase 1, 2 | LLM integration, extraction logic, tests |
| **Phase 4**: Query API | 3-4 days | Phase 1 | REST API, handlers, tests |
| **Phase 5**: Schema Analyst | 4-5 days | Phase 1, 3 | Analyst CLI, promotion tool, tests |
| **Phase 6**: TUI Visualization | 3-4 days | Phase 4 | Bubble Tea UI, graph view, tests |
| **Phase 7**: Chat Interface | 4-5 days | Phase 4, 6 | Chat handler, TUI integration, tests |
| **Phase 8**: Integration & Testing | 2-3 days | All phases | End-to-end tests, bug fixes, docs |

**Total Estimate**: 27-37 days (5-7 weeks)

**Critical Path**: Phase 0 → 1 → 2 → 3 → 4 → 5 → 8 (P1-P3 required for POC)  
**Optional Path**: Phase 6 → 7 (P4-P5 demo features, can be done in parallel after Phase 4)

**Assumptions**:
- Single developer working full-time
- No major technical blockers
- Research phase resolves unknowns quickly
- LLM extraction achieves accuracy targets without extensive iteration

---

## Next Steps

### Immediate Actions (Today)

1. **Create Research Document**
   - [ ] Initialize `specs/001-cognitive-backbone-poc/research.md`
   - [ ] Document research questions from Phase 0

2. **Set Up Repository Structure**
   - [ ] Create directory structure: `cmd/`, `internal/`, `ent/`, `pkg/`, `tests/`
   - [ ] Initialize `go.mod`
   - [ ] Create `.gitignore`

3. **Set Up Development Environment**
   - [ ] Install Go 1.21+
   - [ ] Install PostgreSQL + pgvector (Docker Compose)
   - [ ] Install Ollama, pull models

4. **Begin Phase 0 Research**
   - [ ] Inspect Enron email dataset
   - [ ] Test pgvector setup
   - [ ] Prototype Ollama integration

### Week 1 (Phase 0 + Phase 1)
- Complete research and technology validation
- Implement core graph engine with ent + pgvector
- Write tests for repository methods

### Week 2 (Phase 2 + Phase 3 start)
- Implement email loader
- Begin entity extractor implementation
- Test LLM extraction on sample data

### Week 3 (Phase 3 complete + Phase 4)
- Complete entity extractor with embeddings
- Implement query API
- Integration testing for P1-P2

### Week 4 (Phase 5)
- Implement schema analyst
- Test promotion workflow
- Validate P3 user story

### Week 5 (Phase 6 + Phase 7)
- Implement TUI visualization
- Implement chat interface
- Validate P4-P5 user stories

### Week 6-7 (Phase 8)
- Integration and performance testing
- Bug fixes and polish
- Prepare demo and documentation

---

## Definition of Done

The POC is **DONE** when:

- ✅ All code is committed to `001-cognitive-backbone-poc` branch
- ✅ All tests pass (unit, integration, contract)
- ✅ All success criteria (SC-001 to SC-013) are validated
- ✅ All P1-P3 user stories pass acceptance scenarios
- ✅ P4-P5 demo features are functional (best effort)
- ✅ README.md includes setup and usage instructions
- ✅ Demo script is prepared for stakeholder review
- ✅ Lessons learned are documented for production design

---

**Plan Status**: Approved - Ready for Implementation  
**Next Action**: Begin Phase 0 research tasks  
**Started**: 2026-01-24 (Ollama models pulled)
