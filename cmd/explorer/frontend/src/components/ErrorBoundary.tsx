import React, { Component, ErrorInfo, ReactNode } from 'react';
import './ErrorBoundary.css';

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
    onError?: (error: Error, errorInfo: ErrorInfo) => void;
    resetKeys?: Array<string | number>;
    componentName?: string;
}

interface State {
    hasError: boolean;
    error: Error | null;
    errorInfo: ErrorInfo | null;
}

/**
 * Error Boundary Component (T106)
 * 
 * Catches JavaScript errors anywhere in the child component tree,
 * logs those errors, and displays a fallback UI instead of crashing the whole app.
 * 
 * Usage:
 * ```tsx
 * <ErrorBoundary componentName="SchemaPanel">
 *   <SchemaPanel />
 * </ErrorBoundary>
 * ```
 */
class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = {
            hasError: false,
            error: null,
            errorInfo: null
        };
    }

    static getDerivedStateFromError(error: Error): Partial<State> {
        // Update state so the next render shows the fallback UI
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        // Log error details
        const componentName = this.props.componentName || 'Unknown Component';
        console.error(`Error Boundary [${componentName}] caught an error:`, error, errorInfo);

        // Update state with error details
        this.setState({
            error,
            errorInfo
        });

        // Call optional error handler
        if (this.props.onError) {
            this.props.onError(error, errorInfo);
        }
    }

    componentDidUpdate(prevProps: Props) {
        // Reset error state when resetKeys change (allows recovery without full reload)
        if (this.state.hasError && this.props.resetKeys) {
            const prevKeys = prevProps.resetKeys || [];
            const currentKeys = this.props.resetKeys;

            if (prevKeys.length !== currentKeys.length ||
                prevKeys.some((key, idx) => key !== currentKeys[idx])) {
                this.handleReset();
            }
        }
    }

    handleReset = () => {
        this.setState({
            hasError: false,
            error: null,
            errorInfo: null
        });
    };

    render() {
        if (this.state.hasError) {
            // Custom fallback UI if provided
            if (this.props.fallback) {
                return this.props.fallback;
            }

            // Default fallback UI
            const componentName = this.props.componentName || 'This component';

            return (
                <div className="error-boundary-container">
                    <div className="error-boundary-content">
                        <div className="error-icon">⚠️</div>
                        <h2>Something went wrong</h2>
                        <p className="error-component-name">
                            {componentName} encountered an error
                        </p>
                        <p className="error-message">
                            {this.state.error?.message || 'An unexpected error occurred'}
                        </p>
                        <details className="error-details">
                            <summary>Error Details (for debugging)</summary>
                            <div className="error-stack-container">
                                <h4>Stack Trace:</h4>
                                <pre className="error-stack">
                                    {this.state.error?.stack}
                                </pre>
                                {this.state.errorInfo && (
                                    <>
                                        <h4>Component Stack:</h4>
                                        <pre className="error-info">
                                            {this.state.errorInfo.componentStack}
                                        </pre>
                                    </>
                                )}
                            </div>
                        </details>
                        <div className="error-actions">
                            <button onClick={this.handleReset} className="retry-button">
                                Try Again
                            </button>
                            <button onClick={() => window.location.reload()} className="reload-button">
                                Reload Page
                            </button>
                        </div>
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
