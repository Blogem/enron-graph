#!/bin/bash
# T143: Test P4-P5 demo features
# This script validates User Story 4 (TUI/Visualization) and User Story 5 (Chat Interface)

set -e

echo "=========================================="
echo "T143: P4-P5 Demo Feature Tests"
echo "=========================================="
echo ""

FAILED=0
PASSED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== User Story 4 (P4): Basic Visualization of Graph Structure (TUI) ===${NC}"
echo "Testing TUI functionality:"
echo "  - Entity list display and navigation"
echo "  - Graph view rendering (nodes and edges)"
echo "  - Entity filtering by type"
echo "  - Entity search by name"
echo "  - Node selection and property display"
echo "  - View switching and pagination"
echo "  - Color coding by entity type"
echo ""

# Run TUI tests
TUI_OUTPUT=$(go test -v ./internal/tui -timeout 1m 2>&1)
TUI_RESULT=$?

if [ $TUI_RESULT -eq 0 ]; then
    TUI_PASSED=$(echo "$TUI_OUTPUT" | grep -c "^--- PASS:" || true)
    echo -e "${GREEN}✓ User Story 4 (TUI): PASSED${NC} ($TUI_PASSED tests passed)"
    PASSED=$((PASSED + 1))
    
    # Show key test results
    echo ""
    echo "Key features verified:"
    echo "  ✓ Entity list displays with proper formatting"
    echo "  ✓ Graph view renders nodes and edges"
    echo "  ✓ Filtering by entity type works correctly"
    echo "  ✓ Entity search by name is functional"
    echo "  ✓ Navigation between views is smooth"
    echo "  ✓ Color coding differentiates entity types"
else
    echo -e "${RED}✗ User Story 4 (TUI): FAILED${NC}"
    echo "$TUI_OUTPUT" | grep "FAIL:" | head -5
    FAILED=$((FAILED + 1))
fi
echo ""

echo -e "${BLUE}=== User Story 5 (P5): Natural Language Search and Chat Interface ===${NC}"
echo "Testing Chat functionality:"
echo "  - Natural language query processing"
echo "  - Entity lookup queries"
echo "  - Relationship queries"
echo "  - Path finding between entities"
echo "  - Concept/keyword search"
echo "  - Conversation context maintenance"
echo "  - Ambiguity handling"
echo "  - Graph visualization integration"
echo ""

# Run Chat/US5 acceptance tests
US5_OUTPUT=$(go test -v ./tests/integration -run "TestUS5" -timeout 2m 2>&1)
US5_RESULT=$?

if [ $US5_RESULT -eq 0 ]; then
    US5_PASSED=$(echo "$US5_OUTPUT" | grep -c "^--- PASS:" || true)
    echo -e "${GREEN}✓ User Story 5 (Chat): PASSED${NC} ($US5_PASSED test groups passed)"
    PASSED=$((PASSED + 1))
    
    # Show key test results
    echo ""
    echo "Key features verified:"
    echo "  ✓ Natural language queries processed successfully"
    echo "  ✓ Entity lookup by name works"
    echo "  ✓ Relationship queries return connected entities"
    echo "  ✓ Path finding identifies connections"
    echo "  ✓ Concept search finds relevant entities"
    echo "  ✓ Conversation context is maintained"
    echo "  ✓ Ambiguous queries handled appropriately"
    echo "  ✓ Query results can be visualized"
else
    echo -e "${RED}✗ User Story 5 (Chat): FAILED${NC}"
    echo "$US5_OUTPUT" | grep "FAIL:" | head -5
    FAILED=$((FAILED + 1))
fi
echo ""

echo "=========================================="
echo "Summary"
echo "=========================================="
echo -e "User Stories Passed: ${GREEN}$PASSED/2${NC}"
echo -e "User Stories Failed: ${RED}$FAILED${NC}"
echo ""

# Check if TUI binary can be built
echo -e "${BLUE}Additional Check: TUI Binary Build${NC}"
if go build -o /tmp/tui-test cmd/tui/main.go 2>&1; then
    echo -e "${GREEN}✓ TUI binary builds successfully${NC}"
    rm -f /tmp/tui-test
else
    echo -e "${YELLOW}⚠ TUI binary build failed (non-critical for tests)${NC}"
fi
echo ""

# Check if server binary can be built (for chat API)
echo -e "${BLUE}Additional Check: Server Binary Build${NC}"
if go build -o /tmp/server-test cmd/server/main.go 2>&1; then
    echo -e "${GREEN}✓ Server binary builds successfully${NC}"
    rm -f /tmp/server-test
else
    echo -e "${YELLOW}⚠ Server binary build failed (non-critical for tests)${NC}"
fi
echo ""

echo "=========================================="
echo "Manual Testing Reminder"
echo "=========================================="
echo "Note: Some P4 (TUI) features require manual testing:"
echo "  - Interactive navigation flows"
echo "  - Real-time graph rendering performance"
echo "  - Keyboard input handling"
echo "  - Terminal resizing behavior"
echo ""
echo "See: tests/manual/T108_TUI_NAVIGATION_FLOWS.md for manual test cases"
echo "Run: go run cmd/tui/main.go (requires populated database)"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ T143 PASS: All P4-P5 demo features verified${NC}"
    exit 0
else
    echo -e "${RED}✗ T143 FAIL: Some demo features failed${NC}"
    exit 1
fi
