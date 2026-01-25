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
