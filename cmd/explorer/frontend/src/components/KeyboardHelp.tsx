import React, { useState, useEffect } from 'react';
import './KeyboardHelp.css';

interface KeyboardHelpProps {
    chatPanelCollapsed?: boolean;
}

const KeyboardHelp: React.FC<KeyboardHelpProps> = ({ chatPanelCollapsed = false }) => {
    const [isOpen, setIsOpen] = useState(false);

    // Close on Escape key
    useEffect(() => {
        if (!isOpen) return;

        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                setIsOpen(false);
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [isOpen]);

    // Calculate bottom position based on chat panel state
    // Chat panel footer (input) is ~60px, expanded body is ~350px, add 20px margin
    const bottomOffset = chatPanelCollapsed ? 90 : 440; // collapsed: footer + margin, expanded: footer + body + margin

    return (
        <div className="keyboard-help" style={{ bottom: `${bottomOffset}px` }}>
            <button
                className="keyboard-help-button"
                onClick={() => setIsOpen(!isOpen)}
                title="Keyboard shortcuts"
                aria-label="Show keyboard shortcuts"
            >
                ⌨️
            </button>
            {isOpen && (
                <div className="keyboard-help-panel">
                    <div className="keyboard-help-header">
                        <h3>Keyboard Shortcuts</h3>
                        <button
                            className="close-button"
                            onClick={() => setIsOpen(false)}
                            aria-label="Close"
                        >
                            ✕
                        </button>
                    </div>
                    <div className="keyboard-help-content">
                        <div className="shortcut-item">
                            <kbd>Esc</kbd>
                            <span>Clear selection</span>
                        </div>
                        <div className="shortcut-item">
                            <kbd>Space</kbd>
                            <span>Recenter graph</span>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default KeyboardHelp;
