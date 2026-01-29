/** @jsxImportSource react */
import { FC, useState, KeyboardEvent } from 'react';
import './ChatInput.css';
import type { ChatInputProps } from '../types/chat';

const ChatInput: FC<ChatInputProps> = ({
    onSubmit,
    disabled = false,
    placeholder = 'Ask about the graph...',
    value: externalValue,
    onChange: externalOnChange,
    onHistoryNavigate,
}) => {
    const [internalQuery, setInternalQuery] = useState('');
    const query = externalValue !== undefined ? externalValue : internalQuery;
    const setQuery = externalOnChange !== undefined ? externalOnChange : setInternalQuery;

    const handleSubmit = () => {
        const trimmedQuery = query.trim();
        // FR-015: Prevent empty query submission
        if (!trimmedQuery || disabled) {
            return;
        }
        onSubmit(trimmedQuery);
        // Only clear in uncontrolled mode
        if (externalValue === undefined) {
            setQuery('');
        }
    };

    const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
        // FR-003: Enter key submits query
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit();
        }
        // FR-004: Shift+Enter creates newline (default behavior, just don't prevent)
        
        // History navigation
        if (e.key === 'ArrowUp') {
            e.preventDefault();
            onHistoryNavigate?.('up');
        } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            onHistoryNavigate?.('down');
        }
    };

    const isButtonDisabled = disabled || !query.trim();

    return (
        <div className="chat-input">
            <textarea
                className="chat-input__field"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={placeholder}
                disabled={disabled}
                rows={1}
                aria-label="Chat query input"
            />
            <button
                className="chat-input__button"
                onClick={handleSubmit}
                disabled={isButtonDisabled}
                aria-label="Send message"
            >
                Send
            </button>
        </div>
    );
};

export default ChatInput;
