# Enron Graph - Cognitive Backbone POC

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

## Setup

### 1. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Pull Ollama models (if not already done)
ollama pull llama3.1:8b
ollama pull mxbai-embed-large
```

### 2. Start PostgreSQL + pgvector

```bash
# Start database
docker-compose up -d

# Verify it's running
docker-compose ps
```

### 3. Initialize Database Schema

```bash
# Generate ent code
go generate ./ent

# Run migrations
go run cmd/migrate/main.go
```

## Usage

### Load Enron Emails

Download the emails.csv from [Kaggle](https://www.kaggle.com/datasets/wcukierski/enron-email-dataset).

```bash
# Load emails from CSV
go run cmd/loader/main.go --csv-path assets/enron-emails/emails.csv --batch-size 100 --workers 50

# Load with entity extraction
go run cmd/loader/main.go --csv-path assets/enron-emails/emails.csv --extract
```

### Query the Graph

```bash
# Start TUI application
go run cmd/server/main.go
```

### Analyze Schema Evolution

```bash
# Detect type promotion candidates
go run cmd/analyst/main.go analyze

# Promote a type to core schema
go run cmd/analyst/main.go promote --type person
```

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
# Unit tests
go test ./...

# Integration tests
go test ./tests/integration/...

# With coverage
go test -cover ./...
```

### Database Management

```bash
# Stop database
docker-compose down

# Remove volumes (fresh start)
docker-compose down -v

# View logs
docker-compose logs -f postgres
```

## Features

### Phase 1: Core POC (P1-P3)
- ✅ Email data ingestion
- ✅ Entity extraction with LLM
- ✅ Graph storage with embeddings
- ✅ Query API
- ✅ Schema evolution

### Phase 2: Demo Features (P4-P5)
- 🚧 TUI visualization
- 🚧 Natural language chat interface

## Success Criteria

- SC-001: Process 1k emails in <10 minutes
- SC-002: Entity extraction precision (90% persons, 70% orgs)
- SC-003: Entity lookup <500ms (100k nodes)
- SC-004: Shortest path <2s (6 degrees)
- SC-005: Identify 3+ type promotion candidates
- SC-006: Successfully promote 1+ types to core schema

## License

MIT
