# Implementation Plan: Graph Explorer

**Branch**: `002-graph-explorer` | **Date**: 2026-01-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-graph-explorer/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

The Graph Explorer feature provides a graphical interface for exploring the Enron knowledge graph, displaying both promoted schema types (Email, Relationship, SchemaPromotion) and dynamically discovered entities. The explorer will use a hybrid web-local architecture with Wails (Go backend + web frontend in a single binary), featuring a force-directed graph visualization (auto-loading 100 nodes on startup) with interactive navigation, schema browsing, filtering, and detailed entity inspection capabilities. Newly expanded nodes show relationship counts but require explicit clicks to expand further. Filters show ghost node placeholders for relationships extending beyond filtered types.

## Technical Context

**Language/Version**: Go 1.25.3  
**Primary Dependencies**: Wails v2 (Go + webview bridge), ent (existing ORM), React + react-force-graph (frontend)  
**Storage**: PostgreSQL 15+ with existing ent schema (Email, Relationship, SchemaPromotion, DiscoveredEntity)  
**Testing**: Go standard testing, testify (existing), React Testing Library (frontend)  
**Target Platform**: macOS/Linux/Windows desktop (Wails produces single binary per platform)  
**Project Type**: Hybrid desktop application (Go backend + web UI)  
**Performance Goals**: Render 100-500 nodes smoothly (FR-003a), support up to 1000 nodes with <500ms pan/zoom (SC-008)  
**Constraints**: Force-directed layout required (FR-003b), batch loading for >50 relationships (FR-006a), <2 second load times (SC-001, SC-006)  
**Scale/Scope**: Single-user desktop application, reads from existing PostgreSQL database, no write operations in v1

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check (Pre-Phase 0)

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Specification-First | ✅ PASS | Complete spec.md exists with user scenarios, FR-XXX requirements, success criteria, and edge cases |
| II. Independent User Stories | ✅ PASS | 4 user stories (US1-US4) are independently testable with P1-P2 priorities; each delivers standalone value |
| III. Test-Driven Development | ✅ PASS | TDD will be applied (contract tests for Go services, component tests for React, integration tests for Wails bridge) |
| IV. Documentation-Driven Design | ⏳ PENDING | This plan.md in progress; research.md, data-model.md, contracts/, quickstart.md to be generated in Phases 0-1 |
| V. Complexity Justification | ⚠️ REVIEW | Adding Wails + React frontend to Go-only project - see Complexity Tracking table below |
| VI. Measurable Success Criteria | ✅ PASS | 8 success criteria (SC-001 through SC-008) with specific metrics (time limits, node counts, performance targets) |
| VII. File Editing Discipline | ✅ ACKNOWLEDGED | Will use replace_string_in_file for edits, create_file only for new files |
| VIII. Terminal Command Discipline | ✅ ACKNOWLEDGED | Will avoid multi-line terminal commands, use file creation for complex inputs |
| IX. Commit Confirmation | ✅ ACKNOWLEDGED | Will request user approval before any git commits |

**Gate Decision**: ⚠️ **CONDITIONAL PASS** - Proceed to Phase 0 research. Complexity justification required for hybrid architecture (documented in Complexity Tracking).

### Post-Phase 1 Re-evaluation

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Specification-First | ✅ PASS | No changes from initial check |
| II. Independent User Stories | ✅ PASS | No changes; all stories remain independently testable |
| III. Test-Driven Development | ✅ PASS | Contract test specs defined in contracts/; TDD workflow documented in quickstart.md |
| IV. Documentation-Driven Design | ✅ PASS | **NOW COMPLETE**: research.md (8 decisions), data-model.md (7 entities), contracts/ (3 interfaces), quickstart.md (dev workflow) |
| V. Complexity Justification | ✅ PASS | **JUSTIFIED**: Wails+React complexity documented in Complexity Tracking; no simpler alternative meets spec requirements |
| VI. Measurable Success Criteria | ✅ PASS | No changes; criteria remain measurable and achievable with chosen stack |
| VII. File Editing Discipline | ✅ ACKNOWLEDGED | No changes |
| VIII. Terminal Command Discipline | ✅ ACKNOWLEDGED | No changes |
| IX. Commit Confirmation | ✅ ACKNOWLEDGED | No changes |

**Final Gate Decision**: ✅ **FULL PASS** - All constitution principles satisfied. Ready to proceed to Phase 2 (tasks.md generation via `/speckit.tasks`).

