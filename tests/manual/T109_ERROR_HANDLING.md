# T109: Manual Test - Error Handling

**Task ID**: T109  
**User Story**: US4 - Basic Visualization of Graph Structure  
**Test Type**: Manual Testing - Error Scenarios  
**Status**: To be executed  
**Test Date**: TBD  
**Tester**: TBD

## Objective

Verify that the TUI (Terminal User Interface) gracefully handles error conditions including empty database, network disconnection, and invalid entity selection. Ensure error messages are clear and the application remains stable.

## Prerequisites

- TUI binary is built: `go build -o tui cmd/tui/main.go`
- PostgreSQL database is available (may be stopped for testing)
- Test database environment set up

## Test Environment Setup

```bash
# Build TUI
go build -o tui cmd/tui/main.go

# Prepare test scenarios
# - Empty database: Create fresh database with schema but no data
# - Network test: Running database that can be stopped mid-session
# - Invalid selection: Database with test data
```

---

## Test Cases

### TC1: Empty Database Handling

**Objective**: Verify TUI handles database with schema but no data gracefully

**Setup Steps**:
```bash
# Option 1: Fresh database with schema only
docker-compose up -d postgres
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "TRUNCATE emails, discovered_entities, relationships CASCADE;"

# Option 2: Run migrations on empty database
go run cmd/migrate/main.go
```

**Test Steps**:
1. Launch TUI with empty database: `./tui`
2. Observe welcome screen behavior
3. Navigate to entity list view
4. Attempt to search/filter on empty list
5. Try to enter graph view
6. Attempt to view entity details

**Expected Results**:
- ✓ TUI launches without crashing
- ✓ Welcome screen appears normally
- ✓ Entity list view shows "No entities found" or similar message
- ✓ Empty state message is clear and helpful
- ✓ Suggests action (e.g., "Run loader to import data")
- ✓ Search/filter operations handle empty results gracefully
- ✓ Graph view shows empty state or disables visualization
- ✓ No panic/crash when navigating empty database
- ✓ Status bar correctly shows "0 entities"

**Actual Results**: works

**Pass/Fail**: ⬜ **PASS** ⬜ FAIL

**Error Messages Observed**: *(Document exact error messages)*

**Notes/Issues**: *(Fill in after testing)*

---

### TC2: Network Disconnection During Active Session

**Objective**: Verify TUI handles database connection loss gracefully

**Setup Steps**:
```bash
# Start with populated database
docker-compose up -d postgres
# Ensure entities exist
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT COUNT(*) FROM discovered_entities;"
# Launch TUI
./tui
```

**Test Steps**:
1. Launch TUI with working database connection
2. Navigate to entity list (confirm data loads)
3. **While TUI is running**, stop database:
   ```bash
   docker-compose stop postgres
   ```
4. Attempt to navigate entities
5. Try to load entity details
6. Try to render graph view
7. Attempt search/filter operations
8. Restart database and observe recovery behavior

**Expected Results**:
- ✓ TUI detects connection loss
- ✓ Clear error message displayed: "Database connection lost" or similar
- ✓ Application doesn't crash or hang indefinitely
- ✓ Timeout is reasonable (e.g., 5-10 seconds)
- ✓ User is informed about retry options or exit
- ✓ Cached/loaded data remains visible (if applicable)
- ✓ User can gracefully exit TUI after connection loss
- ✓ If database reconnects, TUI recovers or prompts restart

**Actual Results**: *(Fill in after testing)*

**Pass/Fail**: ⬜ PASS ⬜ FAIL

**Error Messages Observed**: *(Document exact error messages)*

**Recovery Behavior**: *(Document how app behaves when DB reconnects)*

**Notes/Issues**: *(Fill in after testing)*

---

### TC3: Network Disconnection on TUI Startup

**Objective**: Verify TUI handles database unavailability at launch

**Setup Steps**:
```bash
# Ensure database is stopped
docker-compose stop postgres
```

**Test Steps**:
1. Launch TUI with database stopped: `./tui`
2. Observe startup behavior and error messages
3. Wait for timeout
4. Attempt to retry connection (if option available)
5. Exit application

**Expected Results**:
- ✓ TUI attempts connection with reasonable timeout
- ✓ Clear error message: "Unable to connect to database" or similar
- ✓ Error includes connection details (host, port) for debugging
- ✓ Suggests troubleshooting steps
- ✓ Application exits gracefully, doesn't hang
- ✓ Non-zero exit code for scripting/automation
- ✓ Option to retry connection (if implemented)

**Actual Results**: *(Fill in after testing)*

**Pass/Fail**: ⬜ PASS ⬜ FAIL

**Error Messages Observed**: *(Document exact error messages)*

**Notes/Issues**: *(Fill in after testing)*

---

### TC4: Invalid Entity Selection

**Objective**: Verify TUI handles edge cases in entity selection and navigation

**Setup Steps**:
```bash
# Ensure database has limited test data
docker-compose up -d postgres
./tui
```

**Test Steps**:
1. Navigate to entity list with data
2. **Test Case A**: Select last entity, press down arrow (beyond list bounds)
3. **Test Case B**: Select first entity, press up arrow (before list start)
4. **Test Case C**: Navigate to entity details, delete entity from DB (in another terminal), try to navigate
   ```bash
   docker exec enron-graph-postgres psql -U enron -d enron_graph -c "DELETE FROM discovered_entities WHERE id = <selected_id>;"
   ```
