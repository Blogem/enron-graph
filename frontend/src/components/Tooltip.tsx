import React, { useState, useRef, useEffect } from 'react';
import './Tooltip.css';

interface TooltipProps {
    content: string | React.ReactNode;
    children: React.ReactElement;
    position?: 'top' | 'bottom' | 'left' | 'right';
    delay?: number;
    disabled?: boolean;
}

/**
 * Tooltip Component (T108)
 * 
 * Provides accessible tooltips for UI controls and interactive elements.
 * Supports multiple positions, custom delay, and keyboard accessibility.
 * 
 * Usage:
 * ```tsx
 * <Tooltip content="Click to refresh schema">
 *   <button>Refresh</button>
 * </Tooltip>
 * ```
 */
export const Tooltip: React.FC<TooltipProps> = ({
    content,
    children,
    position = 'top',
    delay = 300,
    disabled = false
}) => {
    const [isVisible, setIsVisible] = useState(false);
    const timeoutRef = useRef<number | null>(null);

    const handleMouseEnter = () => {
        if (disabled) return;

        timeoutRef.current = window.setTimeout(() => {
            setIsVisible(true);
        }, delay);
    };

    const handleMouseLeave = () => {
        if (timeoutRef.current) {
            window.clearTimeout(timeoutRef.current);
            timeoutRef.current = null;
        }
        setIsVisible(false);
    };

    const handleFocus = () => {
        if (disabled) return;
        setIsVisible(true);
    };

    const handleBlur = () => {
        setIsVisible(false);
    };

    useEffect(() => {
        return () => {
            if (timeoutRef.current) {
                window.clearTimeout(timeoutRef.current);
            }
        };
    }, []);

    // Clone child element to add event handlers
    const childWithHandlers = React.cloneElement(children, {
        onMouseEnter: handleMouseEnter,
        onMouseLeave: handleMouseLeave,
        onFocus: handleFocus,
        onBlur: handleBlur,
        'aria-describedby': isVisible ? 'tooltip' : undefined
    });

    return (
        <div className="tooltip-wrapper">
            {childWithHandlers}
            {isVisible && !disabled && (
                <div
                    className={`tooltip tooltip-${position}`}
                    role="tooltip"
                    id="tooltip"
                >
                    {content}
                    <div className={`tooltip-arrow tooltip-arrow-${position}`}></div>
                </div>
            )}
        </div>
    );
};

export default Tooltip;
