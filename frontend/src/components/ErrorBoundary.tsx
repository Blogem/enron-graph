import React, { Component, ReactNode } from 'react';

interface Props {
    children: ReactNode;
}

interface State {
    hasError: boolean;
    error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
        console.error('ErrorBoundary caught an error:', error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            return (
                <div style={{
                    padding: '20px',
                    background: '#ff6b6b',
                    color: 'white',
                    borderRadius: '8px',
                    margin: '20px'
                }}>
                    <h2>Something went wrong</h2>
                    <pre style={{
                        background: 'rgba(0,0,0,0.2)',
                        padding: '10px',
                        borderRadius: '4px',
                        overflow: 'auto',
                        fontSize: '12px'
                    }}>
                        {this.state.error?.toString()}
                        {'\n\n'}
                        {this.state.error?.stack}
                    </pre>
                    <button
                        onClick={() => window.location.reload()}
                        style={{
                            marginTop: '10px',
                            padding: '10px 20px',
                            background: 'white',
                            color: '#ff6b6b',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer'
                        }}
                    >
                        Reload Page
                    </button>
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
