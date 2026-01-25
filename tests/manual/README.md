# Manual Testing Documentation

This directory contains documentation and scripts for manual testing of the TUI (Terminal User Interface).

## Test Documents

### T108: TUI Navigation Flows
**File**: [T108_TUI_NAVIGATION_FLOWS.md](./T108_TUI_NAVIGATION_FLOWS.md)

Tests all navigation flows in the TUI including:
- Entity list navigation
- Filtering by entity type
- Searching by name
- Viewing entity details
- Graph visualization
- Node expansion and navigation

### T109: Error Handling
**File**: [T109_ERROR_HANDLING.md](./T109_ERROR_HANDLING.md)

Tests error handling scenarios including:
- Empty database handling
- Network disconnection during active session
- Network disconnection on startup
- Invalid entity selection
- Malformed data handling
- Resource exhaustion (large datasets)

## Quick Start

### Prerequisites Check
```bash
# Run the automated setup and guide script
./run_manual_tests.sh
```

This script will:
1. Verify database is running
2. Check entity data exists
3. Build TUI binary if needed
4. Provide step-by-step testing instructions

### Manual Execution

If you prefer to run tests manually:

```bash
# 1. Ensure database is running
docker-compose up -d postgres

# 2. Verify data exists
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT COUNT(*) FROM discovered_entities;"

# 3. Build TUI
go build -o tui cmd/tui/main.go

# 4. Launch TUI and follow test cases in documentation
./tui
```

## Test Execution Workflow

1. **Read the test document** - Understand test objectives and expected results
2. **Set up test environment** - Follow prerequisite steps in each document
3. **Execute test cases** - Run each TC (Test Case) step by step
4. **Document results** - Fill in "Actual Results" and "Pass/Fail" sections
5. **Note issues** - Record any bugs, unexpected behavior, or improvements
6. **Sign off** - Complete the summary and sign-off sections

## Recording Results

For each test case, document:

- **Actual Results**: What actually happened during the test
- **Pass/Fail**: Check the appropriate box
- **Error Messages**: Copy exact error text (if any)
- **Notes/Issues**: Observations, bugs found, or suggestions

## Test Status Tracking

After completing tests, update the status in the main task file:

```markdown
specs/001-cognitive-backbone-poc/tasks.md
```

Mark tasks as complete:
```markdown
- [X] T108 [US4] Manual test: TUI navigation flows
- [X] T109 [US4] Manual test: Error handling
```

## Automated vs Manual Testing

**Automated Tests** (Unit, Integration, Contract):
- Located in `tests/integration/`, `internal/*/test.go`
- Run via `go test ./...`
- Continuous verification

**Manual Tests** (This directory):
- Exploratory testing
- User experience validation
- Visual/UI verification
- Error message quality assessment
- Performance observation

## Tips for Effective Manual Testing

1. **Take notes in real-time** - Don't rely on memory
2. **Screenshot errors** - Visual evidence helps debugging
3. **Test edge cases** - Go beyond happy path
4. **Verify error messages** - Are they clear and actionable?
5. **Check performance** - Note any lag or delays
6. **Document environment** - OS, terminal, screen size, etc.
7. **Retest after fixes** - Verify issues are resolved

## Common Issues and Solutions

### Database Connection Fails
```bash
# Check if database is running
docker ps | grep postgres

# Restart database
docker-compose restart postgres

# Check logs
docker logs enron-graph-postgres
```

### No Entities in Database
```bash
# Run loader to import data
go run cmd/loader/main.go --file loader/sample_emails.csv

# Run extractor to generate entities
go run cmd/extractor/main.go
```

### TUI Build Fails
```bash
# Check Go version (requires 1.21+)
go version

# Clean build cache
go clean -cache

# Rebuild
go build -o tui cmd/tui/main.go
```

## Questions or Issues?

If you encounter issues during manual testing:

1. Check the test document for troubleshooting steps
2. Review the main specification: `specs/001-cognitive-backbone-poc/spec.md`
3. Check implementation details in `internal/tui/`
4. Document the issue in the test results

## Test Metrics

Track these metrics during testing:

- **Test Duration**: How long did testing take?
- **Issues Found**: Count of bugs/problems discovered
- **Pass Rate**: Percentage of test cases passed
- **Performance**: Response times, render times
- **Usability**: Subjective experience rating

Record metrics in the test document summary sections.
