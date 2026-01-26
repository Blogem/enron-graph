import React from 'react';
import type { GraphNodeWithPosition } from '../types/graph';
import LoadMoreButton from './LoadMoreButton';
import './DetailPanel.css';

interface DetailPanelProps {
    node: GraphNodeWithPosition | null;
    loading?: boolean;
    expandedNodeState?: {
        offset: number;
        hasMore: boolean;
        totalRelationships: number;
    } | null;
    onLoadMore?: (nodeId: string) => void;
    onClose?: () => void;
}

const DetailPanel: React.FC<DetailPanelProps> = ({
    node,
    loading = false,
    expandedNodeState = null,
    onLoadMore,
    onClose
}) => {
    const copyToClipboard = async (text: string, successMessage: string) => {
        try {
            await navigator.clipboard.writeText(text);
            // TODO: Add toast notification for success
            console.log(successMessage);
        } catch (err) {
            console.error('Failed to copy:', err);
            // Fallback for older browsers
            const textArea = document.createElement('textarea');
            textArea.value = text;
            textArea.style.position = 'fixed';
            textArea.style.left = '-999999px';
            document.body.appendChild(textArea);
            textArea.select();
            try {
                document.execCommand('copy');
                console.log(successMessage);
            } catch (e) {
                console.error('Fallback copy failed:', e);
            }
            document.body.removeChild(textArea);
        }
    };

    const handleCopyId = () => {
        if (node) {
            copyToClipboard(node.id, 'Node ID copied to clipboard');
        }
    };

    const handleCopyAsJson = () => {
        if (node) {
            const jsonData = {
                id: node.id,
                type: node.type,
                label: node.label,
                properties: node.properties
            };
            copyToClipboard(JSON.stringify(jsonData, null, 2), 'Node data copied as JSON');
        }
    };

    if (!node) {
        return (
            <div className="detail-panel empty">
                <p className="empty-message">Select a node to view details</p>
                <p className="empty-hint">Click on a node in the graph or right-click to expand relationships</p>
            </div>
        );
    }

    const renderPropertyValue = (value: any): string => {
        if (value === null || value === undefined) {
            return 'null';
        }
        if (typeof value === 'object') {
            return JSON.stringify(value, null, 2);
        }
        return String(value);
    };

    return (
        <div className="detail-panel">
            <div className="detail-header">
                <div className="detail-title">
                    <h2>{node.label || node.id}</h2>
                    <span className={`type-badge ${node.type.toLowerCase()}`}>
                        {node.type}
                    </span>
                </div>
                {onClose && (
                    <button className="close-button" onClick={onClose} title="Close details (Escape)">
                        ✕
                    </button>
                )}
            </div>

            {loading && (
                <div className="detail-loading">
                    <div className="spinner-large"></div>
                    <p>Loading details...</p>
                </div>
            )}

            {!loading && (
                <div className="detail-content">
                    <div className="detail-section">
                        <h3>Node Information</h3>
                        <div className="info-grid">
                            <div className="info-item">
                                <span className="info-label">ID:</span>
                                <span className="info-value">{node.id}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Type:</span>
                                <span className="info-value">{node.type}</span>
                            </div>
                            {node.properties?.degree !== undefined && (
                                <div className="info-item">
                                    <span className="info-label">Relationships:</span>
                                    <span className="info-value">
                                        {node.properties.degree}
                                    </span>
                                </div>
                            )}
                        </div>
                    </div>

                    {node.properties && Object.keys(node.properties).length > 0 && (
                        <div className="detail-section">
                            <h3>Properties</h3>
                            <div className="properties-list">
                                {Object.entries(node.properties)
                                    .filter(([key]) => key !== 'degree')
                                    .map(([key, value]) => (
                                        <div key={key} className="property-item">
                                            <div className="property-key">{key}</div>
                                            <div className="property-value">
                                                <code>{renderPropertyValue(value)}</code>
                                            </div>
                                        </div>
                                    ))}
                            </div>
                        </div>
                    )}

                    {expandedNodeState && (
                        <div className="detail-section">
                            <h3>Relationships</h3>
                            <div className="relationships-info">
                                <p>
                                    Loaded {expandedNodeState.offset} of {expandedNodeState.totalRelationships} relationships
                                </p>
                                {expandedNodeState.hasMore && onLoadMore && (
                                    <LoadMoreButton
                                        nodeId={node.id}
                                        loaded={expandedNodeState.offset}
                                        total={expandedNodeState.totalRelationships}
                                        batchSize={50}
                                        onLoadMore={onLoadMore}
                                    />
                                )}
                            </div>
                        </div>
                    )}

                    <div className="detail-section">
                        <h3>Actions</h3>
                        <div className="action-buttons">
                            <button className="action-button" onClick={handleCopyId}>
                                📋 Copy ID
                            </button>
                            <button className="action-button" onClick={handleCopyAsJson}>
                                🔗 Copy as JSON
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default DetailPanel;
