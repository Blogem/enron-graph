# Graph Explorer - Quick Start Guide

**Feature**: 002-graph-explorer  
**Last Updated**: 2026-01-26

## Prerequisites

Before starting development on the Graph Explorer, ensure you have:

- **Go 1.25.3+** installed
- **Node.js 18+** and **npm** installed  
- **Wails CLI** installed: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **PostgreSQL 15+** with existing Enron graph database running
- **Existing enron-graph-2 repository** cloned and dependencies installed

## Project Structure

```
enron-graph-2/
├── cmd/
│   └── explorer/              # NEW: Wails application entry point
│       └── main.go
├── internal/
│   └── explorer/              # NEW: Backend services
│       ├── graph_service.go
│       ├── schema_service.go
│       └── models.go
├── frontend/                  # NEW: React application
│   ├── src/
│   │   ├── components/
│   │   ├── types/
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   └── vite.config.ts
└── specs/002-graph-explorer/  # THIS DIRECTORY
    ├── spec.md
    ├── plan.md
    ├── research.md
    ├── data-model.md
    ├── contracts/
    └── quickstart.md (you are here)
```

---

## Development Workflow

### Initial Setup (First Time Only)

1. **Install Wails CLI** (if not already installed):
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

2. **Verify Wails installation**:
   ```bash
   wails doctor
   ```
   This checks that you have all required dependencies (Go, Node.js, npm).

3. **Initialize Wails project** (adapt to monorepo):
   ```bash
   # From repository root
   wails init -n explorer -t react-ts
   
   # This creates a template structure - we'll adapt it to fit our monorepo
   # Move files to match our project structure (cmd/explorer/, frontend/)
   ```

4. **Install frontend dependencies**:
   ```bash
   cd frontend
   npm install react-force-graph three
   npm install -D @types/three
   cd ..
   ```

---

### Daily Development

#### Option 1: Wails Dev Mode (Recommended)

Start the application with hot reload for both backend and frontend:

```bash
# From repository root
cd cmd/explorer
wails dev
```

This will:
- Start the Go backend with hot reload
- Start the Vite dev server for React frontend
- Open the Wails app window
- Auto-reload on file changes

The app will be available in a native window, with DevTools accessible via right-click → Inspect.

#### Option 2: Separate Backend/Frontend (for debugging)

Terminal 1 - Backend:
```bash
cd cmd/explorer
go run main.go --dev
```

Terminal 2 - Frontend:
```bash
cd frontend
npm run dev
```

Then open browser to `http://localhost:5173` (Vite default port).

---

### Building for Production

Create a production-ready binary:

```bash
# From repository root
cd cmd/explorer
wails build

# Output: build/bin/explorer (or explorer.exe on Windows, explorer.app on macOS)
```

The binary includes:
- Compiled Go backend
- Bundled React frontend assets
- Embedded webview

Run the binary:
```bash
./build/bin/explorer
```

---

## Implementation Order (by User Story)

Follow Test-Driven Development (Constitution Principle III):

### Phase 1: User Story 1 - View Schema Overview (P1)

**Goal**: Display promoted and discovered entity types in schema panel.

1. **Write contract tests** for `SchemaService`:
   ```bash
   # Create tests/contract/schema_service_contract_test.go
   # Run: go test ./tests/contract -v
   ```

2. **Implement SchemaService**:
   - Create `internal/explorer/schema_service.go`
   - Implement `GetSchema()` using ent schema introspection
   - Query `DiscoveredEntity` table for discovered types
   - Verify contract tests pass

3. **Create frontend SchemaPanel component**:
   ```bash
   # Create frontend/src/components/SchemaPanel.tsx
   # Write component tests: frontend/src/components/SchemaPanel.test.tsx
   # Run: npm test
   ```

4. **Wire up Wails binding**:
   - Update `cmd/explorer/main.go` to bind `SchemaService`
   - Generate TypeScript types: `cd frontend && wails generate module`
   - Call `GetSchema()` from React in `App.tsx`

5. **Verify acceptance scenarios** (User Story 1):
   - Open explorer → see promoted types (Email, Relationship, etc.)
   - See discovered types listed separately
   - Click type → see property details
   - Refresh → see new discovered types

---

### Phase 2: User Story 2 - Browse Graph Visually (P1)

**Goal**: Render force-directed graph with 50-100 random nodes on startup.

1. **Write contract tests** for `GraphService`:
   ```bash
   # tests/contract/graph_service_contract_test.go
   # Focus on: GetRandomNodes(), GetRelationships() with batching
   ```

2. **Implement GraphService**:
   - Create `internal/explorer/graph_service.go`
   - Implement `GetRandomNodes(limit int)`
   - Implement `GetRelationships(nodeID, offset, limit)` for batching
   - Verify contract tests pass

3. **Create GraphCanvas component**:
   ```bash
   # frontend/src/components/GraphCanvas.tsx
   # Use react-force-graph library
   # Configure force-directed layout (d3-force)
   # Add directional arrows on edges
   ```

4. **Implement node expansion** ("Load 50 more" button):
   - Create `LoadMoreButton.tsx` component
   - Track `expandedNodes` state in parent component
   - Call `GetRelationships()` on button click

5. **Verify acceptance scenarios** (User Story 2):
   - App starts → see 50-100 nodes automatically loaded
   - Click node → see properties
   - Click expand → see connected nodes added
   - Expand node with >50 relationships → see "Load 50 more" button
   - Pan and zoom → smooth interaction

---

