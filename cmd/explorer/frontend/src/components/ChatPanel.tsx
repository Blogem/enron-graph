/** @jsxImportSource react */
import { FC, useState, useEffect, useRef } from 'react';
import './ChatPanel.css';
import ChatInput from './ChatInput';
import ChatMessage from './ChatMessage';
import type { ChatPanelProps, ChatMessage as ChatMessageType } from '../types/chat';
import { processChatQuery, clearChatContext, ChatServiceError } from '../services/chat';

const STORAGE_KEY = 'chatPanelCollapsed';

const ChatPanel: FC<ChatPanelProps> = ({
    initialCollapsed = false,
    onCollapseChange,
}) => {
    // FR-009: Restore collapsed state from sessionStorage
    const [isCollapsed, setIsCollapsed] = useState(() => {
        try {
            const stored = sessionStorage.getItem(STORAGE_KEY);
            if (stored !== null) {
                return stored === 'true';
            }
        } catch (e) {
            // SessionStorage not available (e.g., in tests or private browsing)
            console.warn('sessionStorage not available:', e);
        }
        return initialCollapsed;
    });

    // US2: Conversation state management
    const [messages, setMessages] = useState<ChatMessageType[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [retryable, setRetryable] = useState(false);
    const [lastQuery, setLastQuery] = useState<string | null>(null);
    const [currentInput, setCurrentInput] = useState<string>('');
    const [isClearing, setIsClearing] = useState(false);
    const [showClearConfirmation, setShowClearConfirmation] = useState(false);
    const conversationRef = useRef<HTMLDivElement>(null);

    // FR-009: Persist collapsed state to sessionStorage
    useEffect(() => {
        try {
            sessionStorage.setItem(STORAGE_KEY, String(isCollapsed));
        } catch (e) {
            // SessionStorage not available
            console.warn('Failed to save to sessionStorage:', e);
        }
        onCollapseChange?.(isCollapsed);
    }, [isCollapsed, onCollapseChange]);

    // FR-014: Auto-scroll to latest message
    useEffect(() => {
        if (conversationRef.current && !isLoading) {
            conversationRef.current.scrollTo({
                top: conversationRef.current.scrollHeight,
                behavior: 'smooth'
            });
        }
    }, [messages, isLoading]);

    const handleToggleCollapse = () => {
        setIsCollapsed(!isCollapsed);
    };

    // US2: Query submission handler
    const handleSubmit = async (query: string) => {
        // Auto-expand panel when user submits a query
        if (isCollapsed) {
            setIsCollapsed(false);
        }

        // Clear any previous error
        setError(null);
        setLastQuery(query);
        setIsLoading(true);
        setCurrentInput(query); // Save for restoration on error

        // Add user message to conversation
        const userMessage: ChatMessageType = {
            id: `user-${Date.now()}-${Math.random()}`,
            text: query,
            sender: 'user',
            timestamp: new Date()
        };
        setMessages(prev => [...prev, userMessage]);

        try {
            // FR-024: Call chat service with timeout handling
            const response = await processChatQuery(query);

            // Add system response to conversation
            const systemMessage: ChatMessageType = {
                id: `system-${Date.now()}-${Math.random()}`,
                text: response,
                sender: 'system',
                timestamp: new Date()
            };
            setMessages(prev => [...prev, systemMessage]);
            setCurrentInput(''); // Clear input on success
        } catch (err) {
            // FR-011: User-friendly error handling
            const chatError = err as ChatServiceError;
            let errorMessage = chatError.message || 'An error occurred while processing your query';

            // Ensure error message contains "error" or "failed" for accessibility and testing
            if (!/error|failed?/i.test(errorMessage)) {
                errorMessage = `Query failed: ${errorMessage}`;
            }

            setError(errorMessage);
            setRetryable(chatError.canRetry ?? true);
            // Keep currentInput so it can be restored
        } finally {
            setIsLoading(false);
        }
    };

    // US4: Retry functionality
    const handleRetry = () => {
        if (lastQuery) {
            setError(null); // Clear error before retry
            handleSubmit(lastQuery);
        }
    };

    // US4: Clear conversation handler - show confirmation
    const handleClear = () => {
        console.log('Clear button clicked, messages.length:', messages.length);
        if (messages.length > 0) {
            setShowClearConfirmation(true);
        }
    };

    // US4: Actual clear logic after confirmation
    const confirmClear = async () => {
        console.log('Clearing conversation...');
        setShowClearConfirmation(false);
        setIsClearing(true);
        try {
            // T051: Call clearChatContext via chat service
            await clearChatContext();
            console.log('clearChatContext succeeded');
            // T052: Reset local conversation state
            setMessages([]);
            setError(null);
            setLastQuery(null);
        } catch (err) {
            const chatError = err as ChatServiceError;
            console.error('clearChatContext failed:', err);
            setError(chatError.message || 'Failed to clear conversation');
        } finally {
            setIsClearing(false);
        }
    };

    // US4: Cancel clear operation
    const cancelClear = () => {
        setShowClearConfirmation(false);
    };

    const panelClasses = [
        'chat-panel',
        'chat-panel--bottom',
        isCollapsed && 'chat-panel--collapsed'
    ].filter(Boolean).join(' ');

    return (
        <div
            className={panelClasses}
            role="region"
            aria-label="Chat interface"
            aria-expanded={!isCollapsed}
        >
            <div className="chat-panel__body" style={{
                display: isCollapsed ? 'none' : 'flex',
                flexDirection: 'column',
                height: isCollapsed ? '0' : '350px',
                opacity: isCollapsed ? 0 : 1,
                overflow: 'hidden',
                transition: 'height 0.3s ease, opacity 0.3s ease'
            }}>
                <div style={{ position: 'relative', flex: 1, minHeight: 0, display: 'flex', flexDirection: 'column' }}>
                    <div
                        className="chat-panel__conversation"
                        role="log"
                        ref={conversationRef}
                        style={{
                            overflowY: 'auto',
                            flex: 1,
                            minHeight: 0,
                            paddingBottom: '40px'
                        }}
                    >
                        {messages.length === 0 ? (
                            <div
                                className="chat-panel__empty"
                                role="status"
                                aria-live="polite"
                            >
                                Start a conversation by asking a question about the graph...
                            </div>
                        ) : (
                            <>
                                {messages.map((msg) => (
                                    <ChatMessage key={msg.id} message={msg} />
                                ))}
                            </>
                        )}

                        {/* FR-011: Loading indicator */}
                        {isLoading && (
                            <div className="chat-panel__loading" role="status" aria-live="polite">
                                <div className="chat-panel__loading-spinner"></div>
                                <span>Processing your query...</span>
                            </div>
                        )}

                        {/* FR-011, FR-022: Error display with retry */}
                        {error && (
                            <div className="chat-panel__error" role="alert">
                                <span className="chat-panel__error-message">{error}</span>
                                {retryable && (
                                    <button
                                        className="chat-panel__error-retry"
                                        onClick={handleRetry}
                                        aria-label="Retry last query"
                                    >
                                        Retry
                                    </button>
                                )}
                            </div>
                        )}
                    </div>

                    {/* T050: Clear button - positioned absolutely at bottom right */}
                    {messages.length > 0 && (
                        <div className="chat-panel__clear-button-container">
                            <button
                                className="chat-panel__clear-button"
                                onClick={handleClear}
                                disabled={isClearing}
                                aria-label="Clear conversation"
                                title="Clear conversation history"
                            >
                                Clear
                            </button>
                        </div>
                    )}
                </div>
            </div>

            <div className="chat-panel__footer">
                <div style={{ flex: 1 }}>
                    <ChatInput
                        onSubmit={handleSubmit}
                        placeholder="Ask about the graph..."
                        disabled={isLoading} // FR-021: Disable while loading
                        value={currentInput}
                        onChange={setCurrentInput}
                    />
                </div>
                <button
                    className="chat-panel__toggle-compact"
                    onClick={handleToggleCollapse}
                    aria-label={isCollapsed ? 'Expand chat panel' : 'Collapse chat panel'}
                    title={isCollapsed ? 'Expand chat' : 'Collapse chat'}
                >
                    {isCollapsed ? '▲' : '▼'}
                </button>
            </div>

            {/* US4: Clear confirmation dialog */}
            {showClearConfirmation && (
                <div
                    className="chat-panel__confirm-overlay"
                    onClick={cancelClear}
                    role="dialog"
                    aria-modal="true"
                    aria-labelledby="clear-dialog-title"
                    aria-describedby="clear-dialog-description"
                >
                    <div className="chat-panel__confirm-dialog" onClick={(e) => e.stopPropagation()}>
                        <h3
                            id="clear-dialog-title"
                            className="chat-panel__confirm-title"
                        >
                            Clear Conversation
                        </h3>
                        <p
                            id="clear-dialog-description"
                            className="chat-panel__confirm-message"
                        >
                            Are you sure you want to clear the conversation history? This action cannot be undone.
                        </p>
                        <div className="chat-panel__confirm-buttons">
                            <button
                                className="chat-panel__confirm-button chat-panel__confirm-button--cancel"
                                onClick={cancelClear}
                                autoFocus
                                aria-label="Cancel clearing conversation"
                            >
                                Cancel
                            </button>
                            <button
                                className="chat-panel__confirm-button chat-panel__confirm-button--confirm"
                                onClick={confirmClear}
                                aria-label="Confirm clearing conversation"
                            >
                                Clear
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ChatPanel;
