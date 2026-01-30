import { useState, useEffect } from 'react';
import './EntityPromotion.css';
import { wailsAPI } from '../services/wails';
import { main } from '../wailsjs/go/models';

interface EntityPromotionProps {
    typeName: string | null;
    onCancel: () => void;
    onSuccess: () => void;
    onViewInGraph?: (typeName: string) => void;
}

function EntityPromotion({ typeName, onCancel, onSuccess, onViewInGraph }: EntityPromotionProps) {
    const [loading, setLoading] = useState<boolean>(false);
    const [result, setResult] = useState<main.PromotionResponse | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [rebuilding, setRebuilding] = useState<boolean>(false);

    // Reset state when typeName changes
    useEffect(() => {
        setResult(null);
        setError(null);
    }, [typeName]);

    if (!typeName) {
        return null;
    }

    const handleConfirm = async () => {
        try {
            setLoading(true);
            setError(null);

            const request = new main.PromotionRequest({
                typeName
            });

            const response = await wailsAPI.promoteEntity(request);

            // Check if the response indicates failure
            if (!response.success && response.error) {
                setError(response.error);
                return;
            }

            setResult(response);
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err);
            setError(errorMessage || 'Failed to promote entity');
            console.error('Error promoting entity:', err);
            console.error('Error details:', JSON.stringify(err, null, 2));
        } finally {
            setLoading(false);
        }
    };

    const handleCancel = () => {
        if (!loading && !rebuilding) {
            onCancel();
        }
    };

    const handleRebuild = async () => {
        try {
            setRebuilding(true);
            await wailsAPI.regenerateAndReload();
            // The app will rebuild automatically, show message
        } catch (err) {
            console.error('Failed to trigger rebuild:', err);
            alert('Failed to trigger rebuild. Please restart manually.');
        } finally {
            setRebuilding(false);
        }
    };

    return (
        <div className="entity-promotion">
            <div className="promotion-header">
                <div>
                    <h2>Promote Entity Type</h2>
                    <p className="promotion-description">
                        Convert discovered entity type to a formal Ent schema with database migration.
                    </p>
                </div>
                <button
                    className="close-button"
                    onClick={handleCancel}
                    disabled={loading}
                    aria-label="Close"
                >
                    ✕
                </button>
            </div>

            {/* Promotion Preview (before confirmation) */}
            {!result && (
                <div className="promotion-preview">
                    <div className="preview-section">
                        <h3>Type to Promote</h3>
                        <div className="type-name-display">{typeName}</div>
                    </div>

                    <div className="preview-section">
                        <p className="preview-description">
                            This will:
                        </p>
                        <ul className="preview-steps">
                            <li>Analyze entity properties and generate schema</li>
                            <li>Create an Ent schema file in <code>ent/schema/</code></li>
                            <li>Run database migration to create the table</li>
                            <li>Migrate existing discovered entities to the new schema</li>
                        </ul>
                    </div>

                    {error && (
                        <div className="promotion-error">
                            <p><strong>⚠️ Error:</strong></p>
                            <p>{error}</p>
                        </div>
                    )}

                    <div className="promotion-actions">
                        <button
                            className="cancel-button"
                            onClick={handleCancel}
                            disabled={loading}
                        >
                            Cancel
                        </button>
                        <button
                            className="confirm-button"
                            onClick={handleConfirm}
                            disabled={loading}
                        >
                            {loading ? 'Promoting...' : 'Confirm Promote'}
                        </button>
                    </div>
                </div>
            )}

            {/* Promotion Results (after execution) */}
            {result && (
                <div className="promotion-results">
                    {result.success ? (
                        <div className="success-results">
                            <div className="success-icon">✓</div>
                            <h3>Promotion Successful!</h3>

                            <div className="result-details">
                                <div className="result-item">
                                    <span className="result-label">Schema File:</span>
                                    <span className="result-value">
                                        <code>{result.schemaFilePath}</code>
                                    </span>
                                </div>

                                <div className="result-item">
                                    <span className="result-label">Entities Migrated:</span>
                                    <span className="result-value">{result.entitiesMigrated}</span>
                                </div>

                                {result.validationErrors > 0 && (
                                    <div className="result-item warning">
                                        <span className="result-label">Validation Errors:</span>
                                        <span className="result-value">{result.validationErrors}</span>
                                        <span className="result-note">
                                            Some entities could not be migrated due to validation errors
                                        </span>
                                    </div>
                                )}

                                {result.properties && result.properties.length > 0 && (
                                    <div className="properties-section">
                                        <h4>Generated Properties ({result.properties.length})</h4>
                                        <div className="properties-list">
                                            {result.properties.map((prop, idx) => (
                                                <div key={idx} className="property-item">
                                                    <span className="property-name">{prop.name}</span>
                                                    <span className="property-type">{prop.type}</span>
                                                    {prop.required && (
                                                        <span className="property-required">Required</span>
                                                    )}
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}
                            </div>

                            <div className="restart-notice">
                                <p>⚠️ <strong>Action Required:</strong> Rebuild the application to use the new type.</p>
                                <p className="restart-instructions">
                                    Click the "Rebuild Now" button below to regenerate code and trigger a rebuild.
                                    The app will automatically reload when ready.
                                </p>
                            </div>

                            <div className="success-actions">
                                <button
                                    className="rebuild-button"
                                    onClick={handleRebuild}
                                    disabled={rebuilding}
                                >
                                    {rebuilding ? 'Rebuilding...' : 'Rebuild Now'}
                                </button>
                                {onViewInGraph && (
                                    <button
                                        className="view-graph-button"
                                        onClick={() => onViewInGraph(typeName!)}
                                        disabled={rebuilding}
                                    >
                                        View in Graph (after rebuild)
                                    </button>
                                )}
                                <button className="done-button" onClick={handleCancel} disabled={rebuilding}>
                                    Done
                                </button>
                            </div>
                        </div>
                    ) : (
                        <div className="failure-results">
                            <div className="failure-icon">✗</div>
                            <h3>Promotion Failed</h3>

                            {result.error && (
                                <div className="failure-error">
                                    <p>{result.error}</p>
                                </div>
                            )}

                            <div className="failure-actions">
                                <button className="retry-button" onClick={handleConfirm} disabled={loading}>
                                    Retry
                                </button>
                                <button className="cancel-button" onClick={handleCancel}>
                                    Cancel
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            )}

            {/* Loading Overlay */}
            {loading && (
                <div className="loading-overlay">
                    <div className="spinner"></div>
                    <p>Promoting entity type...</p>
                    <p className="loading-note">This may take a few moments</p>
                </div>
            )}
        </div>
    );
}

export default EntityPromotion;
