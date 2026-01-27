---

description: "Task breakdown for Graph Explorer Chat Interface implementation"
---

# Tasks: Graph Explorer Chat Interface

**Input**: Design documents from `/specs/004-explorer-chat/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: Tests are NOT explicitly requested in the spec but are included following TDD best practices from quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure verification

- [x] T001 Verify Wails v2 development environment is working (run `wails doctor` from cmd/explorer/)
- [x] T002 [P] Install frontend dependencies in cmd/explorer/frontend/
- [x] T003 [P] Create TypeScript types file at cmd/explorer/frontend/src/types/chat.ts from contracts/chat-api.ts

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Backend infrastructure that MUST be complete before ANY user story UI can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Create stub LLM client implementation (internal/chat/stub_llm.go) implementing chat.LLMClient interface
- [x] T005 Create chat adapter implementation (cmd/explorer/chat_adapter.go) implementing chat.Repository interface using ent client
- [x] T006 Add chat handler initialization to App struct in cmd/explorer/main.go (NewApp function)
- [x] T007 Write contract test for ProcessChatQuery in tests/contract/chat_bindings_test.go (RED phase)
- [x] T008 Implement ProcessChatQuery method in cmd/explorer/app.go with validation and timeout
- [x] T009 Implement ClearChatContext method in cmd/explorer/app.go
- [x] T010 Run contract tests to verify backend integration (GREEN phase)

**Checkpoint**: Foundation ready - backend methods callable from frontend, all tests passing

---

## Phase 3: User Story 1 - Display Chat Interface (Priority: P1) 🎯 MVP

**Goal**: Users can see a chat interface panel positioned as a bottom panel below the graph visualization, with text input field and conversation area. Users can collapse/expand the panel.

**Independent Test**: Open graph explorer, verify chat panel is visible at bottom, can accept input, can be collapsed/expanded

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T011 [P] [US1] Write ChatInput component test in cmd/explorer/frontend/src/components/ChatInput.test.tsx
- [x] T012 [P] [US1] Write ChatMessage component test in cmd/explorer/frontend/src/components/ChatMessage.test.tsx
- [x] T013 [US1] Write ChatPanel component test in cmd/explorer/frontend/src/components/ChatPanel.test.tsx (depends on T011, T012 for understanding component contracts)

### Implementation for User Story 1

- [X] T014 [P] [US1] Create ChatInput component in cmd/explorer/frontend/src/components/ChatInput.tsx
- [X] T015 [P] [US1] Create ChatInput styles in cmd/explorer/frontend/src/components/ChatInput.css
- [X] T016 [P] [US1] Create ChatMessage component in cmd/explorer/frontend/src/components/ChatMessage.tsx
- [X] T017 [P] [US1] Create ChatMessage styles in cmd/explorer/frontend/src/components/ChatMessage.css
- [X] T018 [US1] Create ChatPanel component in cmd/explorer/frontend/src/components/ChatPanel.tsx
- [X] T019 [US1] Create ChatPanel styles in cmd/explorer/frontend/src/components/ChatPanel.css
- [X] T020 [US1] Add chat panel state to App component in cmd/explorer/frontend/src/App.tsx (collapsed state only)
- [X] T021 [US1] Add ChatPanel to App layout in cmd/explorer/frontend/src/App.tsx (bottom panel positioning)
- [X] T022 [US1] Update App.css to support bottom panel layout in cmd/explorer/frontend/src/App.css
- [X] T023 [US1] Implement collapse/expand functionality in ChatPanel component
- [X] T024 [US1] Persist collapsed state in session storage within ChatPanel component
- [X] T025 [US1] Run component tests to verify User Story 1 (npm test from cmd/explorer/frontend/)

**Checkpoint**: Chat panel visible at bottom, can type in input field, can collapse/expand, tests passing

---

## Phase 4: User Story 2 - Send Queries and Display Responses (Priority: P1) 🎯 MVP

**Goal**: Users can type a query, submit it (Enter or button), see both query and response in conversation area with visual distinction and loading indicator

**Independent Test**: Type a query, submit it, verify query appears right-aligned, loading indicator shows, response appears left-aligned with different background

### Tests for User Story 2

- [ ] T026 [US2] Write chat service wrapper test in cmd/explorer/frontend/src/services/chat.test.ts
- [ ] T027 [US2] Extend ChatPanel test to verify query submission and response display

### Implementation for User Story 2

- [ ] T028 [US2] Create chat service wrapper in cmd/explorer/frontend/src/services/chat.ts wrapping Wails API calls
- [ ] T029 [US2] Add conversation state management to ChatPanel (messages array, isLoading, error)
- [ ] T030 [US2] Implement query submission handler in ChatPanel calling ProcessChatQuery via chat service
- [ ] T031 [US2] Add loading indicator to ChatPanel (shown while isLoading is true)
- [ ] T032 [US2] Implement message display logic in ChatPanel rendering ChatMessage components
- [ ] T033 [US2] Add visual distinction styles to ChatMessage.css (user right-aligned, system left-aligned, different backgrounds per FR-006)
- [ ] T034 [US2] Implement Enter key submission in ChatInput component (FR-003)
- [ ] T035 [US2] Implement Shift+Enter newline in ChatInput component (FR-004)
- [ ] T036 [US2] Implement empty query prevention in ChatInput component (FR-015)
- [ ] T037 [US2] Implement query submission disabling while loading in ChatInput component (FR-021)
- [ ] T038 [US2] Add error handling with user-friendly messages in ChatPanel (FR-011)
- [ ] T039 [US2] Add timeout error handling (60 seconds) in chat service wrapper (FR-024)
- [ ] T040 [US2] Add retry functionality for errors and timeouts in ChatPanel (FR-022)
- [ ] T041 [US2] Run component tests to verify User Story 2 (npm test from cmd/explorer/frontend/)
- [ ] T042 [US2] Manual test: Submit query, verify query and response appear with correct styling. Verify SC-002: measure that submitting a query requires ≤2 clicks/keystrokes (type + Enter = 1 keystroke)

**Checkpoint**: Full query submission and response display working, visual distinction clear, loading states correct, error handling functional

---

## Phase 5: User Story 3 - Conversation History (Priority: P2)

**Goal**: Users can scroll through conversation history to review previous queries and responses within the current session

**Independent Test**: Submit multiple queries, verify all are visible, conversation scrolls, auto-scrolls to latest

### Tests for User Story 3

- [ ] T043 [US3] Write test for conversation scrolling behavior in ChatPanel.test.tsx

### Implementation for User Story 3

- [ ] T044 [US3] Add scrollable container to conversation area in ChatPanel.tsx (FR-013)
- [ ] T045 [US3] Implement auto-scroll to latest message in ChatPanel component (FR-014)
- [ ] T046 [US3] Add scroll-to-bottom on new message logic in ChatPanel useEffect hook
- [ ] T047 [US3] Test with 50+ messages to verify performance (SC-006). Measure query response time for 10 sample queries to verify SC-003: 95% complete within 3 seconds
- [ ] T048 [US3] Run component tests to verify User Story 3 (npm test from cmd/explorer/frontend/)

**Checkpoint**: Conversation history scrollable, auto-scrolls to new messages, handles large message counts

---

## Phase 6: User Story 4 - Clear Conversation (Priority: P3)

**Goal**: Users can clear conversation history to start fresh without restarting the application

**Independent Test**: Build conversation history, click clear button, verify conversation area empty, can start new conversation

### Tests for User Story 4

- [ ] T049 [US4] Write test for clear conversation functionality in ChatPanel.test.tsx

### Implementation for User Story 4

- [ ] T050 [US4] Add clear button to ChatPanel UI in ChatPanel.tsx
- [ ] T051 [US4] Implement clear handler calling ClearChatContext via chat service in ChatPanel
- [ ] T052 [US4] Reset local conversation state in ChatPanel when clearing (messages array, error, lastQuery)
- [ ] T053 [US4] Run component tests to verify User Story 4 (npm test from cmd/explorer/frontend/)

**Checkpoint**: All user stories complete - users can display, use, scroll, and clear chat interface

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation across all user stories

- [ ] T054 [P] Add accessibility attributes (ARIA labels) to chat components
- [ ] T055 [P] Verify special characters display correctly (FR-023)
- [ ] T056 Verify all 24 functional requirements are implemented (FR-001 through FR-024)
- [ ] T057 Verify all success criteria (SC-001 through SC-008) are met
- [ ] T058 [P] Update README.md with chat interface usage instructions
- [ ] T059 Run full quickstart.md validation workflow (wails dev, manual testing)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-6)**: All depend on Foundational phase completion
  - Phase 3 (US1 - Display) is independent, can start immediately after Phase 2
  - Phase 4 (US2 - Send/Receive) depends on Phase 3 (needs ChatPanel, ChatMessage, ChatInput)
  - Phase 5 (US3 - History) depends on Phase 4 (needs message display working)
  - Phase 6 (US4 - Clear) depends on Phase 3 and Phase 4 (needs panel and messages)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
  - Creates ChatPanel, ChatMessage, ChatInput components
  - Establishes bottom panel layout
  - Implements collapse/expand
  
- **User Story 2 (P1)**: Depends on User Story 1 completion - Needs ChatPanel structure
  - Adds query submission logic
  - Adds response display logic
  - Adds loading and error states
  - Integrates with backend ProcessChatQuery method
  
- **User Story 3 (P2)**: Depends on User Story 2 completion - Needs message display working
  - Adds scrolling behavior
  - Adds auto-scroll to latest
  
- **User Story 4 (P3)**: Depends on User Story 1 and User Story 2 - Needs panel and messages
  - Adds clear button to existing panel
  - Calls ClearChatContext backend method
  - Resets frontend state

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Components with styles: create .tsx and .css in parallel
- ChatInput and ChatMessage can be implemented in parallel (both P1 for US1)
- ChatPanel depends on understanding ChatInput and ChatMessage contracts
- Integration tests run after all components implemented
- Manual testing confirms all acceptance criteria

### Parallel Opportunities

- **Phase 1**: All setup tasks (T002, T003) can run in parallel
- **Phase 2**: Stub LLM (T004) and chat adapter (T005) can be created in parallel
- **Phase 3 Tests**: ChatInput test (T011) and ChatMessage test (T012) can run in parallel
- **Phase 3 Implementation**: 
  - ChatInput component + styles (T014, T015) parallel
  - ChatMessage component + styles (T016, T017) parallel
  - Above pairs can run in parallel with each other
- **Phase 7**: Documentation (T059) and accessibility (T054) and special chars (T055) can run in parallel

---

## Parallel Example: User Story 1 Components

Developer A and Developer B working together:

```bash
# Developer A: ChatInput component
cd cmd/explorer/frontend/src/components
# Write test first
npm test -- ChatInput.test.tsx  # Should FAIL (RED)
# Implement component
touch ChatInput.tsx ChatInput.css
# ... implement ...
npm test -- ChatInput.test.tsx  # Should PASS (GREEN)

