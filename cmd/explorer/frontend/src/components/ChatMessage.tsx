
/** @jsxImportSource react */
import { FC } from 'react';
import './ChatMessage.css';
import type { ChatMessageProps } from '../types/chat';

const ChatMessage: FC<ChatMessageProps> = ({ message, onEntityClick }) => {
    const { text, sender, timestamp, entities } = message;

    // Format timestamp for display
    const formatTime = (date: Date) => {
        return date.toLocaleTimeString('en-US', {
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const senderLabel = sender === 'user' ? 'You' : 'Assistant';

    // Render message text with clickable entity references
    const renderMessageContent = () => {
        // If no entities, render plain text (preserving newlines via CSS white-space: pre-wrap)
        if (!entities || entities.length === 0) {
            return <>{text}</>;
        }

        // Build a list of all entity matches in the entire text
        interface EntityMatch {
            entity: typeof entities[0];
            startIndex: number;
            endIndex: number;
        }

        const matches: EntityMatch[] = [];

        entities.forEach((entity) => {
            const escapedName = entity.name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            const regex = new RegExp(`\\b${escapedName}\\b`, 'g');
            let match;

            while ((match = regex.exec(text)) !== null) {
                matches.push({
                    entity,
                    startIndex: match.index,
                    endIndex: match.index + match[0].length
                });
            }
        });

        // Sort matches by start index
        matches.sort((a, b) => a.startIndex - b.startIndex);

        // Remove overlapping matches (keep first occurrence)
        const nonOverlapping: EntityMatch[] = [];
        let lastEnd = -1;
        for (const match of matches) {
            if (match.startIndex >= lastEnd) {
                nonOverlapping.push(match);
                lastEnd = match.endIndex;
            }
        }

        // If no matches found, return text as-is
        if (nonOverlapping.length === 0) {
            return <>{text}</>;
        }

        // Build the rendered content with clickable entities
        const parts: (string | JSX.Element)[] = [];
        let currentIndex = 0;

        nonOverlapping.forEach((match, idx) => {
            // Add text before this match
            if (match.startIndex > currentIndex) {
                const textSegment = text.substring(currentIndex, match.startIndex);
                // Split by newlines and add <br /> elements
                textSegment.split('\n').forEach((line, lineIdx, arr) => {
                    parts.push(line);
                    if (lineIdx < arr.length - 1) {
                        parts.push(<br key={`br-before-${idx}-${lineIdx}`} />);
                    }
                });
            }

            // Add clickable entity
            parts.push(
                <span
                    key={`entity-${idx}`}
                    className="chat-message__entity-link"
                    onClick={() => {
                        console.log('Entity clicked:', {
                            name: match.entity.name,
                            type: match.entity.type,
                            id: match.entity.id,
                            unique_id: match.entity.unique_id,
                            fullEntity: match.entity
                        });
                        if (!match.entity.unique_id) {
                            console.error('Entity unique_id is empty or undefined!', match.entity);
                        }
                        onEntityClick?.(match.entity.unique_id);
                    }}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            if (!match.entity.unique_id) {
                                console.error('Entity unique_id is empty or undefined!', match.entity);
                            }
                            onEntityClick?.(match.entity.unique_id);
                        }
                    }}
                    title={`Click to view ${match.entity.name} (${match.entity.type})`}
                >
                    {match.entity.name}
                </span>
            );

            currentIndex = match.endIndex;
        });

        // Add remaining text after last match
        if (currentIndex < text.length) {
            const textSegment = text.substring(currentIndex);
            textSegment.split('\n').forEach((line, lineIdx, arr) => {
                parts.push(line);
                if (lineIdx < arr.length - 1) {
                    parts.push(<br key={`br-after-${lineIdx}`} />);
                }
            });
        }

        return <>{parts}</>;
    };

    return (
        <div
            className={`chat-message chat-message--${sender}`}
            role="article"
            aria-label={`${senderLabel} message`}
        >
            <div className="chat-message__content">
                {renderMessageContent()}
                <div className="chat-message__timestamp">{formatTime(timestamp)}</div>
            </div>
        </div>
    );
};

export default ChatMessage;
