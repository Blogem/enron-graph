import React from 'react';
import './LoadingSkeleton.css';

interface LoadingSkeletonProps {
    type?: 'schema' | 'graph' | 'detail' | 'text' | 'card';
    count?: number;
}

/**
 * Loading Skeleton Component (T107)
 * 
 * Provides visual placeholders during content loading to improve perceived performance
 * and user experience. Supports different skeleton types for various UI components.
 */
export const LoadingSkeleton: React.FC<LoadingSkeletonProps> = ({
    type = 'text',
    count = 1
}) => {
    const renderSkeleton = () => {
        switch (type) {
            case 'schema':
                return <SchemaSkeleton />;
            case 'graph':
                return <GraphSkeleton />;
            case 'detail':
                return <DetailSkeleton />;
            case 'card':
                return <CardSkeleton count={count} />;
            case 'text':
            default:
                return <TextSkeleton count={count} />;
        }
    };

    return <>{renderSkeleton()}</>;
};

/**
 * Schema Panel Loading Skeleton
 */
const SchemaSkeleton: React.FC = () => {
    return (
        <div className="schema-panel">
            <div className="schema-header">
                <div className="skeleton skeleton-title"></div>
                <div className="skeleton skeleton-button"></div>
            </div>
            <div className="schema-skeleton-content">
                <div className="skeleton skeleton-section-header"></div>
                <div className="skeleton-type-list">
                    {[1, 2, 3, 4].map(i => (
                        <div key={i} className="skeleton-type-item">
                            <div className="skeleton skeleton-type-icon"></div>
                            <div className="skeleton-type-info">
                                <div className="skeleton skeleton-type-name"></div>
                                <div className="skeleton skeleton-type-count"></div>
                            </div>
                        </div>
                    ))}
                </div>
                <div className="skeleton skeleton-section-header"></div>
                <div className="skeleton-type-list">
                    {[1, 2, 3].map(i => (
                        <div key={i} className="skeleton-type-item">
                            <div className="skeleton skeleton-type-icon"></div>
                            <div className="skeleton-type-info">
                                <div className="skeleton skeleton-type-name"></div>
                                <div className="skeleton skeleton-type-count"></div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

/**
 * Graph Canvas Loading Skeleton
 */
const GraphSkeleton: React.FC = () => {
    return (
        <div className="graph-skeleton">
            <div className="skeleton-graph-overlay">
                <div className="skeleton-spinner"></div>
                <div className="skeleton skeleton-graph-text"></div>
            </div>
            {/* Simulate node positions */}
            <div className="skeleton-graph-nodes">
                {Array.from({ length: 15 }).map((_, i) => (
                    <div
                        key={i}
                        className="skeleton skeleton-node"
                        style={{
                            left: `${Math.random() * 80 + 10}%`,
                            top: `${Math.random() * 80 + 10}%`,
                            animationDelay: `${Math.random() * 0.5}s`
                        }}
                    ></div>
                ))}
            </div>
        </div>
    );
};

/**
 * Detail Panel Loading Skeleton
 */
const DetailSkeleton: React.FC = () => {
    return (
        <div className="detail-skeleton">
            <div className="skeleton-detail-header">
                <div className="skeleton skeleton-detail-title"></div>
                <div className="skeleton skeleton-detail-badge"></div>
            </div>
            <div className="skeleton-detail-content">
                <div className="skeleton skeleton-section-title"></div>
                {[1, 2, 3, 4, 5].map(i => (
                    <div key={i} className="skeleton-property">
                        <div className="skeleton skeleton-property-key"></div>
                        <div className="skeleton skeleton-property-value"></div>
                    </div>
                ))}
            </div>
        </div>
    );
};

/**
 * Card-style Loading Skeleton
 */
const CardSkeleton: React.FC<{ count: number }> = ({ count }) => {
    return (
        <>
            {Array.from({ length: count }).map((_, i) => (
                <div key={i} className="skeleton-card">
                    <div className="skeleton skeleton-card-header"></div>
                    <div className="skeleton skeleton-card-body"></div>
                    <div className="skeleton skeleton-card-footer"></div>
                </div>
            ))}
        </>
    );
};

/**
 * Text-style Loading Skeleton
 */
const TextSkeleton: React.FC<{ count: number }> = ({ count }) => {
    return (
        <>
            {Array.from({ length: count }).map((_, i) => (
                <div key={i} className="skeleton skeleton-text"></div>
            ))}
        </>
    );
};

export default LoadingSkeleton;
