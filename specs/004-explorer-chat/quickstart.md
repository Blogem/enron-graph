# Quickstart: Graph Explorer Chat Interface

**Feature**: 004-explorer-chat  
**Last Updated**: 2026-01-27

## Overview

This guide helps developers set up their environment, understand the architecture, and start implementing the chat interface feature.

---

## Prerequisites

- Go 1.21 or later
- Node.js 18 or later
- Wails CLI v2 installed (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Basic familiarity with React and Go
- Understanding of the existing cmd/explorer application

---

## Architecture Overview

```
┌─────────────────────────────────────────┐
│         React Frontend (TypeScript)      │
│  ┌────────────┐      ┌───────────────┐  │
│  │ ChatPanel  │──────│  ChatMessage  │  │
│  │            │      │  ChatInput    │  │
│  └────────────┘      └───────────────┘  │
│         │                                │
│         │ Wails API Call                 │
│         ↓                                │
└─────────────────────────────────────────┘
          │
          │ ProcessChatQuery(query string)
          ↓
┌─────────────────────────────────────────┐
│         Go Backend (cmd/explorer)        │
│  ┌────────────────────────────────────┐ │
│  │  App.ProcessChatQuery()            │ │
│  │    ├─ Validate query               │ │
│  │    ├─ Call chat.Handler            │ │
│  │    └─ Return response              │ │
│  └────────────────────────────────────┘ │
│         │                                │
│         ↓                                │
│  ┌────────────────────────────────────┐ │
│  │  internal/chat package             │ │
│  │    ├─ Handler.ProcessQuery()       │ │
│  │    ├─ Context (conversation state) │ │
│  │    └─ Repository (graph queries)   │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

---

## Development Setup

### 1. Clone and Navigate

```bash
cd /Users/jochem/code/enron-graph-2
git checkout 004-explorer-chat
```

### 2. Install Dependencies

Frontend:
```bash
cd cmd/explorer/frontend
npm install
```

Backend (if new Go dependencies are added):
```bash
cd ../..
go mod tidy
```

### 3. Verify Existing Setup

Run the explorer to ensure everything works:
```bash
cd cmd/explorer
wails dev
```

The application should launch. If it doesn't, troubleshoot existing setup before proceeding.

---

## Implementation Order (TDD Workflow)

Follow this sequence to implement the feature using Test-Driven Development:

### Phase 1: Backend Integration

1. **Write contract test** for `ProcessChatQuery`
   - Location: `tests/contract/chat_bindings_test.go`
   - Verify method is callable, handles empty queries, returns responses

2. **Implement stub dependencies**
   - Create `StubLLMClient` (returns mock responses)
   - Create `ChatAdapter` (implements Repository using ent client)

3. **Add methods to App struct**
   - Add `ProcessChatQuery(query string) (string, error)` to `app.go`
   - Add `ClearChatContext() error` to `app.go`
   - Initialize chat handler in `NewApp()` in `main.go`

4. **Run tests** - should pass
   ```bash
   go test ./tests/contract/...
   ```

### Phase 2: Frontend Components

5. **Create TypeScript types**
   - Copy `specs/004-explorer-chat/contracts/chat-api.ts` to `cmd/explorer/frontend/src/types/chat.ts`

6. **Write test** for `ChatInput` component
   - Location: `cmd/explorer/frontend/src/components/ChatInput.test.tsx`
   - Verify Enter submits, Shift+Enter adds newline, empty queries prevented

7. **Implement ChatInput component**
   - Create `ChatInput.tsx` and `ChatInput.css`
   - Implement keyboard handling, submit logic

8. **Run test** - should pass
   ```bash
   cd cmd/explorer/frontend
   npm test -- ChatInput.test.tsx
   ```

9. **Repeat for ChatMessage** (test first, then implement)

10. **Repeat for ChatPanel** (test first, then implement)

### Phase 3: Integration

11. **Update App.tsx**
    - Add chat panel to layout
    - Add state management
    - Wire up Wails API calls

12. **Update App.css**
    - Add bottom panel styles
    - Add flexbox layout adjustments

13. **End-to-end test**
    - Run `wails dev`
    - Manually test all user stories
    - Verify all acceptance criteria

---

## File Locations Quick Reference

| What | Where |
|------|-------|
| Backend Go methods | `cmd/explorer/app.go` |
| Chat handler init | `cmd/explorer/main.go` |
| TypeScript types | `cmd/explorer/frontend/src/types/chat.ts` |
| ChatPanel component | `cmd/explorer/frontend/src/components/ChatPanel.tsx` |
| ChatMessage component | `cmd/explorer/frontend/src/components/ChatMessage.tsx` |
| ChatInput component | `cmd/explorer/frontend/src/components/ChatInput.tsx` |
| Chat service wrapper | `cmd/explorer/frontend/src/services/chat.ts` |
| Backend contract tests | `tests/contract/chat_bindings_test.go` |
| Frontend component tests | `cmd/explorer/frontend/src/components/*.test.tsx` |

---

## Running Tests

### Backend Tests

```bash
# All tests
go test ./...

# Just contract tests
go test ./tests/contract/...

# With verbose output
go test -v ./tests/contract/...
```

### Frontend Tests

```bash
cd cmd/explorer/frontend

# All tests
npm test

# Specific component
npm test -- ChatPanel.test.tsx

# With coverage
npm test -- --coverage
```

### Manual Testing

```bash
cd cmd/explorer
wails dev
```

Test checklist:
- [ ] Chat panel visible at bottom
- [ ] Can type in input field
- [ ] Enter key submits query
- [ ] Shift+Enter adds newline
- [ ] Loading indicator appears
- [ ] Response displays below query
- [ ] Messages visually distinct (right vs left, different colors)
- [ ] Can scroll conversation history
- [ ] Clear button empties conversation
- [ ] Collapse/expand button works
- [ ] Empty queries prevented
- [ ] Error messages display with retry button
- [ ] Timeout after 60 seconds shows error

---

## Common Issues & Solutions

### Issue: Wails bindings not working

**Solution**: Rebuild Wails bindings
```bash
cd cmd/explorer
wails generate module
```

### Issue: Frontend can't call Go methods

**Solution**: Ensure methods are exported (capitalized) and on the App struct

### Issue: Chat handler returns errors

**Solution**: Check that stub LLM client and chat adapter are properly initialized in `main.go`

### Issue: Tests fail with "method not found"

**Solution**: Rebuild the app first, then run tests
```bash
wails build
go test ./tests/contract/...
```

---

## Code Snippets

### Creating a ChatMessage

```typescript
const userMessage: ChatMessage = {
    id: `${Date.now()}-user`,
    text: query,
    sender: 'user',
    timestamp: new Date(),
};
```

### Calling Wails API

```typescript
import { ProcessChatQuery } from '../wailsjs/go/main/App';

try {
    const response = await ProcessChatQuery(query);
    // Handle response
} catch (error) {
    // Handle error
}
```

### Adding timeout to frontend call

```typescript
const queryWithTimeout = async (query: string): Promise<string> => {
    const timeout = new Promise<never>((_, reject) =>
        setTimeout(() => reject(new Error('Query timeout after 60 seconds')), 60000)
    );
    return Promise.race([ProcessChatQuery(query), timeout]);
};
```

---

## Next Steps

After completing implementation:

1. Run full test suite: `go test ./... && cd cmd/explorer/frontend && npm test`
2. Manual testing against acceptance criteria
3. Code review (self or peer)
4. Update this quickstart if any pain points discovered
5. Request user confirmation before committing (Constitution Principle IX)

---

## Additional Resources

- Wails Documentation: https://wails.io/docs
- React Testing Library: https://testing-library.com/react
- Existing Components: See `cmd/explorer/frontend/src/components/` for patterns
- internal/chat Package: See `internal/chat/` for handler implementation details

---

## Getting Help

- Review existing components (SchemaPanel, DetailPanel) for patterns
- Check `specs/004-explorer-chat/research.md` for design decisions
- Refer to `specs/004-explorer-chat/data-model.md` for entity definitions
- See `specs/004-explorer-chat/contracts/` for API contracts