## Project Structure

### Documentation (this feature)

```text
specs/002-graph-explorer/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── graph-service.go.md   # Go interface for graph data access
│   └── frontend-api.ts.md    # TypeScript types for Wails bindings
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
├── explorer/            # NEW: Wails application entry point
│   └── main.go

internal/
├── explorer/            # NEW: Graph explorer backend services
│   ├── graph_service.go      # Unified graph data provider
│   ├── graph_service_test.go
│   ├── schema_service.go     # Schema introspection (promoted + discovered)
│   ├── schema_service_test.go
│   └── models.go             # Common DTOs (GraphNode, GraphEdge, SchemaType)
└── graph/               # EXISTING: Repository already has graph queries
    └── repository.go    # REUSE: Existing graph traversal logic

frontend/                # NEW: React application for Wails
├── src/
│   ├── components/
│   │   ├── GraphCanvas.tsx         # Force-directed graph visualization
│   │   ├── SchemaPanel.tsx         # Schema browser (left sidebar)
│   │   ├── DetailPanel.tsx         # Entity detail viewer (right sidebar)
│   │   ├── FilterBar.tsx           # Type and search filters
│   │   └── LoadMoreButton.tsx      # Batch loading control
│   ├── services/
│   │   └── wails.ts               # Wails Go function bindings
│   ├── types/
│   │   └── graph.ts               # TypeScript types matching Go models
│   ├── App.tsx
│   └── main.tsx
├── package.json
└── vite.config.ts

tests/
├── integration/
│   └── explorer/        # NEW: Wails integration tests
│       └── graph_explorer_test.go
└── contract/            # NEW: Contract tests for graph service
    └── graph_service_contract_test.go
```

**Structure Decision**: Hybrid architecture using Wails. The `cmd/explorer` entry point creates a desktop application embedding the React frontend. Backend logic reuses existing `internal/graph` repository and adds new `internal/explorer` services. This is a new project within the existing monorepo, following the established `cmd/` + `internal/` pattern.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Hybrid Wails architecture (Go + React frontend) | FR-003b requires force-directed layout with WebGL performance (SC-008: 1000 nodes <500ms). Terminal-only solutions (bubbletea/termui) cannot render physics-based layouts or provide smooth pan/zoom for 1000+ nodes. | Pure TUI (bubbletea): Cannot achieve force-directed physics layout or handle 1000-node rendering smoothly. SVG-based web libraries: Cannot meet SC-008 performance target. CLI-only: Violates core requirement for "graphical interface" (spec input). |
| React + react-force-graph library | Spec explicitly requires force-directed layout (Clarification Q2), directional arrows (Q3), and high-performance rendering (SC-008). `react-force-graph` is the leading solution using Three.js/WebGL, proven to handle 10k+ nodes. | D3.js force layout: Requires significant custom code for 3D/WebGL, batching, and arrow rendering that react-force-graph provides out-of-box. Go-native graph rendering: No mature force-directed libraries exist; would require months of physics engine development. |
| New `frontend/` directory | Wails requires a separate frontend project with its own build toolchain (Vite/npm). Cannot embed React components directly in Go without a build step and Wails bridge. | Serve HTML from Go templates: Cannot achieve force-directed layout or meet performance requirements without a real framework and WebGL support. |

---

## Phase Completion Summary

### ✅ Phase 0: Research - COMPLETE

**Output**: [research.md](./research.md)

**Key Decisions**: Wails v2, react-force-graph, Ent Schema API, OFFSET/LIMIT pagination, three-layer testing, TypeScript generation, Vite integration, schema caching

**Unknowns Resolved**: All 8 research questions answered. No blockers.

### ✅ Phase 1: Design & Contracts - COMPLETE

**Outputs**: data-model.md (7 entities), contracts/ (3 service interfaces), quickstart.md (dev workflow)

**Agent Context**: Updated with Wails, React, react-force-graph

**Constitution Re-check**: All principles satisfied ✅

---

## Next Steps

Specification has been clarified with 7 Q&A items covering expansion behavior (FR-006b) and filter edge handling (FR-007a). Tasks.md has been generated with 122 tasks organized by user story.

**Ready for**: `/speckit.implement` - Begin TDD implementation starting with Phase 1 (Setup) and Phase 2 (Foundation)

## Plan Status: ✅ COMPLETE

**Ready for**: Implementation (all design artifacts complete, spec clarified, tasks broken down)
