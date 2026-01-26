import { useState, useEffect } from 'react';
import './App.css';
import SchemaPanel from './components/SchemaPanel';
import { wailsAPI } from './services/wails';
import type { explorer } from './wailsjs/go/models';

function App() {
    const [schema, setSchema] = useState<explorer.SchemaResponse | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedType, setSelectedType] = useState<explorer.SchemaType | null>(null);
    const [detailsLoading, setDetailsLoading] = useState<boolean>(false);

    // Load schema on mount
    useEffect(() => {
        loadSchema();
    }, []);

    const loadSchema = async () => {
        try {
            setLoading(true);
            setError(null);
            const data = await wailsAPI.getSchema();
            setSchema(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load schema');
            console.error('Error loading schema:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = async () => {
        try {
            setLoading(true);
            setError(null);
            await wailsAPI.refreshSchema();
            const data = await wailsAPI.getSchema();
            setSchema(data);
            setSelectedType(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to refresh schema');
            console.error('Error refreshing schema:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleTypeClick = async (typeName: string) => {
        try {
            setDetailsLoading(true);
            const details = await wailsAPI.getTypeDetails(typeName);
            setSelectedType(details);
        } catch (err) {
            console.error('Error loading type details:', err);
            setError(err instanceof Error ? err.message : 'Failed to load type details');
        } finally {
            setDetailsLoading(false);
        }
    };

    return (
        <div id="App">
            <div className="app-header">
                <h1>Graph Explorer</h1>
            </div>
            <div className="app-container">
                <div className="sidebar">
                    <SchemaPanel
                        schema={schema}
                        loading={loading}
                        error={error}
                        onRefresh={handleRefresh}
                        onTypeClick={handleTypeClick}
                    />
                </div>
                <div className="main-content">
                    {detailsLoading && (
                        <div className="details-loading">Loading type details...</div>
                    )}
                    {!detailsLoading && selectedType && (
                        <div className="type-details">
                            <h2>{selectedType.name}</h2>
                            <div className="details-stats">
                                <div className="stat">
                                    <span className="stat-label">Count:</span>
                                    <span className="stat-value">{selectedType.count}</span>
                                </div>
                                <div className="stat">
                                    <span className="stat-label">Status:</span>
                                    <span className={`stat-value ${selectedType.is_promoted ? 'promoted' : 'discovered'}`}>
                                        {selectedType.is_promoted ? 'Promoted' : 'Discovered'}
                                    </span>
                                </div>
                            </div>
                            {selectedType.properties && selectedType.properties.length > 0 && (
                                <div className="properties-section">
                                    <h3>Properties</h3>
                                    <div className="properties-grid">
                                        {selectedType.properties.map((prop, idx) => (
                                            <div key={idx} className="property-item">
                                                <div className="property-name">{prop.name}</div>
                                                <div className="property-type">{prop.data_type}</div>
                                                {prop.sample_value && prop.sample_value.length > 0 && (
                                                    <div className="property-samples">
                                                        {prop.sample_value.slice(0, 3).map((val: string, i: number) => (
                                                            <span key={i} className="sample-value">{val}</span>
                                                        ))}
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                    {!detailsLoading && !selectedType && !error && (
                        <div className="placeholder">
                            <p>Select a type from the schema panel to view details</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

export default App;