# Developer B (in parallel): ChatMessage component
cd cmd/explorer/frontend/src/components
# Write test first
npm test -- ChatMessage.test.tsx  # Should FAIL (RED)
# Implement component
touch ChatMessage.tsx ChatMessage.css
# ... implement ...
npm test -- ChatMessage.test.tsx  # Should PASS (GREEN)

# Both developers finished - now integrate in ChatPanel
npm test -- ChatPanel.test.tsx   # Should FAIL initially (RED)
# Implement ChatPanel using both components
npm test -- ChatPanel.test.tsx   # Should PASS (GREEN)
```

---

## Parallel Example: User Story 2 Query Flow

Different aspects can be implemented in parallel:

```bash
# Terminal 1: Chat service wrapper (backend integration)
cd cmd/explorer/frontend/src/services
npm test -- chat.test.ts  # RED
# Implement chat.ts wrapping ProcessChatQuery
npm test -- chat.test.ts  # GREEN

# Terminal 2: ChatInput keyboard handling (UI interaction)
cd cmd/explorer/frontend/src/components
# Extend ChatInput test with Enter/Shift+Enter
npm test -- ChatInput.test.tsx  # RED (new test cases)
# Implement keyboard logic
npm test -- ChatInput.test.tsx  # GREEN

# Terminal 3: ChatMessage styling (visual distinction)
cd cmd/explorer/frontend/src/components
# Update ChatMessage.css with user/system styles
# Manual visual check in wails dev
```

---

## Implementation Strategy

### MVP Definition

The MVP consists of **User Story 1 (P1) + User Story 2 (P1)**:
- Display chat interface panel at bottom
- Send queries and receive responses
- Loading indicator and error handling
- Basic visual distinction

This provides complete core functionality for users to interact with the chat system.

### Incremental Delivery

1. **Sprint 1** (MVP): Phases 1-4 (Setup + Foundation + US1 + US2)
   - Deliverable: Functional chat interface with query/response capability
   - Testable: Users can submit queries and see responses
   
2. **Sprint 2** (Enhanced): Phase 5 (US3)
   - Deliverable: Scrollable conversation history
   - Testable: Users can review previous queries
   
3. **Sprint 3** (Polish): Phase 6 + Phase 7 (US4 + Polish)
   - Deliverable: Clear functionality and final polish
   - Testable: Complete feature per specification

### Testing Strategy

Following TDD workflow from quickstart.md:

1. **Red Phase**: Write failing test
2. **Green Phase**: Implement minimum code to pass
3. **Refactor Phase**: Clean up while keeping tests green

Test levels:
- **Contract tests**: Backend Wails bindings (Go)
- **Component tests**: React components (TypeScript)
- **Integration tests**: Manual testing with `wails dev`
- **End-to-end tests**: All acceptance scenarios from spec.md

---

## Task Summary

- **Total Tasks**: 59
- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundation)**: 7 tasks (BLOCKS all user stories)
- **Phase 3 (US1 - Display)**: 15 tasks (3 tests + 12 implementation)
- **Phase 4 (US2 - Send/Receive)**: 17 tasks (2 tests + 15 implementation)
- **Phase 5 (US3 - History)**: 6 tasks (1 test + 5 implementation)
- **Phase 6 (US4 - Clear)**: 5 tasks (1 test + 4 implementation)
- **Phase 7 (Polish)**: 6 tasks

**Parallel Opportunities**: 19 tasks marked [P] for parallel execution

**Independent Test Criteria**:
- US1: Open app → chat panel visible at bottom, accepts input, can collapse
- US2: Submit query → appears right-aligned, loading shows, response appears left-aligned
- US3: Submit 5+ queries → all visible, scrollable, auto-scrolls to latest
- US4: Clear button → conversation empty, can submit new query

**MVP Scope**: Phases 1-4 (T001-T042) = 42 tasks for minimum viable chat interface
