/** @jsxImportSource react */
import { FC, useState, useEffect } from 'react';
import './ChatPanel.css';
import ChatInput from './ChatInput';
import ChatMessage from './ChatMessage';
import type { ChatPanelProps } from '../types/chat';

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

    // Placeholder state - will be extended in US2
    const [messages] = useState<any[]>([]);

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

    const handleToggleCollapse = () => {
        setIsCollapsed(!isCollapsed);
    };

    const handleSubmit = (query: string) => {
        // Auto-expand panel when user submits a query
        if (isCollapsed) {
            setIsCollapsed(false);
        }
        // Placeholder - will be implemented in US2
        console.log('Query submitted:', query);
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
                display: 'flex',
                flexDirection: 'column',
                height: isCollapsed ? '0' : '350px',
                opacity: isCollapsed ? 0 : 1,
                overflow: 'hidden',
                transition: 'height 0.3s ease, opacity 0.3s ease'
            }}>
                <div className="chat-panel__conversation" role="log" style={{
                    overflowY: 'auto',
                    flex: 1,
                    minHeight: 0
                }}>
                    {messages.length === 0 ? (
                        <div className="chat-panel__empty">
                            Start a conversation by asking a question about the graph...
                        </div>
                    ) : (
                        messages.map((msg) => (
                            <ChatMessage key={msg.id} message={msg} />
                        ))
                    )}
                </div>
            </div>

            <div className="chat-panel__footer">
                <ChatInput onSubmit={handleSubmit} placeholder="Ask about the graph..." />
                <button
                    className="chat-panel__toggle-compact"
                    onClick={handleToggleCollapse}
                    aria-label={isCollapsed ? 'Expand chat panel' : 'Collapse chat panel'}
                    title={isCollapsed ? 'Expand chat' : 'Collapse chat'}
                >
                    {isCollapsed ? '▲' : '▼'}
                </button>
            </div>
        </div>
    );
};

export default ChatPanel;
