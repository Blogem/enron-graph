import React, { useState, useMemo } from 'react';
import type { GraphNodeWithPosition, GraphEdge } from '../types/graph';
import LoadMoreButton from './LoadMoreButton';
import LoadingSkeleton from './LoadingSkeleton';
import Tooltip from './Tooltip';
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
    relatedEntities?: Array<{
        edge: GraphEdge;
        node: GraphNodeWithPosition;
    }>;
    onExpandRelationship?: (nodeId: string) => void;
}

const DetailPanel: React.FC<DetailPanelProps> = ({
    node,
    loading = false,
    expandedNodeState = null,
    onLoadMore,
    onClose,
    relatedEntities = [],
    onExpandRelationship
}) => {
    // Collapsible section states
    const [sectionsExpanded, setSectionsExpanded] = useState({
        properties: true,
        metadata: true,
        relationships: true
    });

    const toggleSection = (section: keyof typeof sectionsExpanded) => {
        setSectionsExpanded(prev => ({
            ...prev,
            [section]: !prev[section]
        }));
    };

    // Categorize properties into user-defined and system metadata
    const { userProperties, metadata } = useMemo(() => {
        if (!node?.properties) {
            return { userProperties: {}, metadata: {} };
        }

        const metadataKeys = [
            'degree', 'created_at', 'updated_at', 'discovered_at',
            'confidence', 'source', 'category', 'timestamp',
            'created', 'modified', 'last_seen'
        ];

        const user: Record<string, any> = {};
        const meta: Record<string, any> = {};

        Object.entries(node.properties).forEach(([key, value]) => {
            if (metadataKeys.includes(key.toLowerCase())) {
                meta[key] = value;
            } else {
                user[key] = value;
            }
        });

        return { userProperties: user, metadata: meta };
    }, [node?.properties]);

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
                <p className="empty-hint">💡 Click on a node in the graph</p>
                <p className="empty-hint">💡 Right-click to expand relationships</p>
            </div>
        );
    }

    const renderPropertyValue = (value: any): React.ReactNode => {
        // Handle null and undefined (T099)
        if (value === null) {
            return <span className="null-value">null</span>;
        }
        if (value === undefined) {
            return <span className="null-value">undefined</span>;
        }
        // Handle empty strings
        if (value === '') {
            return <span className="empty-value">(empty string)</span>;
        }
        // Handle objects and arrays
        if (typeof value === 'object') {
            return <pre className="json-value">{JSON.stringify(value, null, 2)}</pre>;
        }
        // Handle booleans
        if (typeof value === 'boolean') {
            return <span className={`boolean-value ${value}`}>{String(value)}</span>;
        }
        // Handle numbers
        if (typeof value === 'number') {
            return <span className="number-value">{value}</span>;
        }
        // Handle strings (default)
        return <span className="string-value">{String(value)}</span>;
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
                    <button className="close-button" onClick={onClose} aria-label="Close details">
                        ✕
                    </button>
                )}
            </div>

            {loading && <LoadingSkeleton type="detail" />}

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
                            <div className="info-item">
                                <span className="info-label">Category:</span>
                                <span className={`info-value category-${node.category || 'unknown'}`}>
                                    {node.category || 'unknown'}
                                </span>
                            </div>
                        </div>
                    </div>

                    {/* T096: Enhanced property list view with categorization */}
                    {Object.keys(userProperties).length > 0 && (
                        <div className="detail-section">
                            <div
                                className="section-header collapsible"
                                onClick={() => toggleSection('properties')}
                            >
                                <h3>Properties ({Object.keys(userProperties).length})</h3>
                                <span className="collapse-icon">
                                    {sectionsExpanded.properties ? '▼' : '▶'}
                                </span>
                            </div>
                            {sectionsExpanded.properties && (
                                <div className="properties-list">
                                    {Object.entries(userProperties)
                                        .sort(([a], [b]) => a.localeCompare(b))
                                        .map(([key, value]) => (
                                            <div key={key} className="property-item">
                                                <div className="property-header">
                                                    <div className="property-key">{key}</div>
                                                    <Tooltip content={`Copy ${key} value`} position="left">
                                                        <button
                                                            className="copy-button-small"
                                                            onClick={() => copyToClipboard(String(value), `Copied ${key}`)}
                                                            aria-label={`Copy ${key}`}
                                                        >
                                                            📋
                                                        </button>
                                                    </Tooltip>
                                                </div>
                                                <div className="property-value">
                                                    {renderPropertyValue(value)}
                                                </div>
                                            </div>
                                        ))}
                                </div>
                            )}
                        </div>
                    )}

                    {/* T097: Metadata section (timestamps, category, system properties) */}
                    {Object.keys(metadata).length > 0 && (
                        <div className="detail-section">
                            <div
                                className="section-header collapsible"
                                onClick={() => toggleSection('metadata')}
                            >
                                <h3>Metadata ({Object.keys(metadata).length})</h3>
                                <span className="collapse-icon">
                                    {sectionsExpanded.metadata ? '▼' : '▶'}
                                </span>
                            </div>
                            {sectionsExpanded.metadata && (
                                <div className="metadata-list">
                                    {Object.entries(metadata)
                                        .sort(([a], [b]) => a.localeCompare(b))
                                        .map(([key, value]) => (
                                            <div key={key} className="metadata-item">
                                                <span className="metadata-label">{key}:</span>
                                                <span className="metadata-value">
                                                    {renderPropertyValue(value)}
                                                </span>
                                            </div>
                                        ))}
                                </div>
                            )}
                        </div>
                    )}

                    {/* T098: Related entities list with T100: expand buttons */}
                    {relatedEntities.length > 0 && (
                        <div className="detail-section">
                            <div
                                className="section-header collapsible"
                                onClick={() => toggleSection('relationships')}
                            >
                                <h3>Related Entities ({relatedEntities.length})</h3>
                                <span className="collapse-icon">
                                    {sectionsExpanded.relationships ? '▼' : '▶'}
                                </span>
                            </div>
                            {sectionsExpanded.relationships && (
                                <div className="related-entities-list">
                                    {relatedEntities.map(({ edge, node: relatedNode }, idx) => {
                                        // Handle both string IDs and object references (react-force-graph modifies these)
                                        const sourceId = typeof edge.source === 'string' ? edge.source : (edge.source as any).id;
                                        const isOutgoing = sourceId === node?.id;

                                        return (
                                            <div
                                                key={idx}
                                                className="related-entity-item"
                                                onClick={() => onExpandRelationship && onExpandRelationship(relatedNode.id)}
                                                style={{ cursor: onExpandRelationship ? 'pointer' : 'default' }}
                                            >
                                                <div className="relationship-header">
                                                    <span className="relationship-type">{edge.type}</span>
                                                    <span className="relationship-direction">
                                                        {isOutgoing ? '→' : '←'}
                                                    </span>
                                                </div>
                                                <div className="related-node-info">
                                                    <div className="related-node-label">
                                                        {relatedNode.label || relatedNode.id}
                                                    </div>
                                                    <div className="related-node-type">
                                                        <span className={`type-badge ${relatedNode.type.toLowerCase()}`}>
                                                            {relatedNode.type}
                                                        </span>
                                                    </div>
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </div>
                    )}

                    {/* Relationship loading state */}
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
