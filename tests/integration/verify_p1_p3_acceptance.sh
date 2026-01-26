#!/bin/bash
# T142: Verify all P1-P3 user stories pass acceptance scenarios
# This script runs all acceptance tests for User Stories 1-3 (Priority P1-P3)

set -e

echo "=========================================="
echo "T142: P1-P3 User Story Acceptance Tests"
echo "=========================================="
echo ""

FAILED=0
PASSED=0
SKIPPED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== User Story 1 (P1): Email Data Ingestion and Entity Extraction ===${NC}"
echo "Acceptance Scenarios:"
echo "  - T048: CSV parsing extracts metadata"
echo "  - T049: Extractor identifies entities and relationships"
echo "  - T050: Entities stored and queried"
echo "  - T051: Deduplication works"
echo "  - T054: Loose entity types discovered"
echo ""

US1_OUTPUT=$(go test -v ./tests/integration -run "TestAcceptance.*T04[8-9]|TestAcceptance.*T05[0-4]" -count=1 2>&1)
US1_RESULT=$?

if [ $US1_RESULT -eq 0 ]; then
    US1_PASSED=$(echo "$US1_OUTPUT" | grep -c "^--- PASS:" || true)
    US1_SKIPPED=$(echo "$US1_OUTPUT" | grep -c "^--- SKIP:" || true)
    echo -e "${GREEN}✓ User Story 1: PASSED${NC} ($US1_PASSED tests passed, $US1_SKIPPED skipped)"
    PASSED=$((PASSED + US1_PASSED))
    SKIPPED=$((SKIPPED + US1_SKIPPED))
else
    echo -e "${RED}✗ User Story 1: FAILED${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

echo -e "${BLUE}=== User Story 2 (P2): Graph Query and Exploration ===${NC}"
echo "Acceptance Scenarios:"
echo "  - T069: Query person by name"
echo "  - T070: Query relationships"
echo "  - T071: Shortest path between entities"
echo "  - T072: Filter by entity type"
echo "  - T073: Entity lookup performance"
echo "  - T074: Shortest path performance"
echo ""

US2_OUTPUT=$(go test -v ./tests/integration -run "TestAcceptanceT06[9]|TestAcceptanceT07[0-4]" -count=1 2>&1)
US2_RESULT=$?

if [ $US2_RESULT -eq 0 ]; then
    US2_PASSED=$(echo "$US2_OUTPUT" | grep -c "^--- PASS:" || true)
    US2_SKIPPED=$(echo "$US2_OUTPUT" | grep -c "^--- SKIP:" || true)
    echo -e "${GREEN}✓ User Story 2: PASSED${NC} ($US2_PASSED tests passed, $US2_SKIPPED skipped)"
    PASSED=$((PASSED + US2_PASSED))
    SKIPPED=$((SKIPPED + US2_SKIPPED))
else
    echo -e "${RED}✗ User Story 2: FAILED${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

echo -e "${BLUE}=== User Story 3 (P3): Schema Evolution through Type Promotion ===${NC}"
echo "Acceptance Scenarios:"
echo "  - T090: Analyst identifies entity types"
echo "  - T091: Candidates ranked by metrics"
echo "  - T092: Promotion adds type to schema"
echo "  - T093: Entities validated against schema"
echo "  - T094: Audit log captures events"
echo "  - T095: SC-005 - 3+ candidates identified"
echo "  - T096: SC-006 - 1+ type successfully promoted"
echo "  - T097: SC-010 - Audit log complete"
echo ""

US3_OUTPUT=$(go test -v ./tests/integration -run "TestAcceptance.*T09[0-7]" -count=1 2>&1)
US3_RESULT=$?

if [ $US3_RESULT -eq 0 ]; then
    US3_PASSED=$(echo "$US3_OUTPUT" | grep -c "^--- PASS:" || true)
    US3_SKIPPED=$(echo "$US3_OUTPUT" | grep -c "^--- SKIP:" || true)
    echo -e "${GREEN}✓ User Story 3: PASSED${NC} ($US3_PASSED tests passed, $US3_SKIPPED skipped)"
    PASSED=$((PASSED + US3_PASSED))
    SKIPPED=$((SKIPPED + US3_SKIPPED))
else
    echo -e "${RED}✗ User Story 3: FAILED${NC}"
    FAILED=$((FAILED + 1))
fi
echo ""

echo "=========================================="
echo "Summary"
echo "=========================================="
echo -e "User Stories Passed: ${GREEN}3/3${NC}"
echo -e "Total Tests Passed:  ${GREEN}$PASSED${NC}"
echo -e "Total Tests Skipped: ${YELLOW}$SKIPPED${NC}"
echo -e "Total Tests Failed:  ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ T142 PASS: All P1-P3 user stories pass acceptance scenarios${NC}"
    exit 0
else
    echo -e "${RED}✗ T142 FAIL: Some user stories failed acceptance scenarios${NC}"
    exit 1
fi
