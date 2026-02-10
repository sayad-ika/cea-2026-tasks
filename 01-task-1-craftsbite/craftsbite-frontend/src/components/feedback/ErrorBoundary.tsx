import React from 'react';

interface ErrorBoundaryProps {
    children: React.ReactNode;
    fallback?: React.ReactNode;
}

interface ErrorBoundaryState {
    hasError: boolean;
    error: Error | null;
}

/**
 * Global error boundary — catches React render errors and displays a
 * claymorphism-styled fallback instead of a blank white screen.
 */
export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
    constructor(props: ErrorBoundaryProps) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error: Error): ErrorBoundaryState {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, info: React.ErrorInfo) {
        console.error('[ErrorBoundary]', error, info.componentStack);
    }

    handleReset = () => {
        this.setState({ hasError: false, error: null });
    };

    render() {
        if (this.state.hasError) {
            if (this.props.fallback) return this.props.fallback;

            return (
                <div className="min-h-screen flex items-center justify-center bg-[var(--color-background-light)] font-display px-6">
                    <div
                        className="max-w-md w-full bg-[var(--color-background-light)] rounded-3xl p-10 text-center"
                        style={{ boxShadow: 'var(--shadow-clay)' }}
                    >
                        <span className="material-symbols-outlined text-6xl text-red-400 mb-4 block">
                            error
                        </span>
                        <h2 className="text-2xl font-black text-[var(--color-background-dark)] mb-2">
                            Something went wrong
                        </h2>
                        <p className="text-sm text-[var(--color-text-sub)] mb-6">
                            {this.state.error?.message || 'An unexpected error occurred.'}
                        </p>
                        <button
                            onClick={this.handleReset}
                            className="px-8 py-3 rounded-2xl bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-primary-dark)] text-white font-bold hover:scale-105 active:scale-95 transition-all duration-200"
                            style={{ boxShadow: 'var(--shadow-clay-button)' }}
                        >
                            Try Again
                        </button>
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}
