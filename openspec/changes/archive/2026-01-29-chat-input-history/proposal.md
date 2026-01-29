# Chat Input History Navigation

## Why

The chat interface currently lacks command history navigation, forcing users to retype previous queries when they want to rerun or modify them. This is a standard UX pattern (terminal-style arrow key navigation) that users expect in conversational interfaces.

## What Changes

- Add arrow up/down key navigation to cycle through previous user messages
- Track message history in ChatPanel component
- Restore previous messages to the input field when navigating
- Preserve current draft when starting history navigation
- Reset history position after submitting a new message

## Capabilities

### New Capabilities
- `chat-input-history`: Users can press ArrowUp to navigate backward through their message history and ArrowDown to navigate forward, with the current draft preserved

## Impact

- `cmd/explorer/frontend/src/components/ChatPanel.tsx`: Add history state management and pass history navigation callbacks to ChatInput
- `cmd/explorer/frontend/src/components/ChatInput.tsx`: Add ArrowUp/ArrowDown key handlers and integrate with history navigation
- `cmd/explorer/frontend/src/types/chat.ts`: Add history-related props to ChatInputProps interface
