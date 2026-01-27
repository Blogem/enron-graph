import React, { useState, useEffect } from 'react';
import './KeyboardHelp.css';

const KeyboardHelp: React.FC = () => {
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

    return (
        <div className="keyboard-help">
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
