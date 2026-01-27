
/** @jsxImportSource react */
import { FC } from 'react';
import './ChatMessage.css';
import type { ChatMessageProps } from '../types/chat';

const ChatMessage: FC<ChatMessageProps> = ({ message }) => {
    const { text, sender, timestamp } = message;

    // Format timestamp for display
    const formatTime = (date: Date) => {
        return date.toLocaleTimeString('en-US', {
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const senderLabel = sender === 'user' ? 'You' : 'Assistant';

    return (
        <div
            className={`chat-message chat-message--${sender}`}
            role="article"
            aria-label={`${senderLabel} message`}
        >
            <div className="chat-message__content">
                {text}
                <div className="chat-message__timestamp">{formatTime(timestamp)}</div>
            </div>
        </div>
    );
};

export default ChatMessage;
