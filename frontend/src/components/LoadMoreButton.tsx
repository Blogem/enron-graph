import React from 'react';
import './LoadMoreButton.css';

interface LoadMoreButtonProps {
    nodeId: string;
    loaded: number;
    total: number;
    batchSize: number;
    onLoadMore: (nodeId: string) => void;
    loading?: boolean;
}

const LoadMoreButton: React.FC<LoadMoreButtonProps> = ({
    nodeId,
    loaded,
    total,
    batchSize,
    onLoadMore,
    loading = false
}) => {
    const remaining = total - loaded;
    const nextBatch = Math.min(batchSize, remaining);

    const handleClick = () => {
        if (!loading) {
            onLoadMore(nodeId);
        }
    };

    if (remaining <= 0) {
        return null;
    }

    return (
        <button
            className="load-more-button"
            onClick={handleClick}
            disabled={loading}
            title={`Load ${nextBatch} more relationships (${remaining} remaining)`}
        >
            {loading ? (
                <>
                    <span className="spinner"></span>
                    Loading...
                </>
            ) : (
                <>
                    ⬇ Load {nextBatch} more ({remaining} remaining)
                </>
            )}
        </button>
    );
};

export default LoadMoreButton;
