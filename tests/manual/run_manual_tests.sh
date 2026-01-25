#!/bin/bash
# T108 & T109 Manual Test Execution Guide
# This script provides step-by-step instructions for manual testing

set -e

REPO_ROOT="/Users/jochem/code/enron-graph-2"
cd "$REPO_ROOT"

echo "======================================================================"
echo "T108 & T109: TUI Manual Test Execution Guide"
echo "======================================================================"
echo ""

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
echo -e "${YELLOW}[1] Checking Prerequisites...${NC}"
echo ""

# Check if database is running
if docker ps | grep -q enron-graph-postgres; then
    echo -e "${GREEN}✓${NC} PostgreSQL database is running"
else
    echo -e "${RED}✗${NC} PostgreSQL database is NOT running"
    echo "  Starting database..."
    docker-compose up -d postgres
    sleep 5
fi

# Check if entities exist
ENTITY_COUNT=$(docker exec enron-graph-postgres psql -U enron -d enron_graph -t -c "SELECT COUNT(*) FROM discovered_entities;" | tr -d ' ')
echo -e "${GREEN}✓${NC} Database has $ENTITY_COUNT entities"

# Check if TUI binary exists
if [ -f "./tui" ]; then
    echo -e "${GREEN}✓${NC} TUI binary exists ($(ls -lh tui | awk '{print $5}'))"
else
    echo -e "${YELLOW}⚠${NC}  TUI binary not found, building..."
    go build -o tui cmd/tui/main.go
    echo -e "${GREEN}✓${NC} TUI binary built successfully"
fi

echo ""
echo "======================================================================"
echo -e "${GREEN}Prerequisites met! Ready for manual testing.${NC}"
echo "======================================================================"
echo ""

# T108 Testing Instructions
echo -e "${YELLOW}[T108] TUI Navigation Flows - Manual Testing Instructions${NC}"
echo ""
echo "Follow these steps and document results in:"
echo "  tests/manual/T108_TUI_NAVIGATION_FLOWS.md"
echo ""
echo "1. Launch TUI:"
echo "   ./tui"
echo ""
echo "2. Test Cases to Execute:"
echo "   - TC1: Navigate entity list with arrow keys"
echo "   - TC2: Filter entities by type"
echo "   - TC3: Search entities by name"
echo "   - TC4: View entity details"
echo "   - TC5: Visualize entity as graph"
echo "   - TC6: Navigate and expand graph nodes"
echo ""
echo "3. Document observations in test document"
echo ""
echo "Press ENTER when ready to view T109 instructions..."
read -r

# T109 Testing Instructions
echo ""
echo "======================================================================"
echo -e "${YELLOW}[T109] Error Handling - Manual Testing Instructions${NC}"
echo ""
echo "Follow these steps and document results in:"
echo "  tests/manual/T109_ERROR_HANDLING.md"
echo ""
echo "TEST 1: Empty Database"
echo "  # Backup and clear data"
echo "  docker exec enron-graph-postgres psql -U enron -d enron_graph -c 'TRUNCATE discovered_entities, relationships CASCADE;'"
echo "  ./tui"
echo "  # Document behavior with empty database"
echo "  # Restore data when done"
echo ""
echo "TEST 2: Network Disconnection During Session"
echo "  ./tui"
echo "  # In another terminal, stop database:"
echo "  docker-compose stop postgres"
echo "  # Try navigating in TUI, document behavior"
echo "  # Restart database:"
echo "  docker-compose up -d postgres"
echo ""
echo "TEST 3: Network Disconnection on Startup"
echo "  docker-compose stop postgres"
echo "  ./tui"
echo "  # Document error messages and behavior"
echo "  docker-compose up -d postgres"
echo ""
echo "TEST 4: Invalid Entity Selection"
echo "  ./tui"
echo "  # Test boundary conditions (first/last entity navigation)"
echo "  # Delete entity while viewing details"
echo "  # Test filters with no results"
echo ""
echo "TEST 5: Malformed Data"
echo "  # Insert entities with NULL/invalid data"
echo "  docker exec enron-graph-postgres psql -U enron -d enron_graph -c \"INSERT INTO discovered_entities (name, type, confidence) VALUES (NULL, 'PERSON', 0.5);\""
echo "  ./tui"
echo "  # Document how TUI handles malformed data"
echo ""
echo "======================================================================"
echo ""

echo -e "${GREEN}Manual testing guide complete.${NC}"
echo ""
echo "Next steps:"
echo "  1. Execute T108 test cases and fill in: tests/manual/T108_TUI_NAVIGATION_FLOWS.md"
echo "  2. Execute T109 test cases and fill in: tests/manual/T109_ERROR_HANDLING.md"
echo "  3. Mark tasks as complete in tasks.md"
echo ""
echo "To launch TUI now, run:"
echo "  ./tui"
echo ""
