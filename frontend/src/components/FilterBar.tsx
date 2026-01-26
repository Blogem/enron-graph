import React, { useState, useEffect } from 'react';
import type { NodeFilter } from '../types/graph';
import type { explorer } from '../wailsjs/go/models';
import './FilterBar.css';

interface FilterBarProps {
    schema: explorer.SchemaResponse | null;
    onFilterChange: (filter: NodeFilter) => void;
    initialFilter?: NodeFilter;
}

const FilterBar: React.FC<FilterBarProps> = ({ schema, onFilterChange, initialFilter }) => {
    const [selectedTypes, setSelectedTypes] = useState<string[]>(initialFilter?.types || []);
    const [category, setCategory] = useState<string>(initialFilter?.category || 'all');
    const [searchQuery, setSearchQuery] = useState<string>(initialFilter?.search_query || '');
    const [limit, setLimit] = useState<number>(initialFilter?.limit || 100);
    const [isExpanded, setIsExpanded] = useState<boolean>(false);

    // Debounce search query
    const [debouncedSearchQuery, setDebouncedSearchQuery] = useState<string>(searchQuery);

    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedSearchQuery(searchQuery);
        }, 300);

        return () => clearTimeout(timer);
    }, [searchQuery]);

    // Update filter when any value changes
    useEffect(() => {
        const filter: NodeFilter = {
            types: selectedTypes.length > 0 ? selectedTypes : undefined,
            category: category !== 'all' ? category : undefined,
            search_query: debouncedSearchQuery || undefined,
            limit
        };
        onFilterChange(filter);
    }, [selectedTypes, category, debouncedSearchQuery, limit, onFilterChange]);

    // Get all available types from schema
    const allTypes = React.useMemo(() => {
        if (!schema) return [];
        const types = [
            ...(schema.promoted_types?.map(t => t.name) || []),
            ...(schema.discovered_types?.map(t => t.name) || [])
        ];
        return Array.from(new Set(types)).sort();
    }, [schema]);

    const handleTypeToggle = (typeName: string) => {
        setSelectedTypes(prev => {
            if (prev.includes(typeName)) {
                return prev.filter(t => t !== typeName);
            } else {
                return [...prev, typeName];
            }
        });
    };

    const handleCategoryChange = (newCategory: string) => {
        setCategory(newCategory);
    };

    const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchQuery(e.target.value);
    };

    const handleClearFilters = () => {
        setSelectedTypes([]);
        setCategory('all');
        setSearchQuery('');
        setLimit(100);
    };

    const hasActiveFilters = selectedTypes.length > 0 || category !== 'all' || searchQuery !== '';

    return (
        <div className="filter-bar">
            <div className="filter-bar-header">
                <button
                    className="filter-toggle"
                    onClick={() => setIsExpanded(!isExpanded)}
                    title={isExpanded ? 'Collapse filters' : 'Expand filters'}
                >
                    {isExpanded ? '▼' : '▶'} Filters
                    {hasActiveFilters && <span className="active-indicator">{selectedTypes.length > 0 ? ` (${selectedTypes.length} types)` : ''}</span>}
                </button>
                <div className="filter-bar-actions">
                    <input
                        type="text"
                        className="search-input-compact"
                        placeholder="Quick search..."
                        value={searchQuery}
                        onChange={handleSearchChange}
                    />
                    {hasActiveFilters && (
                        <button
                            className="clear-btn-compact"
                            onClick={handleClearFilters}
                            title="Clear all filters"
                        >
                            Clear
                        </button>
                    )}
                </div>
            </div>
            {isExpanded && (
                <div className="filter-bar-content">
                    <div className="filter-section">
                        <label className="filter-label">Category</label>
                        <div className="category-buttons">
                            <button
                                className={`category-btn ${category === 'all' ? 'active' : ''}`}
                                onClick={() => handleCategoryChange('all')}
                                title="Show all entities"
                            >
                                All
                            </button>
                            <button
                                className={`category-btn promoted ${category === 'promoted' ? 'active' : ''}`}
                                onClick={() => handleCategoryChange('promoted')}
                                title="Show promoted types only"
                            >
                                ⭐ Promoted
                            </button>
                            <button
                                className={`category-btn discovered ${category === 'discovered' ? 'active' : ''}`}
                                onClick={() => handleCategoryChange('discovered')}
                                title="Show discovered types only"
                            >
                                🔍 Discovered
                            </button>
                        </div>
                    </div>

                    <div className="filter-section">
                        <label className="filter-label">Type Filter</label>
                        <div className="type-selector">
                            {allTypes.length === 0 ? (
                                <p className="empty-message">No types available</p>
                            ) : (
                                <div className="type-checkboxes">
                                    {allTypes.map(typeName => (
                                        <label key={typeName} className="type-checkbox">
                                            <input
                                                type="checkbox"
                                                checked={selectedTypes.includes(typeName)}
                                                onChange={() => handleTypeToggle(typeName)}
                                            />
                                            <span className="type-name">{typeName}</span>
                                        </label>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>

                    <div className="filter-section">
                        <label className="filter-label" htmlFor="search-input">Search Properties</label>
                        <input
                            id="search-input"
                            type="text"
                            className="search-input"
                            placeholder="Search by property value..."
                            value={searchQuery}
                            onChange={handleSearchChange}
                        />
                    </div>

                    <div className="filter-section">
                        <label className="filter-label" htmlFor="limit-input">Result Limit</label>
                        <input
                            id="limit-input"
                            type="number"
                            className="limit-input"
                            min="10"
                            max="1000"
                            step="10"
                            value={limit}
                            onChange={(e) => setLimit(parseInt(e.target.value, 10) || 100)}
                        />
                    </div>

                    <div className="filter-actions">
                        {hasActiveFilters && (
                            <div className="active-filters-badge">
                                {selectedTypes.length > 0 && <span>{selectedTypes.length} types</span>}
                                {category !== 'all' && <span>{category}</span>}
                                {searchQuery && <span>search: "{searchQuery}"</span>}
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
};

export default FilterBar;
