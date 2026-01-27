/**
 * Chat Service
 * Wrapper around Wails API for chat functionality
 * Location: cmd/explorer/frontend/src/services/chat.ts
 */

import { ProcessChatQuery, ClearChatContext } from '../wailsjs/go/main/App';

/**
 * Timeout duration for chat queries (60 seconds per FR-024)
 */
const CHAT_QUERY_TIMEOUT_MS = 60000;

/**
 * Custom error class for chat service errors
 */
export class ChatServiceError extends Error {
    public canRetry: boolean;
    public originalError: string;

    constructor(message: string, canRetry: boolean, originalError: string) {
        super(message);
        this.name = 'ChatServiceError';
        this.canRetry = canRetry;
        this.originalError = originalError;
    }
}

/**
 * Determine if an error is retryable based on error message
 */
function isRetryableError(errorMessage: string): boolean {
    const nonRetryablePatterns = [
        /invalid/i,
        /validation failed/i,
        /max length exceeded/i,
        /exceeds maximum length/i,
        /query cannot be empty/i
    ];

    return !nonRetryablePatterns.some(pattern => pattern.test(errorMessage));
}

/**
 * Chat service object providing chat functionality
 */
export const chatService = {
    /**
     * Process a chat query with timeout handling
     * @param query The user's natural language query
     * @returns The chat response text
     * @throws ChatServiceError with retry information
     */
    async processQuery(query: string): Promise<string> {
        let timeoutId: number | null = null;

        try {
            // Create a timeout promise
            const timeoutPromise = new Promise<never>((_, reject) => {
                timeoutId = setTimeout(() => {
                    reject(new Error('timeout'));
                }, CHAT_QUERY_TIMEOUT_MS);
            });

            // Race the API call against the timeout
            const response = await Promise.race([
                ProcessChatQuery(query),
                timeoutPromise
            ]);

            // Clear timeout if query completes
            if (timeoutId) {
                clearTimeout(timeoutId);
                timeoutId = null;
            }

            return response;
        } catch (error) {
            // Clear timeout on error
            if (timeoutId) {
                clearTimeout(timeoutId);
                timeoutId = null;
            }

            const originalError = error instanceof Error ? error.message : 'Unknown error';
            const isTimeout = originalError === 'timeout';
            const canRetry = isTimeout || isRetryableError(originalError);
            const userMessage = isTimeout
                ? 'Query timeout exceeded. Please try again.'
                : `Failed to process query: ${originalError}`;

            throw new ChatServiceError(
                userMessage,
                canRetry,
                originalError
            );
        }
    },

    /**
     * Clear the conversation context and history
     * @throws ChatServiceError if clearing fails
     */
    async clearContext(): Promise<void> {
        try {
            await ClearChatContext();
        } catch (error) {
            const originalError = error instanceof Error ? error.message : 'Unknown error';
            const userMessage = `Failed to clear conversation: ${originalError}`;

            throw new ChatServiceError(
                userMessage,
                true, // Clear operations are always retryable
                originalError
            );
        }
    }
};

// Export standalone functions for backward compatibility
export const processChatQuery = chatService.processQuery.bind(chatService);
export const clearChatContext = chatService.clearContext.bind(chatService);
