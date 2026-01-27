# Research: Graph Explorer Chat Interface

**Feature**: 004-explorer-chat  
**Phase**: 0 (Outline & Research)  
**Date**: 2026-01-27

## Research Tasks Resolved

This document consolidates findings from technical research needed to implement the chat interface feature.

## 1. Wails Binding Patterns

**Research Question**: How are Go methods exposed to React frontend in the existing explorer application?

**Findings**:

From `cmd/explorer/app.go`, the pattern is:
1. Methods on the `App` struct are automatically bound by Wails
2. Method signatures must be exported (capitalized)
3. Methods can accept primitive types and structs
4. Return types can be (value, error) or just value
5. Context is available via `a.ctx` stored during startup

**Example from existing code**:
```go
// GetSchema returns the complete schema metadata
func (a *App) GetSchema() (*explorer.SchemaResponse, error) {
    return a.schemaService.GetSchema(a.ctx)
}
```

**Decision**: 
- Add `ProcessChatQuery(query string) (string, error)` method to App struct
- Add `ClearChatContext() error` method to App struct
- Methods will delegate to internal/chat.Handler.ProcessQuery

**Rationale**: Follows established pattern, maintains separation of concerns (UI in cmd/explorer, logic in internal/chat)

---

## 2. Chat Handler Integration

**Research Question**: What dependencies does internal/chat.Handler require and how should it be initialized?

**Findings**:

From `internal/chat/handler.go` and `internal/chat/types.go`:
```go
// NewHandler creates a new chat handler
func NewHandler(llm LLMClient, repo Repository) Handler
```

Required interfaces:
- `LLMClient` - for GenerateCompletion and GenerateEmbedding
- `Repository` - for graph database operations (FindEntityByName, TraverseRelationships, etc.)

**Current Gap**: 
- The explorer application doesn't currently have LLM integration
- Need to implement or mock these dependencies

**Decision**: 
For initial implementation, create stub implementations:
1. `StubLLMClient` - returns mock responses for development/testing
2. Reuse existing ent client for Repository implementation (needs adapter)

**Alternative Considered**: Wait for full LLM integration
**Rejected Because**: Spec says "reuse existing chat package" - we can integrate chat UI now with stubs, replace with real LLM later

---

## 3. Chat Context Management

**Research Question**: How should conversation context be managed across the Wails bridge?

**Findings**:

From `internal/chat/context.go`:
```go
// Context interface for conversation context management
type Context interface {
    AddQuery(query, response string)
    GetHistory() []HistoryEntry
    TrackEntity(name, entityType string, id int)
    // ...
}
```

Context is stateful and maintains:
- Conversation history (last 5 entries by default)
- Tracked entities for pronoun resolution
- Serialization support

**Decision**:
- Store single global `chat.Context` instance in App struct
- Initialize once in `NewApp()`
- Reuse same context across all queries in a session
- Clear method will call `context.Clear()`

**Rationale**: 
- Matches spec requirement "preserve conversation history within a session"
- Simple implementation, no need for session management
- Aligns with single-user desktop application model

---

## 4. React Component Architecture

**Research Question**: What component patterns are used in existing explorer UI?

**Findings**:

From existing components (`SchemaPanel.tsx`, `DetailPanel.tsx`, `FilterBar.tsx`):
- Functional components with hooks (useState, useEffect, useCallback)
- CSS modules for styling (separate .css file per component)
- Props-based composition
- Services layer for API calls (`services/wails.ts`)
- TypeScript for type safety

**Pattern Example**:
```typescript
interface SchemaPanelProps {
    schema: explorer.SchemaResponse | null;
    onTypeSelect: (typeName: string) => void;
    // ...
}

function SchemaPanel({ schema, onTypeSelect, ... }: SchemaPanelProps) {
    // Component implementation
}
```

**Decision**:
Component hierarchy:
```
ChatPanel (container)
├── ChatMessage[] (display message list)
└── ChatInput (input field + submit)
```

Each component:
- Separate .tsx and .css files
- Props interface defined
- Event handlers passed as props
- Follows existing naming conventions

---

## 5. Bottom Panel Layout Implementation

**Research Question**: How to add a collapsible bottom panel to the existing layout?

**Findings**:

Current `App.tsx` layout uses CSS Flexbox:
- Main container is flex column
- Schema panel and graph area are flex items
- Panels use absolute positioning within flex containers

**Decision**:
Add bottom panel using:
```css
.app-container {
    display: flex;
    flex-direction: column;
    height: 100vh;
}

.main-content {
    flex: 1;
    display: flex;
    min-height: 0; /* Allow flex child to shrink */
}

.chat-panel {
    height: 300px; /* Default expanded height */
    flex-shrink: 0;
    transition: height 0.3s ease;
}

.chat-panel.collapsed {
    height: 40px; /* Collapsed height (just header) */
}
```

**Rationale**:
- Flex layout integrates cleanly with existing structure
- transition provides smooth expand/collapse animation
- flex-shrink: 0 prevents unwanted resizing during graph interactions

---

## 6. Message Visual Distinction

