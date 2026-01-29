# Chat Input History Navigation - Design

## Context

The chat interface currently consists of:
- **ChatPanel**: Parent component managing conversation state (messages array)
- **ChatInput**: Child component with controlled textarea input
- Messages are stored as `ChatMessageType[]` containing both user and system messages

Users expect terminal-style command history (ArrowUp/ArrowDown navigation) but currently have no way to access previous queries.

## Goals / Non-Goals

**Goals:**
- Enable ArrowUp/ArrowDown navigation through user message history
- Preserve user's current draft when navigating history
- Keep implementation simple and maintainable
- Match familiar terminal/shell behavior patterns

**Non-Goals:**
- Persist history across browser sessions (use sessionStorage/localStorage)
- Search/filter through history
- Edit historical messages (they're read-only when recalled)
- Navigation through system responses (only user messages)

## Decisions

### Decision 1: State Management Location

**Choice:** Manage history state in ChatPanel, pass navigation callbacks to ChatInput

**Rationale:**
- ChatPanel already owns the messages array with full conversation history
- Can easily filter for user messages from existing state
- Keeps ChatInput focused on input handling, not history management
- Follows existing parent-child data flow pattern in the codebase

**Alternative considered:** Manage history in ChatInput
- Rejected: Would require duplicating message state or complex prop passing

### Decision 2: History Navigation State

**State needed:**
- `messageHistory: string[]` - Array of user message texts
- `historyIndex: number` - Current position in history (-1 = not navigating, 0 = most recent, etc.)
- `savedDraft: string` - Draft content saved when navigation starts

**Flow:**
1. User presses ArrowUp → save current input as draft (if historyIndex === -1), set historyIndex to 0, show messageHistory[0]
2. User presses ArrowUp again → increment historyIndex, show messageHistory[historyIndex]
3. User presses ArrowDown → decrement historyIndex, show messageHistory[historyIndex] or draft if at end
4. User submits → add to messageHistory, reset historyIndex to -1, clear savedDraft

### Decision 3: Draft Preservation Strategy

**Choice:** Save draft on first ArrowUp, restore when navigating past end with ArrowDown

**Rationale:**
- Matches terminal behavior (bash, zsh, etc.)
- User doesn't lose partially-typed work
- Simple state machine: draft exists only while historyIndex >= 0

### Decision 4: Message Filtering

**Choice:** Build history array from messages.filter(m => m.sender === 'user')

**Rationale:**
- System responses aren't useful in input history
- Keeps history navigation focused on what the user said
- Simple filter on existing data structure

## Implementation Approach

### ChatPanel Changes

1. Add state variables for history management
2. Build `messageHistory` array from `messages` filtered by sender === 'user'
3. Implement `handleHistoryNavigate(direction: 'up' | 'down')` function
4. Pass navigation handler and current value to ChatInput

### ChatInput Changes

1. Add `onHistoryNavigate` prop to ChatInputProps
2. Enhance `handleKeyDown` to intercept ArrowUp/ArrowDown
3. Call `onHistoryNavigate('up' | 'down')` appropriately
4. Prevent default behavior for arrow keys to avoid cursor movement

### Edge Cases

- Empty history → navigation does nothing
- At boundary (oldest/newest) → stays at boundary, no wrap-around
- Editing historical message → becomes new draft if user types before submitting
