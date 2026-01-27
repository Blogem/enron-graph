# Implementation Plan: Graph Explorer Chat Interface

**Branch**: `004-explorer-chat` | **Date**: 2026-01-27 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-explorer-chat/spec.md`

## Summary

Add a chat interface panel to the graph explorer GUI that integrates with the existing `internal/chat` package. The chat panel will be positioned as a bottom panel below the graph visualization, allowing users to query the knowledge graph using natural language. The implementation reuses the existing chat.Handler.ProcessQuery method for all query processing logic, focusing solely on the UI integration layer.

## Technical Context

**Language/Version**: Go 1.21+ (backend), TypeScript/React 18+ (frontend)
**Primary Dependencies**: 
- Backend: Wails v2 (Go-React bridge), existing internal/chat package
- Frontend: React 18, Vite, existing Wails bindings
**Storage**: N/A (conversation state managed in-memory, reuses existing chat.Context)
**Testing**: Go testing package (backend), React Testing Library (frontend)
**Target Platform**: Desktop application (macOS, Windows, Linux) via Wails
**Project Type**: Desktop GUI application with Go backend and React frontend
**Performance Goals**: Query responses < 3 seconds for 95% of queries, UI remains responsive with 50+ messages
**Constraints**: Must integrate with existing cmd/explorer application, no modifications to internal/chat package allowed
**Scale/Scope**: Single-user desktop application, ~5 new React components, ~2 new Go methods in App struct

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Specification-First Development ✅
- Complete specification exists at `specs/004-explorer-chat/spec.md`
- Includes user scenarios with Given-When-Then acceptance criteria
- Functional requirements defined (FR-001 through FR-024)
- Success criteria with measurable outcomes (SC-001 through SC-008)
- Edge cases documented and clarified

### Principle II: Independent User Stories ✅
- 4 user stories defined with priorities (P1-P3)
- Each story is independently testable:
  - US1: Display chat interface (P1)
  - US2: Send queries and display responses (P1)
  - US3: Conversation history (P2)
  - US4: Clear conversation (P3)
- Each delivers standalone value

### Principle III: Test-Driven Development ✅
- Plan includes test strategy for both frontend and backend
- Tests will be written before implementation (Red-Green-Refactor)
- Contract tests for Wails bindings, component tests for React UI

### Principle IV: Documentation-Driven Design ✅
- This plan.md exists
- Phase 0 will generate research.md (integration patterns, Wails bindings)
- Phase 1 will generate data-model.md (ChatMessage, ConversationSession entities)
- Phase 1 will generate contracts/ (TypeScript/Go interface contracts)
- Phase 1 will generate quickstart.md (developer setup)

### Principle V: Complexity Justification ✅
- No complexity violations
- Solution reuses existing patterns from cmd/explorer
- No new abstractions introduced beyond standard React component patterns
- Direct integration with existing internal/chat package

### Principle VI: Measurable Success Criteria ✅
- All success criteria are technology-agnostic
- Measurable outcomes defined (response time, message count, user actions)
- Testable through automated tests and manual verification

### Principle VII: File Editing Discipline ✅
- Plan documented to use replace_string_in_file for existing files
- create_file only for genuinely new components
- Verification steps included in implementation tasks

### Principle VIII: Terminal Command Discipline ✅
- No complex multi-line commands planned
- File creation via tools, simple commands for execution
- Git operations split into atomic commands

### Principle IX: Commit Confirmation ✅
- Plan requires user confirmation before commits
- Task completion will be summarized before requesting commit approval

**GATE STATUS**: ✅ PASS - All constitutional principles satisfied, proceed to Phase 0

## Project Structure

### Documentation (this feature)

```text
specs/004-explorer-chat/
├── plan.md              # This file
├── research.md          # Phase 0: Integration patterns, Wails binding research
├── data-model.md        # Phase 1: ChatMessage, ConversationSession entities
├── quickstart.md        # Phase 1: Developer setup and testing guide
├── contracts/           # Phase 1: TypeScript/Go interface definitions
│   ├── chat-api.ts      # Frontend TypeScript types for chat API
│   └── chat-api.go      # Backend Go method signatures
├── checklists/
│   └── requirements.md  # Existing spec validation checklist
└── tasks.md             # Phase 2: Granular implementation tasks (created by /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/explorer/
├── app.go                          # [MODIFY] Add ProcessChatQuery, ClearChatContext methods
├── main.go                         # [MODIFY] Initialize chat handler and dependencies
├── frontend/
│   └── src/
│       ├── App.tsx                 # [MODIFY] Add chat panel state and layout
│       ├── App.css                 # [MODIFY] Add bottom panel layout styles
│       ├── components/
│       │   ├── ChatPanel.tsx       # [NEW] Main chat panel component
│       │   ├── ChatPanel.css       # [NEW] Chat panel styles
│       │   ├── ChatMessage.tsx     # [NEW] Individual message component
│       │   ├── ChatMessage.css     # [NEW] Message bubble styles
│       │   ├── ChatInput.tsx       # [NEW] Input field with submit button
│       │   └── ChatInput.css       # [NEW] Input field styles
│       ├── services/
│       │   └── chat.ts             # [NEW] Chat API service wrapper
│       └── types/
│           └── chat.ts             # [NEW] TypeScript types for chat

internal/chat/
└── [NO MODIFICATIONS]              # Reuse existing Handler, Context, types

tests/
├── contract/
│   └── chat_bindings_test.go       # [NEW] Contract tests for Wails bindings
└── frontend/
    └── src/
        └── components/
            ├── ChatPanel.test.tsx  # [NEW] ChatPanel component tests
            ├── ChatMessage.test.tsx # [NEW] ChatMessage component tests
            └── ChatInput.test.tsx   # [NEW] ChatInput component tests
```

**Structure Decision**: This feature extends the existing cmd/explorer Wails application by:
1. Adding new Wails-bound Go methods to App struct (app.go) for chat operations
2. Creating new React components in frontend/src/components/ following existing patterns
3. Maintaining separation between UI (cmd/explorer) and business logic (internal/chat)
4. No modifications to internal/chat package - pure integration layer

## Complexity Tracking

> **No violations** - This feature follows existing patterns and introduces no additional complexity.

The implementation:
- Reuses existing chat.Handler from internal/chat (no new business logic layer)
- Follows established Wails binding patterns from cmd/explorer/app.go
- Uses standard React component composition matching existing components (SchemaPanel, DetailPanel)
- Maintains existing architecture: Go backend for data/logic, React frontend for UI
- No new abstractions, frameworks, or architectural patterns introduced
