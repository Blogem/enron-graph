# Chat Input History Navigation - Tasks

## 1. Type Definitions

- [x] 1.1 Add `onHistoryNavigate?: (direction: 'up' | 'down') => void` to ChatInputProps interface

## 2. ChatPanel State Management

- [x] 2.1 Add state: `historyIndex` (number, initial -1)
- [x] 2.2 Add state: `savedDraft` (string, initial '')
- [x] 2.3 Build `messageHistory` array from messages filtered by sender === 'user'
- [x] 2.4 Implement `handleHistoryNavigate` function with up/down logic
- [x] 2.5 Reset historyIndex to -1 and clear savedDraft after message submission

## 3. ChatInput Arrow Key Handling

- [x] 3.1 Add ArrowUp handler in `handleKeyDown` to call `onHistoryNavigate('up')`
- [x] 3.2 Add ArrowDown handler in `handleKeyDown` to call `onHistoryNavigate('down')`
- [x] 3.3 Prevent default behavior for arrow keys

## 4. Integration

- [x] 4.1 Pass `onHistoryNavigate` prop from ChatPanel to ChatInput

## 5. Verify

- [x] 5.1 Test ArrowUp navigates backward through messages
- [x] 5.2 Test ArrowDown navigates forward and restores draft
- [x] 5.3 Test draft preservation when starting navigation
- [x] 5.4 Test empty history doesn't cause errors
- [x] 5.5 Test history resets after sending message