### Phase 3: User Story 3 - Filter and Search (P2)

**Goal**: Filter by entity type and search by property values.

1. **Extend GraphService**:
   - Implement `GetNodes(filter NodeFilter)`
   - Add tests for type filtering and search

2. **Create FilterBar component**:
   - Type checkboxes (Email, Person, etc.)
   - Category filter (promoted/discovered)
   - Search input box
   - Call `GetNodes()` on filter change

3. **Verify acceptance scenarios** (User Story 3):
   - Filter to "Email" only → see only Email nodes
   - Search "john@enron.com" → see matching nodes highlighted
   - Clear filters → restore full graph

---

### Phase 4: User Story 4 - Navigate Entity Details (P2)

**Goal**: Show detail panel with all properties for selected node.

1. **Create DetailPanel component**:
   - Display all node properties
   - Show relationship list with types
   - For discovered entities: show confidence, discovery timestamp

2. **Wire up node selection**:
   - GraphCanvas `onNodeClick` → set selected node
   - DetailPanel receives selected node as prop
   - Call `GetNodeDetails(nodeID)` for complete data

3. **Verify acceptance scenarios** (User Story 4):
   - Click node → detail panel opens
   - See all properties displayed
   - See connected relationships listed
   - Discovered entity → see metadata

---

## Testing Strategy

### Contract Tests (Go)

```bash
# Run all contract tests
go test ./tests/contract -v

# Run specific test
go test ./tests/contract -run TestGraphService_GetRandomNodes -v
```

### Component Tests (React)

```bash
# Run all frontend tests
cd frontend
npm test

# Run specific test file
npm test SchemaPanel.test.tsx

# Run with coverage
npm test -- --coverage
```

### Integration Tests (Wails)

```bash
# Run integration tests (requires Wails app running)
go test ./tests/integration/explorer -v
```

---

## Common Commands

### Backend Development

```bash
# Run Go tests
go test ./internal/explorer -v

# Run contract tests
go test ./tests/contract -v

# Format code
go fmt ./...

# Lint
golangci-lint run
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Run dev server (standalone)
npm run dev

# Run tests
npm test

# Type check
npm run type-check

# Lint
npm run lint

# Build (creates dist/ folder)
npm run build
```

### Wails Commands

```bash
# Start dev mode
wails dev

# Build production binary
wails build

# Generate TypeScript bindings from Go
wails generate module

# Check environment
wails doctor
```

---

## Troubleshooting

### Wails app won't start

**Problem**: `wails dev` fails with "command not found"

**Solution**: Ensure Wails CLI is installed and in PATH:
```bash
which wails
# If not found:
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

---

### Frontend can't call Go functions

**Problem**: TypeScript error: "Cannot find module '../wailsjs/go/explorer/GraphService'"

**Solution**: Generate TypeScript bindings:
```bash
cd frontend
wails generate module
```

---

### Database connection errors

**Problem**: "could not connect to database"

**Solution**: Ensure PostgreSQL is running and connection string is correct:
```bash
# Check database is running
docker ps | grep postgres

# Check connection string in Go code matches docker-compose.yml
# Default: postgres://postgres:postgres@localhost:5432/enron_graph?sslmode=disable
```

---

### Force-graph not rendering

**Problem**: Blank screen in GraphCanvas component

**Solution**: 
1. Check browser console for errors
2. Verify `react-force-graph` is installed: `npm list react-force-graph`
3. Ensure data format matches library expectations (nodes[], links[])
4. Check that nodes have `id` field and links have `source`, `target` fields

---

## Environment Variables

Create a `.env` file in repository root:

```bash
# Database connection
DATABASE_URL=postgres://postgres:postgres@localhost:5432/enron_graph?sslmode=disable

# Development mode
DEV_MODE=true

# Wails dev server port (default: 34115)
WAILS_PORT=34115
```

---

## Performance Optimization

### If graph rendering is slow (>500ms pan/zoom):

1. **Reduce initial node count**:
   - Change `GetRandomNodes(50)` instead of 100
   - Verify with SC-002: "100-500 nodes smooth"

2. **Enable WebGL in react-force-graph**:
   - Use `ForceGraph3D` instead of `ForceGraph2D` for better performance
   - Check docs: https://github.com/vasturiano/react-force-graph

3. **Optimize force simulation**:
   ```typescript
   <ForceGraph
     d3AlphaDecay={0.02}  // Faster stabilization
     d3VelocityDecay={0.3}  // More damping
     warmupTicks={100}  // Pre-calculate layout
   />
   ```

---

## Next Steps

After implementing all user stories:

1. **Run full test suite**:
   ```bash
   go test ./... -v
   cd frontend && npm test
   ```

2. **Verify all acceptance scenarios** pass (see spec.md)

3. **Check success criteria** (SC-001 through SC-008)

4. **Build production binary**:
   ```bash
   wails build
   ```

5. **Test production build** on target platforms (macOS/Linux/Windows)

6. **Document any deviations** from plan in tasks.md

7. **Request user approval** before committing (Constitution Principle IX)

---

## Resources

- **Wails Documentation**: https://wails.io/docs/introduction
- **react-force-graph**: https://github.com/vasturiano/react-force-graph
- **ent Documentation**: https://entgo.io/docs/getting-started
- **Specification**: [../spec.md](../spec.md)
- **Implementation Plan**: [../plan.md](../plan.md)
- **Data Model**: [../data-model.md](../data-model.md)
- **Contracts**: [../contracts/](../contracts/)
