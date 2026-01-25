# T108: Manual Test - TUI Navigation Flows

**Task ID**: T108  
**User Story**: US4 - Basic Visualization of Graph Structure  
**Test Type**: Manual Testing  
**Status**: To be executed  
**Test Date**: TBD  
**Tester**: TBD

## Objective

Verify that the TUI (Terminal User Interface) provides smooth navigation flows for browsing entities, filtering, searching, viewing details, and visualizing relationships in the knowledge graph.

## Prerequisites

- Database is populated with email data (run loader)
- Entities have been extracted (run extractor)
- TUI binary is built: `go build -o tui cmd/tui/main.go`
- PostgreSQL database is running

## Test Environment Setup

```bash
# 1. Ensure database is running
docker-compose up -d postgres

# 2. Verify data is loaded
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT COUNT(*) FROM emails;"
docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT COUNT(*) FROM discovered_entities;"

# 3. Build and launch TUI
go build -o tui cmd/tui/main.go
./tui
```

## Test Cases

### TC1: Start TUI and Navigate Entity List

**Steps**:
1. Launch TUI application: `./tui`
2. Observe welcome screen
3. Press Enter or designated key to enter entity list view
4. Use arrow keys (↑/↓) to navigate through entities
5. Observe pagination if entity count > screen height

**Expected Results**:
- ✓ TUI launches without errors
- ✓ Welcome screen displays with instructions
- ✓ Entity list view shows entities with their properties
- ✓ Arrow keys smoothly navigate up/down the list
- ✓ Current selection is clearly highlighted
- ✓ Status bar shows total entity count and current view
- ✓ Pagination works smoothly for large datasets

**Actual Results**: works

**Pass/Fail**: ⬜ **PASS** ⬜ FAIL

**Notes/Issues**: fixed some issues during testing

---

### TC2: Filter by Entity Type

**Steps**:
1. In entity list view, press the filter key (e.g., 'f' or 'F1')
2. Observe filter options (PERSON, ORG, PROJECT, etc.)
3. Select a specific entity type (e.g., PERSON)
4. Verify list updates to show only filtered entities
5. Clear filter and verify all entities return

**Expected Results**:
- ✓ Filter menu appears with all available entity types
- ✓ Selecting a type updates the list immediately
- ✓ Only entities matching the selected type are shown
- ✓ Status bar indicates active filter
- ✓ Entity count updates to reflect filtered count
- ✓ Clearing filter restores full list

**Actual Results**: works

**Pass/Fail**: ⬜ **PASS** ⬜ FAIL

**Notes/Issues**: some fixes applied

---

### TC3: Search by Name

**Steps**:
1. In entity list view, press the search key (e.g., '/' or 's')
2. Type a search query (e.g., "enron", "jeff", "meeting")
3. Press Enter to execute search
4. Verify results show entities matching the query
5. Test case sensitivity (if applicable)
6. Clear search and verify all entities return

**Expected Results**:
- ✓ Search input box appears
- ✓ Typing query updates search field
- ✓ Search executes on Enter
- ✓ Results show only matching entities
- ✓ Search is case-insensitive (or follows specified behavior)
- ✓ Partial matches work correctly
- ✓ Status bar shows search term
- ✓ Clearing search restores full list

**Actual Results**: works

**Pass/Fail**: ⬜ **PASS** ⬜ FAIL

**Notes/Issues**: none

---

### TC4: Select Entity and View Details

**Steps**:
1. Navigate to an entity in the list
2. Press Enter or detail key (e.g., 'd') to view entity details
3. Observe detailed information display
4. Verify all entity properties are shown
5. Navigate back to entity list

**Expected Results**:
- ✓ Detail view opens for selected entity
- ✓ All entity properties displayed (name, type, confidence, etc.)
- ✓ Formatting is clear and readable
- ✓ Relationships to other entities are listed
- ✓ Back button/key (e.g., 'Esc' or 'b') returns to entity list
- ✓ Selected entity remains highlighted after returning

**Actual Results**: works

**Pass/Fail**: ⬜ **PASS** ⬜ FAIL

**Notes/Issues**: fixed some stuff during testing

---

### TC5: Visualize Entity as Graph

**Steps**:
1. In entity list or detail view, select an entity
2. Press graph visualization key (e.g., 'g' or 'v')
3. Observe ASCII graph rendering
4. Verify selected entity appears as central node
5. Verify connected entities appear as connected nodes
6. Verify edges/relationships are rendered

**Expected Results**:
- ✓ Graph view renders without errors
- ✓ ASCII graph displays nodes and edges clearly
- ✓ Selected entity is visually distinguished (highlighted/centered)
- ✓ Connected entities appear with correct relationships
- ✓ Graph layout is readable and not overlapping
- ✓ Legend shows node types and edge types (if applicable)

**Actual Results**: works

**Pass/Fail**: ⬜ **PASS** ⬜ FAIL

**Notes/Issues**: none

---

### TC6: Navigate Graph View and Expand Nodes

**Steps**:
1. In graph view, use arrow keys to navigate between nodes
2. Select a connected node
3. Press expand/collapse key (e.g., 'e' or Enter)
4. Verify node expands to show its connections
5. Navigate to newly expanded nodes
6. Test collapsing nodes
7. Navigate back to entity list view

**Expected Results**:
- ✓ Arrow keys navigate between visible nodes
- ✓ Current node is clearly highlighted
- ✓ Expanding a node reveals its connections
- ✓ Graph re-renders with new nodes/edges
- ✓ Layout adjusts to accommodate new nodes
- ✓ Collapsing nodes removes their connections from view
- ✓ Graph remains navigable during expansion/collapse
- ✓ Exit key (e.g., 'Esc' or 'q') returns to entity list

**Actual Results**: works

**Pass/Fail**: ⬜ **PASS** ⬜ FAIL

**Notes/Issues**: none

---

## Performance Observations

**Success Criteria**: SC-007 - Render time <3s for 500 nodes

- Graph rendering time for 100 nodes: _______ ms
- Graph rendering time for 500 nodes: _______ ms (must be < 3000ms)
- Scrolling performance in entity list: _______ (Smooth/Laggy)
- Filter/search response time: _______ ms

**Pass/Fail Performance**: ⬜ **PASS** ⬜ FAIL

---

## Overall Test Summary

**Total Test Cases**: 6  
**Passed**: _____  
**Failed**: _____  
**Pass Rate**: _____%

**Critical Issues Found**: *(List any blocking issues)*

**Minor Issues Found**: *(List any non-blocking issues)*

**Recommendations**: *(Fill in after testing)*

---

## Sign-off

**Tester Name**: _____________________  
**Date Completed**: __________________  
**Overall Result**: ⬜ PASS ⬜ FAIL  
**Approved By**: _____________________
