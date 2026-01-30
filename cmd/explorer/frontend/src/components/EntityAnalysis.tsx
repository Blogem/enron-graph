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
                        <div className="sort-controls">
                            <label htmlFor="sort-by">Sort by:</label>
                            <select
                                id="sort-by"
                                value={sortColumn}
                                onChange={(e) => handleSort(e.target.value as SortColumn)}
                            >
                                <option value="rank">Rank</option>
                                <option value="typeName">Type Name</option>
                                <option value="frequency">Frequency</option>
                                <option value="density">Density</option>
                                <option value="consistency">Consistency</option>
                                <option value="score">Score</option>
                            </select>
                            <button
                                className="sort-direction-button"
                                onClick={() => setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')}
                                title={sortDirection === 'asc' ? 'Ascending' : 'Descending'}
                            >
                                {sortDirection === 'asc' ? '↑' : '↓'}
                            </button>
                        </div>
                    </div>

                    <div className="results-list">
                        {sortedCandidates.map((candidate) => (
                            <div
                                key={candidate.typeName}
                                className={`candidate-card ${selectedCandidate?.typeName === candidate.typeName ? 'selected' : ''}`}
                                onClick={() => handleRowClick(candidate)}
                            >
                                <div className="candidate-header">
                                    <span className="candidate-rank">#{candidate.rank}</span>
                                    <span className="candidate-name">{candidate.typeName}</span>
                                </div>
                                <div className="candidate-metrics">
                                    <div className="metric">
                                        <span className="metric-label">Freq</span>
                                        <span className="metric-value">{candidate.frequency}</span>
                                    </div>
                                    <div className="metric">
                                        <span className="metric-label">Density</span>
                                        <span className="metric-value">{candidate.density.toFixed(2)}</span>
                                    </div>
                                    <div className="metric">
                                        <span className="metric-label">Consistency</span>
                                        <span className="metric-value">{candidate.consistency.toFixed(2)}</span>
                                    </div>
                                    <div className="metric">
                                        <span className="metric-label">Score</span>
                                        <span className="metric-value score">{candidate.score.toFixed(2)}</span>
                                    </div>
                                </div>
                                <button
                                    className="promote-button"
                                    onClick={(e) => handlePromoteClick(candidate, e)}
                                    title="Promote this entity type to a formal schema"
                                >
                                    Promote
                                </button>
                            </div>
                        ))}
                    </div>
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
