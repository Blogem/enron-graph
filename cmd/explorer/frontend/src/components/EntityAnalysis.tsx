import { useState } from 'react';
import './EntityAnalysis.css';
import { wailsAPI } from '../services/wails';
import { main } from '../wailsjs/go/models';
import LoadingSkeleton from './LoadingSkeleton';

interface EntityAnalysisProps {
    onPromote: (typeName: string) => void;
}

type SortColumn = 'rank' | 'typeName' | 'frequency' | 'density' | 'consistency' | 'score';
type SortDirection = 'asc' | 'desc';

function EntityAnalysis({ onPromote }: EntityAnalysisProps) {
    // Configuration state
    const [minOccurrences, setMinOccurrences] = useState<number>(5);
    const [minConsistency, setMinConsistency] = useState<number>(0.4);
    const [topN, setTopN] = useState<number>(10);

    // Analysis state
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const [results, setResults] = useState<main.AnalysisResponse | null>(null);

    // Table state
    const [sortColumn, setSortColumn] = useState<SortColumn>('rank');
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
    const [selectedCandidate, setSelectedCandidate] = useState<main.TypeCandidate | null>(null);

    // Validation errors
    const [validationErrors, setValidationErrors] = useState<string[]>([]);

    const validateParameters = (): boolean => {
        const errors: string[] = [];

        if (minOccurrences < 1) {
            errors.push('Minimum occurrences must be at least 1');
        }

        if (minConsistency < 0 || minConsistency > 1) {
            errors.push('Minimum consistency must be between 0 and 1');
        }

        if (topN < 1) {
            errors.push('Top N must be at least 1');
        }

        setValidationErrors(errors);
        return errors.length === 0;
    };

    const handleAnalyze = async () => {
        if (!validateParameters()) {
            return;
        }

        try {
            setLoading(true);
            setError(null);
            setValidationErrors([]);

            const request = new main.AnalysisRequest({
                minOccurrences,
                minConsistency,
                topN
            });

            const response = await wailsAPI.analyzeEntities(request);
            setResults(response);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to analyze entities');
            console.error('Error analyzing entities:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleSort = (column: SortColumn) => {
        if (column === sortColumn) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortColumn(column);
            setSortDirection('asc');
        }
    };

    const getSortedCandidates = (): main.TypeCandidate[] => {
        if (!results || !results.candidates) return [];

        const sorted = [...results.candidates].sort((a, b) => {
            const aVal = a[sortColumn];
            const bVal = b[sortColumn];

            if (typeof aVal === 'string' && typeof bVal === 'string') {
                return sortDirection === 'asc'
                    ? aVal.localeCompare(bVal)
                    : bVal.localeCompare(aVal);
            }

            return sortDirection === 'asc'
                ? (aVal as number) - (bVal as number)
                : (bVal as number) - (aVal as number);
        });

        return sorted;
    };

    const handleRowClick = (candidate: main.TypeCandidate) => {
        setSelectedCandidate(candidate);
    };

    const handlePromoteClick = (candidate: main.TypeCandidate, e: React.MouseEvent) => {
        e.stopPropagation();
        onPromote(candidate.typeName);
    };

    const sortedCandidates = getSortedCandidates();

    return (
        <div className="entity-analysis">
            <div className="analysis-header">
                <h2>Entity Analysis</h2>
                <p className="analysis-description">
                    Analyze discovered entity types and identify candidates for promotion to formal schemas.
                </p>
            </div>

            {/* Configuration Panel */}
            <div className="configuration-panel">
                <h3>Configuration</h3>
                <div className="config-controls">
                    <div className="config-control">
                        <label htmlFor="minOccurrences">
                            Min Occurrences
                            <span className="help-text">Minimum number of entity instances required</span>
                        </label>
                        <input
                            id="minOccurrences"
                            type="number"
                            min="1"
                            value={minOccurrences}
                            onChange={(e) => setMinOccurrences(parseInt(e.target.value) || 1)}
                            disabled={loading}
                        />
                    </div>

                    <div className="config-control">
                        <label htmlFor="minConsistency">
                            Min Consistency
                            <span className="help-text">Minimum consistency score (0-1)</span>
                        </label>
                        <input
                            id="minConsistency"
                            type="number"
                            min="0"
                            max="1"
                            step="0.1"
                            value={minConsistency}
                            onChange={(e) => setMinConsistency(parseFloat(e.target.value) || 0)}
                            disabled={loading}
                        />
                    </div>

                    <div className="config-control">
                        <label htmlFor="topN">
                            Top N
                            <span className="help-text">Number of top candidates to return</span>
                        </label>
                        <input
                            id="topN"
                            type="number"
                            min="1"
                            value={topN}
                            onChange={(e) => setTopN(parseInt(e.target.value) || 1)}
                            disabled={loading}
                        />
                    </div>

                    <button
                        className="analyze-button"
                        onClick={handleAnalyze}
                        disabled={loading}
                    >
                        {loading ? 'Analyzing...' : 'Analyze'}
                    </button>
                </div>

                {/* Validation Errors */}
                {validationErrors.length > 0 && (
                    <div className="validation-errors">
                        {validationErrors.map((err, idx) => (
                            <div key={idx} className="error-message">⚠️ {err}</div>
                        ))}
                    </div>
                )}
            </div>

            {/* Error Display */}
            {error && (
                <div className="analysis-error">
                    <p>⚠️ {error}</p>
                    <button onClick={handleAnalyze}>Retry</button>
                </div>
            )}

            {/* Loading State */}
            {loading && <LoadingSkeleton type="text" />}

            {/* Results Table */}
            {!loading && results && results.candidates && results.candidates.length > 0 && (
                <div className="results-section">
                    <div className="results-header">
                        <h3>Analysis Results</h3>
                        <p className="results-summary">
                            Found {results.candidates.length} candidates out of {results.totalTypes} total types
                        </p>
                    </div>

                    <div className="results-table-container">
                        <table className="results-table">
                            <thead>
                                <tr>
                                    <th onClick={() => handleSort('rank')} className="sortable">
                                        Rank {sortColumn === 'rank' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('typeName')} className="sortable">
                                        Type Name {sortColumn === 'typeName' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('frequency')} className="sortable">
                                        Frequency {sortColumn === 'frequency' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('density')} className="sortable">
                                        Density {sortColumn === 'density' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('consistency')} className="sortable">
                                        Consistency {sortColumn === 'consistency' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('score')} className="sortable">
                                        Score {sortColumn === 'score' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th>Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {sortedCandidates.map((candidate) => (
                                    <tr
                                        key={candidate.typeName}
                                        onClick={() => handleRowClick(candidate)}
                                        className={selectedCandidate?.typeName === candidate.typeName ? 'selected' : ''}
                                    >
                                        <td>{candidate.rank}</td>
                                        <td className="type-name">{candidate.typeName}</td>
                                        <td>{candidate.frequency}</td>
                                        <td>{candidate.density.toFixed(2)}</td>
                                        <td>{candidate.consistency.toFixed(2)}</td>
                                        <td>{candidate.score.toFixed(2)}</td>
                                        <td>
                                            <button
                                                className="promote-button"
                                                onClick={(e) => handlePromoteClick(candidate, e)}
                                                title="Promote this entity type to a formal schema"
                                            >
                                                Promote
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>

                    {/* Selected Candidate Details */}
                    {selectedCandidate && (
                        <div className="candidate-details">
                            <h4>Selected: {selectedCandidate.typeName}</h4>
                            <div className="detail-grid">
                                <div className="detail-item">
                                    <span className="detail-label">Rank:</span>
                                    <span className="detail-value">{selectedCandidate.rank}</span>
                                </div>
                                <div className="detail-item">
                                    <span className="detail-label">Frequency:</span>
                                    <span className="detail-value">{selectedCandidate.frequency}</span>
                                </div>
                                <div className="detail-item">
                                    <span className="detail-label">Density:</span>
                                    <span className="detail-value">{selectedCandidate.density.toFixed(3)}</span>
                                </div>
                                <div className="detail-item">
                                    <span className="detail-label">Consistency:</span>
                                    <span className="detail-value">{selectedCandidate.consistency.toFixed(3)}</span>
                                </div>
                                <div className="detail-item">
                                    <span className="detail-label">Score:</span>
                                    <span className="detail-value">{selectedCandidate.score.toFixed(3)}</span>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            )}

            {/* Empty Results */}
            {!loading && results && results.candidates && results.candidates.length === 0 && (
                <div className="empty-results">
                    <p>No candidates found matching the specified criteria.</p>
                    <p className="help-text">Try adjusting the configuration parameters and analyze again.</p>
                </div>
            )}
        </div>
    );
}

export default EntityAnalysis;
