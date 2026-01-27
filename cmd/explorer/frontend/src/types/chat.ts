/**
 * Chat Interface Types
 * Based on contracts from specs/004-explorer-chat/contracts/chat-api.ts
 */

/**
 * Represents a single message in the chat conversation
 */
export interface ChatMessage {
    /** Unique identifier for the message */
    id: string;

    /** The message content/text */
    text: string;

    /** Who sent the message */
    sender: 'user' | 'system';

    /** When the message was created */
    timestamp: Date;
}

/**
 * Manages the state of the active chat session
 */
export interface ConversationSession {
    /** Array of all messages in chronological order */
    messages: ChatMessage[];

    /** Whether a query is currently being processed */
    isLoading: boolean;

    /** Current error message, or null if no error */
    error: string | null;

    /** Last submitted query (for retry functionality) */
    lastQuery: string | null;

    /** Whether the chat panel is collapsed */
    isCollapsed: boolean;
}

/**
 * Props for the ChatPanel component
 */
export interface ChatPanelProps {
    /** Whether the panel is initially collapsed */
    initialCollapsed?: boolean;

    /** Callback when panel collapse state changes */
    onCollapseChange?: (collapsed: boolean) => void;
}

/**
 * Props for the ChatMessage component
 */
export interface ChatMessageProps {
    /** The message to display */
    message: ChatMessage;
}

/**
 * Props for the ChatInput component
 */
export interface ChatInputProps {
    /** Callback when user submits a query */
    onSubmit: (query: string) => void;

    /** Whether input is disabled (e.g., while loading) */
    disabled?: boolean;

    /** Placeholder text for the input field */
    placeholder?: string;
}

/**
 * Wails API methods for chat functionality
 * These match the Go method signatures in cmd/explorer/app.go
 */
export interface ChatAPI {
    /**
     * Process a chat query and return the response
     * @param query The user's natural language query
     * @returns The chat response text
     * @throws Error if query processing fails or times out
     */
    ProcessChatQuery(query: string): Promise<string>;

    /**
     * Clear the conversation context and history
     * @throws Error if clearing fails
     */
    ClearChatContext(): Promise<void>;
}

/**
 * Helper type for chat service error handling
 */
export interface ChatError {
    message: string;
    canRetry: boolean;
    isTimeout: boolean;
}
