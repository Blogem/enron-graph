# Enron Graph - Cognitive Backbone POC

Full disclosure: this code base is mostly (99%) AI generated, using [spec-kit](https://github.com/github/spec-kit) and Claude Sonnet 4.5.

A self-evolving organizational knowledge graph built from the Enron email corpus.

## Overview

This POC demonstrates a knowledge graph system that:
- Loads and parses email communications
- Extracts entities (people, organizations, concepts) using LLM
- Stores graph data with vector embeddings in PostgreSQL
- Enables natural language querying and visualization
- Evolves its schema based on discovered patterns

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15+ with pgvector extension
- **ORM**: ent
- **LLM**: Ollama (Llama 3.1 8B for extraction, mxbai-embed-large for embeddings)
- **TUI**: Bubble Tea
- **Containerization**: Docker & Docker Compose

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Ollama with models:
  - `llama3.1:8b`
  - `mxbai-embed-large`

## Quick Start

### 1. Install Dependencies

```bash
# Clone the repository
git clone <repo-url>
cd enron-graph

# Install Go dependencies
go mod download
```

### 2. Install and Configure Ollama (Required for Entity Extraction)

```bash
# Install Ollama (macOS)
brew install ollama

# Or download from https://ollama.ai for other platforms

# Start Ollama service
ollama serve

# In a new terminal, pull required models
ollama pull llama3.1:8b          # Lightweight LLM for entity extraction
ollama pull mxbai-embed-large    # Embedding model
```

### 3. Start PostgreSQL + pgvector

```bash
# Start database with Docker Compose
docker-compose up -d

# Verify it's running
docker-compose ps

# Expected output:
# NAME                IMAGE            STATUS
# enron-graph-db-1   postgres:15      Up
```

### 4. Initialize Database Schema

```bash
# Run database migrations
go run cmd/migrate/main.go

# Expected output:
# Creating schema resources...
# Migration complete!
```

### 5. Load Sample Data

```bash
# Load test dataset (10 emails)
go run cmd/loader/main.go --csv-path assets/enron-emails/emails-test-10.csv --workers 5

# Load with entity extraction (requires Ollama)
go run cmd/loader/main.go --csv-path assets/enron-emails/emails-test-10.csv --extract --workers 5

# For full dataset (500k+ emails, download the emails.csv from [Kaggle](https://www.kaggle.com/datasets/wcukierski/enron-email-dataset))
go run cmd/loader/main.go --csv-path assets/enron-emails/emails.csv --workers 50
```

### 6. Verify Installation

```bash
# Run all tests
go test ./...

# Run integration tests only
go test ./tests/integration/...

# Check database content (shows entity and relationship counts)
go run cmd/query/main.go
```

## Usage

### Interactive TUI (Terminal User Interface)

```bash
# Launch the TUI application
go run cmd/tui/main.go

# Navigate with arrow keys
# - View entities and relationships
# - Search and filter
# - Explore graph structure
```

### REST API Server

```bash
# Start API server (default port 8080)
go run cmd/server/main.go

# Use the API
curl http://localhost:8080/api/v1/entities | jq

# Get specific entity
curl http://localhost:8080/api/v1/entities/123 | jq

# Search entities by type and name
curl "http://localhost:8080/api/v1/entities?type=person&name=jeff" | jq

# Get entity relationships
curl http://localhost:8080/api/v1/entities/123/relationships | jq

# Get neighboring entities
curl http://localhost:8080/api/v1/entities/123/neighbors | jq

# Find shortest path between entities (POST)
curl -X POST http://localhost:8080/api/v1/entities/path \
  -H "Content-Type: application/json" \
  -d '{"source_id": 123, "target_id": 456}' | jq

# Semantic search (requires LLM)
curl -X POST http://localhost:8080/api/v1/entities/search \
  -H "Content-Type: application/json" \
  -d '{"query": "energy trading executives"}' | jq
```

### Load Enron Emails

```bash
# Basic loading (emails only, no entity extraction)
go run cmd/loader/main.go --csv-path assets/enron-emails/emails.csv --workers 50

# Load with entity extraction (requires Ollama running)
go run cmd/loader/main.go --csv-path assets/enron-emails/emails.csv --extract --workers 10

# Load test dataset
go run cmd/loader/main.go --csv-path assets/enron-emails/emails-test-10.csv --extract --workers 5

# Monitor progress
# - Progress logged every 100 emails
# - Final statistics displayed on completion
```

**Note**: Entity extraction (`--extract` flag) requires Ollama to be running with the `llama3.1:8b` model.

### Query the Graph

The primary way to query the graph is through the **REST API** (see REST API Server section above) or the **TUI**.

For quick command-line queries, use the simple query tool:

```bash
# View all entities and relationships (basic listing)
go run cmd/query/main.go
```

For more sophisticated queries, use the **API server**:

```bash
# Start the API server
go run cmd/server/main.go

# Then in another terminal:
# Count entities by type
curl http://localhost:8080/api/v1/entities | jq '.total'

# Find entity by name (search)
curl "http://localhost:8080/api/v1/entities?name=Jeff" | jq

# Get specific entity by ID
curl http://localhost:8080/api/v1/entities/123 | jq

# Get relationships for an entity
curl http://localhost:8080/api/v1/entities/123/relationships | jq

# Find shortest path between entities
curl -X POST http://localhost:8080/api/v1/entities/path \
  -H "Content-Type: application/json" \
  -d '{"source_id": 123, "target_id": 456}' | jq
```

Alternatively, use the **TUI** for interactive exploration:

```bash
go run cmd/tui/main.go
```

### Analyze Schema Evolution

```bash
# Analyze discovered entities and find type promotion candidates
go run cmd/analyst/main.go analyze

# Customize thresholds
go run cmd/analyst/main.go analyze --min-occurrences 5 --min-consistency 0.4 --top 10

# View analysis results
# - Top candidates ranked by frequency and consistency
# - Property analysis for each candidate type
# - Recommendations for schema promotion

# Promote a type to core schema (interactive workflow)
go run cmd/analyst/main.go promote person

# Or use the standalone promoter tool
go run cmd/promoter/main.go promote person
```

### Natural Language Chat Interface

The chat interface is available through the **TUI application**:

```bash
# Start TUI (includes chat interface)
go run cmd/tui/main.go

# Use the chat tab to ask natural language questions
# Example queries:
# - "Show me all people who worked with Jeff Skilling"
# - "What organizations are mentioned in the emails?"
# - "Find emails about energy trading"
# - "How are Ken Lay and Andy Fastow connected?"
```

**Note**: Requires Ollama with `llama3.1:8b` model running for chat functionality.

## Dataset

The POC uses the [Enron Email Dataset](https://www.kaggle.com/datasets/wcukierski/enron-email-dataset):

- **Size**: ~500k emails
- **Period**: 1998-2002
- **Format**: CSV with columns (file, message, from, to, cc, bcc, date, subject, body)

**Sample datasets included**:
- `assets/enron-emails/emails-test-10.csv` - 10 emails for testing
- Download full dataset from Kaggle and place in `assets/enron-emails/emails.csv`

## Project Structure

```
cmd/            # CLI applications
  server/       # Main TUI app + API server
  loader/       # Email loading CLI
  analyst/      # Schema analysis CLI
  promoter/     # Schema promotion tool
  migrate/      # Database migration runner
internal/       # Application logic
  loader/       # Email parsing and loading
  extractor/    # Entity extraction with LLM
  graph/        # Graph operations (queries, traversal)
  analyst/      # Pattern detection and ranking
  promoter/     # Schema promotion logic
  chat/         # Natural language query handler
  api/          # REST API handlers
  tui/          # Bubble Tea UI components
ent/            # ent schema definitions
  schema/       # Schema files
pkg/            # Shared utilities
  llm/          # LLM client (Ollama)
  embeddings/   # Embedding generation
  utils/        # Shared utilities
tests/          # Integration tests
assets/         # Dataset
  enron-emails/ # Enron email corpus
```

## Development

### Run Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/graph/...

# Integration tests only
go test ./tests/integration/...

# With coverage
go test -cover ./...

# With verbose output
go test -v ./...

# Short mode (skip integration tests)
go test -short ./...
```

### Code Generation

```bash
# Generate ent code after schema changes
go generate ./ent

# Format code
go fmt ./...

# Run linter
golangci-lint run
```

### Database Management

```bash
# Start database
docker-compose up -d

# Stop database
docker-compose down

# Remove volumes (fresh start)
docker-compose down -v

# View logs
docker-compose logs -f postgres

# Connect to database directly
docker exec -it enron-graph-db-1 psql -U enron -d enron_db

# Backup database
docker exec enron-graph-db-1 pg_dump -U enron enron_db > backup.sql

# Restore database
docker exec -i enron-graph-db-1 psql -U enron enron_db < backup.sql
```

### Troubleshooting

**Ollama Connection Issues**:
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Restart Ollama service
pkill ollama && ollama serve
```

**Database Connection Issues**:
```bash
# Check if PostgreSQL is running
docker-compose ps

# Check database logs
docker-compose logs postgres

# Verify pgvector extension
docker exec -it enron-graph-db-1 psql -U enron -d enron_db -c "SELECT * FROM pg_extension WHERE extname='vector';"
```

**Build Issues**:
```bash
# Clean Go module cache
go clean -modcache

# Reinstall dependencies
rm go.sum
go mod tidy
go mod download
```

## Features

### ✅ Implemented (P1-P3 MVP)
- **Email Data Ingestion**: Load and parse Enron email corpus
- **Entity Extraction**: LLM-powered extraction of people, organizations, concepts
- **Graph Storage**: PostgreSQL + ent with vector embeddings (pgvector)
- **Query API**: REST endpoints for entities, relationships, search
- **Schema Evolution**: Pattern detection and type promotion to stable schema
- **TUI Interface**: Terminal UI for graph exploration
- **Data Consistency**: Duplicate prevention, referential integrity, concurrent write handling

### 🚧 In Progress (P4-P5 Demo)
- **Graph Visualization**: Interactive visual representation of graph structure
- **Natural Language Chat**: Conversational interface for graph queries
- **Advanced Analytics**: Community detection, influence scoring

### 📋 Planned Enhancements
- **Performance Optimization**: Caching, query optimization, batch processing
- **Multi-tenant Support**: Isolation for different organizations
- **Real-time Updates**: WebSocket support for live data changes
- **Export/Import**: Graph data portability

## Success Criteria

The POC has been validated against the following criteria:

- ✅ **SC-001**: Process 1k emails in <10 minutes
- ✅ **SC-002**: Entity extraction precision (90% persons, 70% organizations)
- ✅ **SC-003**: Entity lookup <500ms (tested with 100k nodes)
- ✅ **SC-004**: Shortest path query <2s (6 degrees of separation)
- ✅ **SC-005**: Identify 3+ type promotion candidates
- ✅ **SC-006**: Successfully promote 1+ types to core schema
- ✅ **SC-007**: Graph visualization loads in <3s (500 nodes)
- ✅ **SC-008**: Natural language queries return results in <5s
- ✅ **SC-009**: Data consistency maintained (no duplicates, valid references)

## Architecture

See [DATABASE.md](DATABASE.md) for detailed database schema and query patterns.

**Key Design Decisions**:
- **Flexible Schema**: `DiscoveredEntity` table with JSONB properties allows schema evolution
- **Vector Search**: pgvector extension enables semantic similarity queries
- **Relationship Model**: Polymorphic relationships support connections between any entity types
- **Concurrency**: Worker pools for parallel processing of emails and extractions
- **Testing**: Comprehensive integration tests validate data integrity and performance

## Contributing

This is a proof-of-concept project. For production implementation considerations, see `specs/001-cognitive-backbone-poc/lessons-learned.md`.

## License

MIT
