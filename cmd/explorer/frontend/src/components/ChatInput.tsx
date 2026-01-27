/** @jsxImportSource react */
import { FC, useState, KeyboardEvent } from 'react';
import './ChatInput.css';
import type { ChatInputProps } from '../types/chat';

const ChatInput: FC<ChatInputProps> = ({
    onSubmit,
    disabled = false,
    placeholder = 'Ask about the graph...',
}) => {
    const [query, setQuery] = useState('');

    const handleSubmit = () => {
        const trimmedQuery = query.trim();
        // FR-015: Prevent empty query submission
        if (!trimmedQuery || disabled) {
            return;
        }
        onSubmit(trimmedQuery);
        setQuery('');
    };

    const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
        // FR-003: Enter key submits query
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit();
        }
        // FR-004: Shift+Enter creates newline (default behavior, just don't prevent)
    };

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
                disabled={disabled}
                aria-label="Send message"
            >
                Send
            </button>
        </div>
    );
};

export default ChatInput;
