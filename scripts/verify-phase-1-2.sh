#!/bin/bash
# Verification script for Graph Explorer Phase 1 & 2
# Verifies all tasks T001-T015 are complete

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Graph Explorer Phase 1 & 2 Verification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

PASS=0
FAIL=0

check() {
    local test_name="$1"
    local test_cmd="$2"
    
    if eval "$test_cmd" &>/dev/null; then
        echo "✅ $test_name"
        ((PASS++))
        return 0
    else
        echo "❌ $test_name"
        ((FAIL++))
        return 1
    fi
}

check_file() {
    check "$2" "[ -f '$1' ]"
}

check_dir() {
    check "$2" "[ -d '$1' ]"
}

check_content() {
    local file="$1"
    local pattern="$2"
    local desc="$3"
    check "$desc" "grep -q '$pattern' '$file'"
}

echo "PHASE 1: Setup (T001-T007)"
echo "─────────────────────────────────────────────────────────"

# T001: Wails project initialized
check_file "cmd/explorer/wails.json" "T001: Wails project structure exists"
check_file "cmd/explorer/main.go" "T001: Wails main.go exists"
check_file "cmd/explorer/app.go" "T001: Wails app.go exists"

# T002: Monorepo structure adapted
check_dir "cmd/explorer" "T002: cmd/explorer directory exists"
check_dir "frontend" "T002: frontend directory exists"
check_content "cmd/explorer/wails.json" "frontend:install" "T002: wails.json references frontend/"

# T003: Frontend dependencies installed
check_file "frontend/package.json" "T003: package.json exists"
check_content "frontend/package.json" "react-force-graph" "T003: react-force-graph dependency"
check_content "frontend/package.json" "three" "T003: three.js dependency"
check_content "frontend/package.json" "@types/three" "T003: @types/three dependency"
check_dir "frontend/node_modules" "T003: node_modules installed"

# T004: Wails build configured
check_content "cmd/explorer/wails.json" "frontend:build" "T004: frontend:build command configured"
check_content "cmd/explorer/wails.json" "wailsjsdir" "T004: wailsjsdir configured"

# T005: internal/explorer directory created
check_dir "internal/explorer" "T005: internal/explorer directory exists"

# T006: .gitignore updated
check_content ".gitignore" "frontend/dist" "T006: frontend/dist in .gitignore"
check_content ".gitignore" "frontend/node_modules" "T006: frontend/node_modules in .gitignore"
check_content ".gitignore" "frontend/wailsjs" "T006: frontend/wailsjs in .gitignore"

# T007: Wails can build (check for wails in go.mod)
check_content "go.mod" "github.com/wailsapp/wails/v2" "T007: Wails dependency in go.mod"

echo ""
echo "PHASE 2: Foundation (T008-T015)"
echo "─────────────────────────────────────────────────────────"

# T008-T011: DTO models in models.go
check_file "internal/explorer/models.go" "T008-T011: models.go exists"
check_content "internal/explorer/models.go" "type GraphNode struct" "T008: GraphNode DTO defined"
check_content "internal/explorer/models.go" "type GraphEdge struct" "T008: GraphEdge DTO defined"
check_content "internal/explorer/models.go" "type SchemaType struct" "T008: SchemaType DTO defined"
check_content "internal/explorer/models.go" "type PropertyDefinition struct" "T008: PropertyDefinition DTO defined"
check_content "internal/explorer/models.go" "type GraphResponse struct" "T009: GraphResponse DTO defined"
check_content "internal/explorer/models.go" "type RelationshipsResponse struct" "T009: RelationshipsResponse DTO defined"
check_content "internal/explorer/models.go" "type SchemaResponse struct" "T010: SchemaResponse DTO defined"
check_content "internal/explorer/models.go" "type NodeFilter struct" "T011: NodeFilter DTO defined"

# T012: Database connection setup
check_content "cmd/explorer/main.go" "ent.Open" "T012: Database connection in main.go"
check_content "cmd/explorer/main.go" "utils.LoadConfig" "T012: Config loading in main.go"
check_content "cmd/explorer/app.go" "client \*ent.Client" "T012: ent.Client in App struct"

# T013: Test directories created
check_dir "tests/contract" "T013: tests/contract directory exists"
check_dir "tests/integration/explorer" "T013: tests/integration/explorer directory exists"

# T014: Test helper setup
check_file "tests/contract/test_helper.go" "T014: test_helper.go exists"
check_content "tests/contract/test_helper.go" "func TestClient" "T014: TestClient helper function"
check_content "tests/contract/test_helper.go" "func CleanupDB" "T014: CleanupDB helper function"
check_content "tests/contract/test_helper.go" "func SeedTestData" "T014: SeedTestData helper function"

# T015: TypeScript bindings generated
check_dir "frontend/src/wailsjs" "T015: wailsjs directory exists"
check_file "frontend/src/wailsjs/runtime/runtime.js" "T015: Wails runtime generated"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Results"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  ✅ Passed: $PASS"
echo "  ❌ Failed: $FAIL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "🎉 SUCCESS: Phase 1 and Phase 2 are complete!"
    echo ""
    echo "You can now proceed to Phase 3: User Story 1 Implementation"
    exit 0
else
    echo "⚠️  WARNING: Some checks failed. Please review the output above."
    exit 1
fi