5. **Test Case D**: Apply filter that returns zero results
6. **Test Case E**: Search for non-existent entity
7. **Test Case F**: In graph view, attempt to expand node with no relationships

**Expected Results**:
- ✓ **Bounds checking**: Navigation stays within list bounds, no panic
- ✓ **Deleted entity**: Error message or refresh shows entity removed
- ✓ **Zero results filter**: Clear message "No entities match filter"
- ✓ **Empty search**: Clear message "No results found for '<query>'"
- ✓ **Node with no relationships**: Graph shows isolated node or message
- ✓ All error states allow user to return to normal navigation
- ✓ No crashes or unexpected behavior

**Actual Results**: *(Fill in after testing)*

**Pass/Fail**: ⬜ PASS ⬜ FAIL

**Error Messages Observed**: *(Document exact error messages)*

**Notes/Issues**: *(Fill in after testing)*

---

### TC5: Malformed Data Handling

**Objective**: Verify TUI handles corrupted or unexpected data in database

**Setup Steps**:
```bash
# Insert entity with NULL or malformed fields
docker exec enron-graph-postgres psql -U enron -d enron_graph << EOF
INSERT INTO discovered_entities (name, type_category, confidence_score) 
VALUES (NULL, 'PERSON', 0.5);

INSERT INTO discovered_entities (name, type_category, confidence_score) 
VALUES ('Test Entity', NULL, 0.5);

INSERT INTO discovered_entities (name, type_category, confidence_score) 
VALUES ('Invalid Confidence', 'PERSON', -1.0);
EOF
```

**Test Steps**:
1. Launch TUI
2. Navigate to entity list
3. Attempt to view entities with NULL/malformed data
4. Observe rendering behavior
5. Try to visualize malformed entities in graph view

**Expected Results**:
- ✓ TUI handles NULL values gracefully
- ✓ Missing fields show as "(none)" or similar placeholder
- ✓ Invalid values are sanitized or flagged
- ✓ Malformed data doesn't crash rendering
- ✓ Error logged but app continues functioning
- ✓ Graph view skips or handles malformed entities

**Actual Results**: *(Fill in after testing)*

**Pass/Fail**: ⬜ PASS ⬜ FAIL

**Error Messages Observed**: *(Document exact error messages)*

**Notes/Issues**: *(Fill in after testing)*

---

### TC6: Resource Exhaustion (Large Dataset)

**Objective**: Verify TUI handles very large entity counts without crashing

**Setup Steps**:
```bash
# If available, load large dataset
# Or generate synthetic entities
docker exec enron-graph-postgres psql -U enron -d enron_graph << EOF
INSERT INTO discovered_entities (unique_id, name, type_category, confidence_score, created_at)
SELECT 
  'test-entity-' || generate_series,
  'Entity_' || generate_series || '_Test',
  'PERSON',
  0.8,
  NOW()
FROM generate_series(1, 10000);
EOF
```

**Test Steps**:
1. Launch TUI with 10,000+ entities
2. Navigate entity list
3. Scroll through large dataset
4. Apply filters
5. Search entities
6. Attempt to visualize large subgraphs

**Expected Results**:
- ✓ TUI launches without excessive delay (<10s)
- ✓ Pagination handles large datasets efficiently
- ✓ Memory usage remains reasonable (monitor with `top`/`htop`)
- ✓ No memory leaks during navigation
- ✓ Filtering/search performance acceptable
- ✓ Graph view either limits nodes or handles large graphs
- ✓ Success Criteria SC-007: <3s for 500 nodes still met

**Actual Results**: *(Fill in after testing)*

**Pass/Fail**: ⬜ PASS ⬜ FAIL

**Performance Metrics**:
- Launch time: _______ seconds
- Memory usage: _______ MB
- Scroll lag: _______ (None/Minimal/Noticeable)

**Notes/Issues**: *(Fill in after testing)*

---

## Error Message Quality Assessment

For each error encountered, rate the error messages:

| Error Scenario | Clarity (1-5) | Actionability (1-5) | Technical Detail (1-5) | Notes |
|----------------|---------------|---------------------|------------------------|-------|
| Empty database |               |                     |                        |       |
| DB disconnection |             |                     |                        |       |
| DB unavailable |               |                     |                        |       |
| Invalid selection |            |                     |                        |       |
| Malformed data |               |                     |                        |       |
| Large dataset  |               |                     |                        |       |

**Rating Scale**:
- 1 = Poor (confusing, no help)
- 2 = Below Average (somewhat unclear)
- 3 = Adequate (understandable, basic help)
- 4 = Good (clear, actionable)
- 5 = Excellent (crystal clear, helpful, guides user)

---

## Overall Test Summary

**Total Test Cases**: 6  
**Passed**: _____  
**Failed**: _____  
**Pass Rate**: _____%

**Critical Issues Found**: *(List any blocking issues)*

**Minor Issues Found**: *(List any non-blocking issues)*

**Error Handling Quality**: ⬜ Excellent ⬜ Good ⬜ Adequate ⬜ Needs Improvement

**Recommendations**: *(Fill in after testing)*

---

## Sign-off

**Tester Name**: _____________________  
**Date Completed**: __________________  
**Overall Result**: ⬜ PASS ⬜ FAIL  
**Approved By**: _____________________
