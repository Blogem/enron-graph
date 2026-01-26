import React from 'react';
import type { explorer } from '../wailsjs/go/models';
import LoadingSkeleton from './LoadingSkeleton';
import Tooltip from './Tooltip';
import './SchemaPanel.css';

interface SchemaPanelProps {
    schema: explorer.SchemaResponse | null;
    loading: boolean;
    error: string | null;
    selectedTypeName: string | null;
    onRefresh: () => void;
    onTypeClick: (typeName: string) => void;
}

const SchemaPanel: React.FC<SchemaPanelProps> = ({
    schema,
    loading,
    error,
    selectedTypeName,
    onRefresh,
    onTypeClick
}) => {
    const handleTypeClick = (typeName: string) => {
        onTypeClick(typeName);
    };

    if (loading) {
        return <LoadingSkeleton type="schema" />;
    }

    if (error) {
        return (
            <div className="schema-panel">
                <div className="schema-header">
                    <h2>Schema</h2>
                    <button onClick={onRefresh} className="refresh-btn">Refresh</button>
                </div>
                <div className="error">Error: {error}</div>
            </div>
        );
    }

    if (!schema) {
        return (
            <div className="schema-panel">
                <div className="schema-header">
                    <h2>Schema</h2>
                </div>
                <div className="empty">No schema data</div>
            </div>
        );
    }

    return (
        <div className="schema-panel">
            <div className="schema-header">
                <h2>Schema</h2>
                <div className="schema-stats">
                    <span>{schema.total_entities} entities</span>
                </div>
                <Tooltip content="Refresh schema to see latest types and counts" position="bottom">
                    <button onClick={onRefresh} className="refresh-btn" aria-label="Refresh schema">
                        ⟳
                    </button>
                </Tooltip>
            </div>

            {/* Promoted Types Section */}
            <div className="schema-section">
                <h3 className="section-title">
                    <span className="promoted-icon">⭐</span>
                    Promoted Types ({schema.promoted_types?.length || 0})
                </h3>
                <div className="type-list">
                    {schema.promoted_types?.map((type) => (
                        <div
                            key={type.name}
                            className={`type-item promoted ${selectedTypeName === type.name ? 'selected' : ''}`}
                            onClick={() => handleTypeClick(type.name)}
                        >
                            <div className="type-header">
                                <span className="type-name">{type.name}</span>
                                <span className="type-count">{type.count}</span>
                            </div>
                            {type.properties && type.properties.length > 0 && (
                                <div className="type-properties">
                                    {type.properties.slice(0, 3).map((prop, idx) => (
                                        <span key={idx} className="property-badge">
                                            {prop.name}
                                        </span>
                                    ))}
                                    {type.properties.length > 3 && (
                                        <span className="property-badge more">
                                            +{type.properties.length - 3}
                                        </span>
                                    )}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            </div>

            {/* Discovered Types Section */}
            <div className="schema-section">
                <h3 className="section-title">
                    <span className="discovered-icon">🔍</span>
                    Discovered Types ({schema.discovered_types?.length || 0})
                </h3>
                <div className="type-list">
                    {schema.discovered_types?.map((type) => (
                        <div
                            key={type.name}
                            className={`type-item discovered ${selectedTypeName === type.name ? 'selected' : ''}`}
                            onClick={() => handleTypeClick(type.name)}
                        >
                            <div className="type-header">
                                <span className="type-name">{type.name}</span>
                                <span className="type-count">{type.count}</span>
                            </div>
                            {type.properties && type.properties.length > 0 && (
                                <div className="type-properties">
                                    {type.properties.slice(0, 3).map((prop, idx) => (
                                        <span key={idx} className="property-badge">
                                            {prop.name}
                                        </span>
                                    ))}
                                    {type.properties.length > 3 && (
                                        <span className="property-badge more">
                                            +{type.properties.length - 3}
                                        </span>
                                    )}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default SchemaPanel;
