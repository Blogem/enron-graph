# Data Model: Graph Explorer Chat Interface

**Feature**: 004-explorer-chat  
**Phase**: 1 (Design & Contracts)  
**Date**: 2026-01-27

## Overview

This document defines the data entities for the chat interface feature. These are UI-layer entities that represent chat state and messages, distinct from the business logic entities in internal/chat.

---

## Entities

### 1. ChatMessage (Frontend TypeScript)

**Purpose**: Represents a single message in the conversation display

**Location**: `cmd/explorer/frontend/src/types/chat.ts`

**Properties**:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| id | string | Yes | Unique identifier for the message (UUID or timestamp-based) |
| text | string | Yes | The message content |
| sender | 'user' \| 'system' | Yes | Who sent the message |
| timestamp | Date | Yes | When the message was created |

**Validation Rules**:
- `text` cannot be empty
- `sender` must be exactly 'user' or 'system'
- `timestamp` must be a valid Date object

**Example**:
```typescript
{
    id: "1706371200000-user",
    text: "Show me Kenneth Lay",
    sender: "user",
    timestamp: new Date("2026-01-27T10:00:00Z")
}
```

**Relationships**: None (self-contained value object)

---

### 2. ConversationSession (Frontend TypeScript)

**Purpose**: Manages the UI state for the active chat session

**Location**: `cmd/explorer/frontend/src/types/chat.ts`

**Properties**:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| messages | ChatMessage[] | Yes | Array of all messages in the conversation |
| isLoading | boolean | Yes | Whether a query is currently being processed |
| error | string \| null | Yes | Current error message, or null if no error |
| lastQuery | string \| null | Yes | Last submitted query (for retry functionality) |
| isCollapsed | boolean | Yes | Whether the chat panel is collapsed |

**Validation Rules**:
- `messages` array maintains chronological order (oldest first)
- `isLoading` must be false when `error` is set
- `lastQuery` is null initially, set on first query

**State Transitions**:
```
Initial: { messages: [], isLoading: false, error: null, lastQuery: null, isCollapsed: false }
          ↓ (user submits query)
Loading: { messages: [...prev, userMsg], isLoading: true, error: null, lastQuery: query }
          ↓ (response received)
Success: { messages: [...prev, userMsg, systemMsg], isLoading: false, error: null }
          ↓ (error occurred)
Error:   { messages: [...prev, userMsg], isLoading: false, error: "...", lastQuery: query }
          ↓ (user retries)
Loading: (back to Loading state)
          ↓ (user clears conversation)
Cleared: { messages: [], isLoading: false, error: null, lastQuery: null }
          ↓ (user toggles collapse)
Toggled: { ...prev, isCollapsed: !prev.isCollapsed }
```

**Example**:
```typescript
{
    messages: [
        {
            id: "1706371200000-user",
            text: "Show me Kenneth Lay",
            sender: "user",
            timestamp: new Date("2026-01-27T10:00:00Z")
        },
        {
            id: "1706371201000-system",
            text: "Kenneth Lay (person)\\nProperties:\\n  email: kenneth.lay@enron.com",
            sender: "system",
            timestamp: new Date("2026-01-27T10:00:01Z")
        }
    ],
    isLoading: false,
    error: null,
    lastQuery: "Show me Kenneth Lay",
    isCollapsed: false
}
```

**Relationships**: Contains array of ChatMessage entities

---

### 3. ChatQueryRequest (Backend Go)

**Purpose**: Request payload for processing a chat query

**Location**: `cmd/explorer/app.go` (internal to method signature)

**Properties**:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| query | string | Yes | The user's natural language query |

**Validation Rules**:
- `query` cannot be empty or whitespace-only
- Maximum length: 1000 characters (reasonable query size)

**Example**:
```go
query := "Show me Kenneth Lay"
```

**Note**: This is not a struct, just a method parameter. Included for completeness.

---

### 4. ChatQueryResponse (Backend Go)

**Purpose**: Response payload from processing a chat query

**Location**: `cmd/explorer/app.go` (method return value)

**Properties**:

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| response | string | Yes | The formatted response text from the chat handler |
| error | error | No | Error if query processing failed |

**Validation Rules**:
- If `error` is non-nil, `response` may be empty
- `response` should be non-empty on success

**Example**:
```go
response := "Kenneth Lay (person)\\nProperties:\\n  email: kenneth.lay@enron.com"
err := nil
```

**Note**: Go methods return (value, error), so this is represented as return types. Included for completeness.

---

## Entity Relationships

```
ConversationSession (1) ──── contains ──── (0..n) ChatMessage

Frontend TypeScript          Wails Bridge          Backend Go
──────────────────────       ────────────        ──────────────
ConversationSession     <──>  ProcessChatQuery  <──> chat.Handler
                              (string → string)      (internal/chat)

ChatMessage (display)   <──>  (no direct map)   <──> HistoryEntry
                                                      (internal/chat/types.go)
```

**Notes**:
- Frontend entities are purely for UI state management
- Backend delegates to `internal/chat.Handler` which has its own internal entities
- No direct serialization between frontend ChatMessage and backend HistoryEntry (different concerns)

---

## Data Flow

### Query Submission Flow

```
1. User types query in ChatInput component
   ↓
2. ChatPanel creates ChatMessage entity (sender: 'user')
   ↓
3. ChatPanel adds message to ConversationSession.messages
   ↓
4. ChatPanel calls wailsAPI.ProcessChatQuery(query)
   ↓
5. Go App.ProcessChatQuery receives query string
   ↓
6. App delegates to chat.Handler.ProcessQuery (internal/chat)
   ↓
7. Handler returns response string
   ↓
8. App returns response to frontend
   ↓
9. ChatPanel creates ChatMessage entity (sender: 'system', text: response)
   ↓
10. ChatPanel adds system message to ConversationSession.messages
```

### Error Flow

```
1-5. (same as above)
   ↓
6. Handler returns error
   ↓
7. App returns error to frontend
   ↓
8. ChatPanel sets ConversationSession.error
   ↓
9. UI displays error message with retry button
   ↓
10. User clicks retry → returns to step 1 with lastQuery
```

---

## Persistence

**Session Storage**: None
- All data stored in React component state (useState)
- Lost on app close or page refresh
- Matches spec requirement: "preserve conversation history within a session (until cleared or app closed)"

**Backend Storage**: None
- Chat context managed in-memory by `internal/chat.Context`
- No database persistence required

---

## Validation Summary

| Entity | Frontend Validation | Backend Validation |
|--------|-------------------|-------------------|
| ChatMessage | TypeScript type checking | N/A (not sent to backend) |
| ConversationSession | React state management | N/A (frontend only) |
| ChatQueryRequest | Empty string check before submit | Empty/whitespace check in App method |
| ChatQueryResponse | Type checking on receive | Error handling in chat.Handler |

---

## Migration Path

This is a new feature, so no data migration is required. However, for future extensibility:

**If we add persistence later**:
1. Add database table for conversation sessions
2. Add foreign key relationship to users (if multi-user)
3. Serialize ChatMessage array to JSON column or separate messages table
4. Add created_at/updated_at timestamps

**For now**: Not needed (YAGNI principle)

---

## Summary

- **2 frontend entities**: ChatMessage (value object), ConversationSession (state container)
- **No new backend entities**: Reuses internal/chat types
- **No persistence**: In-memory only, session-scoped
- **Simple validation**: Type checking + empty string checks
- **Clean separation**: UI state (frontend) vs business logic (internal/chat)

All entities defined. Ready for contract generation.
