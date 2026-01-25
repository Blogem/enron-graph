#!/bin/bash
# Quick automated verification for T108 & T109 prerequisites
# This script verifies that the TUI can connect and basic operations work
# Note: This does NOT replace manual testing, but validates environment is ready

set -e

REPO_ROOT="/Users/jochem/code/enron-graph-2"
cd "$REPO_ROOT"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "======================================================================"
echo "T108 & T109: Automated Environment Verification"
echo "======================================================================"
echo ""

PASS_COUNT=0
FAIL_COUNT=0

# Test 1: Database connectivity
echo -n "TEST 1: Database connectivity... "
if docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT 1;" >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    echo -e "${RED}FAIL${NC}"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

# Test 2: Entity table exists
echo -n "TEST 2: Entity table exists... "
if docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT COUNT(*) FROM discovered_entities;" >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    echo -e "${RED}FAIL${NC}"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

# Test 3: Entities exist in database
echo -n "TEST 3: Entities exist in database... "
ENTITY_COUNT=$(docker exec enron-graph-postgres psql -U enron -d enron_graph -t -c "SELECT COUNT(*) FROM discovered_entities;" | tr -d ' ')
if [ "$ENTITY_COUNT" -gt 0 ]; then
    echo -e "${GREEN}PASS${NC} ($ENTITY_COUNT entities)"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    echo -e "${RED}FAIL${NC} (0 entities - run loader and extractor first)"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

# Test 4: TUI binary exists
echo -n "TEST 4: TUI binary exists... "
if [ -f "./tui" ]; then
    echo -e "${GREEN}PASS${NC}"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    echo -e "${YELLOW}SKIP${NC} (building...)"
    go build -o tui cmd/tui/main.go
    if [ -f "./tui" ]; then
        echo -e "${GREEN}PASS${NC} (built successfully)"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo -e "${RED}FAIL${NC}"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
fi

# Test 5: TUI can initialize (dry run check)
echo -n "TEST 5: TUI can compile and run help... "
if ./tui --help >/dev/null 2>&1 || timeout 2 ./tui >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    # TUI might not have --help flag, which is OK
    echo -e "${YELLOW}SKIP${NC} (TUI may not have --help, manual test required)"
fi

# Test 6: Relationship table exists
echo -n "TEST 6: Relationship table exists... "
if docker exec enron-graph-postgres psql -U enron -d enron_graph -c "SELECT COUNT(*) FROM relationships;" >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    echo -e "${RED}FAIL${NC}"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

# Test 7: Check for diverse entity types
echo -n "TEST 7: Multiple entity types exist... "
TYPE_COUNT=$(docker exec enron-graph-postgres psql -U enron -d enron_graph -t -c "SELECT COUNT(DISTINCT type_category) FROM discovered_entities;" 2>/dev/null | tr -d ' ' || echo "0")
if [ "$TYPE_COUNT" -gt 1 ]; then
    echo -e "${GREEN}PASS${NC} ($TYPE_COUNT types)"
    PASS_COUNT=$((PASS_COUNT + 1))
else
    echo -e "${YELLOW}WARN${NC} (Only $TYPE_COUNT entity type - filtering tests limited)"
fi

echo ""
echo "======================================================================"
echo "Results: $PASS_COUNT passed, $FAIL_COUNT failed"
echo "======================================================================"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓ Environment ready for manual testing!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Review test documents:"
    echo "     - tests/manual/T108_TUI_NAVIGATION_FLOWS.md"
    echo "     - tests/manual/T109_ERROR_HANDLING.md"
    echo "  2. Launch TUI: ./tui"
    echo "  3. Execute test cases and document results"
    echo ""
    exit 0
else
    echo -e "${RED}✗ Environment has issues. Fix failures before manual testing.${NC}"
    echo ""
    exit 1
fi