**Research Question**: How to implement the right-aligned user / left-aligned system message pattern?

**Findings**:

Clarification specified: "Different background colors with alignment (user right, system left)"

**Decision**:
```css
.chat-message {
    display: flex;
    margin: 8px 0;
}

.chat-message.user {
    justify-content: flex-end;
}

.chat-message.system {
    justify-content: flex-start;
}

.message-bubble {
    max-width: 70%;
    padding: 8px 12px;
    border-radius: 12px;
}

.message-bubble.user {
    background-color: #007AFF; /* Blue */
    color: white;
}

.message-bubble.system {
    background-color: #E5E5EA; /* Light gray */
    color: black;
}
```

**Rationale**:
- Follows common messaging app patterns (iOS Messages, WhatsApp)
- High contrast for accessibility
- Flexbox makes alignment trivial

---

## 7. Input Handling (Enter vs Shift+Enter)

**Research Question**: How to implement Enter submits, Shift+Enter for newline?

**Findings**:

Standard React pattern using onKeyDown event:

**Decision**:
```typescript
const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSubmit();
    }
    // Shift+Enter naturally adds newline (browser default)
};
```

**Rationale**:
- Leverages browser's natural Shift+Enter behavior
- Clean, simple implementation
- Matches user expectations from other chat apps

---

## 8. Timeout Implementation

**Research Question**: How to implement 60-second timeout for queries?

**Findings**:

Clarification specified: 60 seconds before displaying timeout error

**Decision**:
Frontend implementation using Promise.race:
```typescript
const queryWithTimeout = async (query: string): Promise<string> => {
    const timeout = new Promise<never>((_, reject) => 
        setTimeout(() => reject(new Error('Query timeout after 60 seconds')), 60000)
    );
    
    return Promise.race([
        wailsAPI.ProcessChatQuery(query),
        timeout
    ]);
};
```

Backend: Go's context.WithTimeout:
```go
ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
defer cancel()
// Use ctx for chat handler call
```

**Rationale**:
- Double protection (frontend + backend)
- Frontend timeout provides immediate UI feedback
- Backend timeout prevents resource leaks

---

## 9. Error Handling with Retry

**Research Question**: How to implement "try again" button for errors?

**Findings**:

Spec requires: "provide a 'try again' option when errors occur or responses timeout"

**Decision**:
Store last query in component state:
```typescript
const [lastQuery, setLastQuery] = useState<string>('');
const [error, setError] = useState<string | null>(null);

const handleSubmit = async (query: string) => {
    setLastQuery(query);
    try {
        // ... process query
    } catch (err) {
        setError(err.message);
    }
};

const handleRetry = () => {
    setError(null);
    handleSubmit(lastQuery);
};
```

UI shows retry button when error is present:
```tsx
{error && (
    <div className="error-message">
        {error}
        <button onClick={handleRetry}>Try Again</button>
    </div>
)}
```

**Rationale**:
- Simple state management
- Preserves user's query for retry
- Clear visual feedback

---

## 10. Testing Strategy

**Research Question**: What testing approach aligns with existing project and TDD principles?

**Findings**:

Project has:
- Go tests using standard testing package
- Existing test patterns in internal/chat/*_test.go

**Decision**:

Backend tests (Go):
1. Contract tests for Wails bindings (verify method signatures, error handling)
2. Integration tests for chat handler adapter

Frontend tests (React):
1. Component tests using React Testing Library
2. Test user interactions (typing, submitting, scrolling)
3. Test error states and retry behavior
4. Test expand/collapse behavior

Test-First Workflow:
1. Write failing contract test for ProcessChatQuery method
2. Implement method to pass test
3. Write failing component test for ChatInput
4. Implement ChatInput to pass test
5. Repeat for each component/feature

**Rationale**:
- Follows TDD principle (Principle III)
- Contract tests ensure Wails bindings don't break
- Component tests ensure UI behavior correctness
- Integration tests verify end-to-end flow

---

## Summary of Decisions

| Topic | Decision | Rationale |
|-------|----------|-----------|
| Go Method Binding | Add ProcessChatQuery, ClearChatContext to App struct | Follows existing Wails pattern |
| Chat Dependencies | Use stub LLMClient, adapter for Repository | Enables UI development, LLM integration later |
| Context Management | Single global chat.Context in App struct | Matches single-user session model |
| Component Structure | ChatPanel → ChatMessage[], ChatInput | Follows existing component patterns |
| Bottom Panel Layout | Flexbox with transition animation | Integrates cleanly with existing layout |
| Message Styling | Right-aligned blue (user), left-aligned gray (system) | Common messaging app pattern |
| Enter Key Behavior | Enter submits, Shift+Enter newline via onKeyDown | Standard React pattern |
| Timeout | 60s via Promise.race (frontend) + context.WithTimeout (backend) | Double protection |
| Error Retry | Store lastQuery, show retry button on error | Simple, user-friendly |
| Testing | Contract tests (Go) + component tests (React), TDD workflow | Aligns with constitution |

All research questions resolved. Ready for Phase 1 (Design).
