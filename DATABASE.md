# Database Configuration

This project supports using different database instances for development and testing.

## Available Databases

- **enron_graph** - Production/development database
- **enron_graph_test** - Test database (for experimentation)

## Switching Databases

### Using Environment Variables

```bash
# Use test database
export DB_NAME=enron_graph_test
go run cmd/tui/main.go

# Use development database
export DB_NAME=enron_graph
go run cmd/tui/main.go
```

### Using Environment Files

```bash
# Load test database configuration
source .env.test
go run cmd/tui/main.go

# Load development database configuration
source .env.dev
go run cmd/tui/main.go
```

## Quick Commands

### Test Database

```bash
# Run TUI with test database
DB_NAME=enron_graph_test go run cmd/tui/main.go

# Load data into test database
DB_NAME=enron_graph_test go run cmd/loader/main.go --csv-path tests/fixtures/sample_emails.csv

# Run migrations on test database
DB_NAME=enron_graph_test go run cmd/migrate/main.go
```

### Development Database

```bash
# Run TUI with dev database (default)
go run cmd/tui/main.go

# Or explicitly
DB_NAME=enron_graph go run cmd/tui/main.go
```

## Reset Test Database

To reset the test database with fresh data:

```bash
# Drop and recreate
docker exec enron-graph-postgres psql -U enron -d postgres -c "DROP DATABASE IF EXISTS enron_graph_test;"
docker exec enron-graph-postgres psql -U enron -d postgres -c "CREATE DATABASE enron_graph_test;"

# Run migrations
DB_NAME=enron_graph_test go run cmd/migrate/main.go

# Load test data
docker exec -i enron-graph-postgres psql -U enron -d enron_graph_test < scripts/test-data.sql
```

## Promoted Tables and Relationship Handling

The system uses a two-phase entity lifecycle:

1. **Discovery Phase**: Entities are extracted from emails and stored in the `discovered_entities` table
2. **Promotion Phase**: High-frequency entity types are migrated to dedicated tables for better performance

### Relationship Updates During Promotion

When entities are promoted from `discovered_entities` to a dedicated table (e.g., `persons`), the system automatically:

- **Tracks ID Mapping**: Uses PostgreSQL's `RETURNING` clause to map old IDs to new auto-increment IDs
- **Updates FROM References**: Updates `relationships.from_type` and `relationships.from_id` to point to the promoted table
- **Updates TO References**: Updates `relationships.to_type` and `relationships.to_id` to point to the promoted table
- **Atomic Migration**: All operations occur within a single transaction - either all succeed or all rollback
- **Cleanup**: Optionally deletes migrated entities from `discovered_entities` after successful promotion

### Type-Aware Queries

The repository layer supports type hints for efficient querying:

```go
// O(1) lookup when type is known
entity, err := repo.FindEntityByUniqueID(ctx, "john@enron.com", "Person")

// Three-tier fallback when type is unknown:
// 1. Check discovered_entities (most common)
// 2. Infer type from relationships table
// 3. Parallel search across all promoted tables
entity, err := repo.FindEntityByUniqueID(ctx, "john@enron.com")
```

### Promoted Type Registry

The system maintains a global registry of promoted types:

- **EntityCreator**: Functions to create entities in promoted tables
- **EntityFinder**: Functions to find entities by unique_id in promoted tables
- **Auto-Registration**: Generated code automatically registers new types during initialization

Promoted tables follow the naming convention: `{type_name}s` (e.g., `persons`, `organizations`, `contracts`)
