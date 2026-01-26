import { useState, useEffect, useCallback, useMemo } from 'react';
import './App.css';
import './components/TypeDetailsPanel.css';
import SchemaPanel from './components/SchemaPanel';
import GraphCanvas from './components/GraphCanvas';
import DetailPanel from './components/DetailPanel';
import FilterBar from './components/FilterBar';
import { wailsAPI } from './services/wails';
import type { explorer } from './wailsjs/go/models';
import type { GraphData, GraphNodeWithPosition, ExpandedNodeState, NodeFilter, GraphEdge } from './types/graph';

function App() {
    // Schema state
    const [schema, setSchema] = useState<explorer.SchemaResponse | null>(null);
    const [schemaLoading, setSchemaLoading] = useState<boolean>(true);
    const [schemaError, setSchemaError] = useState<string | null>(null);
    const [selectedType, setSelectedType] = useState<explorer.SchemaType | null>(null);
    const [selectedTypeName, setSelectedTypeName] = useState<string | null>(null);
    const [detailsLoading, setDetailsLoading] = useState<boolean>(false);

    // Graph state
    const [graphData, setGraphData] = useState<GraphData>({ nodes: [], links: [] });
    const [graphLoading, setGraphLoading] = useState<boolean>(true);
    const [graphError, setGraphError] = useState<string | null>(null);
    const [selectedNode, setSelectedNode] = useState<GraphNodeWithPosition | null>(null);
    const [expandedNodes, setExpandedNodes] = useState<Map<string, ExpandedNodeState>>(new Map());
    const [loadingNodeId, setLoadingNodeId] = useState<string | null>(null);

    // Filter state
    const [activeFilter, setActiveFilter] = useState<NodeFilter>({});
    const [highlightedNodeIds, setHighlightedNodeIds] = useState<Set<string>>(new Set());

    // Load schema on mount
    useEffect(() => {
        loadSchema();
    }, []);

    // Load random nodes on mount (auto-load 100 nodes) only if no filter is active
    useEffect(() => {
        const hasActiveFilter = activeFilter.types?.length || activeFilter.category || activeFilter.search_query;
        if (!hasActiveFilter) {
            loadRandomNodes(100);
        }
    }, []);

    // Apply filters when activeFilter changes
    useEffect(() => {
        const hasActiveFilter = activeFilter.types?.length || activeFilter.category || activeFilter.search_query;
        if (hasActiveFilter) {
            applyFilters(activeFilter);
        } else {
            // When filters are cleared, load random nodes again
            loadRandomNodes(100);
        }
    }, [activeFilter]);

    const loadSchema = async () => {
        try {
            setSchemaLoading(true);
            setSchemaError(null);
            const data = await wailsAPI.getSchema();
            setSchema(data);
        } catch (err) {
            setSchemaError(err instanceof Error ? err.message : 'Failed to load schema');
            console.error('Error loading schema:', err);
        } finally {
            setSchemaLoading(false);
        }
    };

    const loadRandomNodes = async (limit: number) => {
        try {
            setGraphLoading(true);
            setGraphError(null);
            console.log('Loading random nodes...');
            const response = await wailsAPI.getRandomNodes(limit);
            console.log('Got response:', response);

            // Transform response to GraphData format
            const nodes: GraphNodeWithPosition[] = response.nodes.map((node: any) => ({
                ...node,
                x: undefined,
                y: undefined,
                vx: undefined,
                vy: undefined,
                fx: undefined,
                fy: undefined
            }));

            console.log('Transformed nodes:', nodes.length, 'edges:', response.edges.length);
            setGraphData({
                nodes,
                links: response.edges
            });

            // Clear highlights when loading random nodes
            setHighlightedNodeIds(new Set());
        } catch (err) {
            setGraphError(err instanceof Error ? err.message : 'Failed to load graph');
            console.error('Error loading graph:', err);
        } finally {
            setGraphLoading(false);
        }
    };

    const applyFilters = async (filter: NodeFilter) => {
        try {
            setGraphLoading(true);
            setGraphError(null);
            console.log('Applying filters:', filter);
            const response = await wailsAPI.getNodes(filter);
            console.log('Filtered response:', response);

            // Transform response to GraphData format
            const nodes: GraphNodeWithPosition[] = response.nodes.map((node: any) => ({
                ...node,
                x: undefined,
                y: undefined,
                vx: undefined,
                vy: undefined,
                fx: undefined,
                fy: undefined
            }));

            console.log('Filtered nodes:', nodes.length, 'edges:', response.edges.length);
            setGraphData({
                nodes,
                links: response.edges
            });

            // Highlight nodes matching search query
            if (filter.search_query) {
                const highlighted = new Set<string>();
                nodes.forEach(node => {
                    // Check if search query appears in any property value
                    const searchLower = filter.search_query!.toLowerCase();
                    const matchesSearch = Object.values(node.properties || {}).some(val =>
                        String(val).toLowerCase().includes(searchLower)
                    );
                    if (matchesSearch) {
                        highlighted.add(node.id);
                    }
                });
                setHighlightedNodeIds(highlighted);
            } else {
                setHighlightedNodeIds(new Set());
            }
        } catch (err) {
            setGraphError(err instanceof Error ? err.message : 'Failed to apply filters');
            console.error('Error applying filters:', err);
        } finally {
            setGraphLoading(false);
        }
    };

    const handleFilterChange = useCallback((filter: NodeFilter) => {
        setActiveFilter(filter);
        // Reset expanded nodes and selection when filter changes
        setExpandedNodes(new Map());
        setSelectedNode(null);
    }, []);

    const handleRefreshSchema = async () => {
        try {
            setSchemaLoading(true);
            setSchemaError(null);
            await wailsAPI.refreshSchema();
            const data = await wailsAPI.getSchema();
            setSchema(data);
            setSelectedType(null);
            setSelectedTypeName(null);
        } catch (err) {
            setSchemaError(err instanceof Error ? err.message : 'Failed to refresh schema');
            console.error('Error refreshing schema:', err);
        } finally {
            setSchemaLoading(false);
        }
    };

    const handleTypeClick = async (typeName: string) => {
        try {
            setDetailsLoading(true);
            setSelectedTypeName(typeName);
            setSelectedNode(null); // Clear node selection when selecting a type
            const details = await wailsAPI.getTypeDetails(typeName);
            setSelectedType(details);
        } catch (err) {
            console.error('Error loading type details:', err);
            setSchemaError(err instanceof Error ? err.message : 'Failed to load type details');
        } finally {
            setDetailsLoading(false);
        }
    };

    const handleNodeClick = useCallback(async (node: GraphNodeWithPosition) => {
        try {
            setSelectedNode(node);
            setSelectedType(null); // Clear type selection when selecting a node
            setSelectedTypeName(null);
            // Optionally load full node details
            const details = await wailsAPI.getNodeDetails(node.id);
            setSelectedNode({
                ...node,
                properties: details.properties,
                category: details.category
            });
        } catch (err) {
            console.error('Error loading node details:', err);
        }
    }, []);

    // T098: Compute related entities from graph edges for the selected node
    const relatedEntities = useMemo(() => {
        if (!selectedNode) return [];

        const nodeId = selectedNode.id;
        const related: Array<{ edge: GraphEdge; node: GraphNodeWithPosition }> = [];

        // Find all edges connected to this node
        graphData.links.forEach(edge => {
            // react-force-graph modifies source/target to be object references after first render
            // We need to handle both string IDs and object references
            const sourceId = typeof edge.source === 'string' ? edge.source : (edge.source as any).id;
            const targetId = typeof edge.target === 'string' ? edge.target : (edge.target as any).id;

            if (sourceId === nodeId || targetId === nodeId) {
                // Determine the connected node ID
                const connectedNodeId = sourceId === nodeId ? targetId : sourceId;

                // Find the connected node in our graph data
                const connectedNode = graphData.nodes.find(n => n.id === connectedNodeId);

                if (connectedNode) {
                    related.push({
                        edge: edge,
                        node: connectedNode
                    });
                }
            }
        });

        return related;
    }, [selectedNode, graphData]);

    // T100: Handle expand relationship button click
    const handleExpandRelationship = useCallback((nodeId: string) => {
        const node = graphData.nodes.find(n => n.id === nodeId);
        if (node) {
            handleNodeClick(node);
        }
    }, [graphData.nodes, handleNodeClick]);

    const handleNodeRightClick = useCallback(async (node: GraphNodeWithPosition) => {
        // Don't auto-expand if already expanded (FR-006b: explicit click required)
        if (expandedNodes.has(node.id)) {
            console.log('Node already expanded, use "Load More" button for additional relationships');
            return;
        }

        // Load first batch of relationships
        await loadRelationshipBatch(node.id, 0);
    }, [expandedNodes]);

    const loadRelationshipBatch = async (nodeId: string, offset: number) => {
        try {
            setLoadingNodeId(nodeId);
            const batchSize = 50;
            const response = await wailsAPI.getRelationships(nodeId, offset, batchSize);

            // Add new nodes and edges to graph
            const newNodes: GraphNodeWithPosition[] = [];
            const nodeIds = new Set(graphData.nodes.map(n => n.id));

            // Extract unique target nodes from relationships
            const edges = response.edges || [];
            edges.forEach((rel: any) => {
                const targetId = rel.target;
                if (!nodeIds.has(targetId)) {
                    // Create placeholder node (will be loaded on demand)
                    newNodes.push({
                        id: targetId,
                        label: targetId,
                        type: 'Unknown',
                        properties: {}
                    });
                    nodeIds.add(targetId);
                }
            });

            // Update graph data
            setGraphData(prev => ({
                nodes: [...prev.nodes, ...newNodes],
                links: [...prev.links, ...edges]
            }));

            // Update expanded nodes state
            const newOffset = offset + edges.length;
            const hasMore = response.total_count > newOffset;

            setExpandedNodes(prev => {
                const updated = new Map(prev);
                updated.set(nodeId, {
                    nodeId,
                    offset: newOffset,
                    hasMore,
                    totalRelationships: response.total_count
                });
                return updated;
            });
        } catch (err) {
            console.error('Error loading relationships:', err);
            setGraphError(err instanceof Error ? err.message : 'Failed to load relationships');
        } finally {
            setLoadingNodeId(null);
        }
    };

    const handleLoadMore = useCallback(async (nodeId: string) => {
        const state = expandedNodes.get(nodeId);
        if (state && state.hasMore) {
            await loadRelationshipBatch(nodeId, state.offset);
        }
    }, [expandedNodes, graphData]);

    const handleCloseDetail = useCallback(() => {
        setSelectedNode(null);
    }, []);

    // Keyboard shortcuts
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                setSelectedNode(null);
                setSelectedType(null);
                setSelectedTypeName(null);
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, []);

    return (
        <div id="App">
            <div className="app-header">
                <h1>Graph Explorer</h1>
            </div>
            <div className="app-container">
                <div className="sidebar">
                    <SchemaPanel
                        schema={schema}
                        loading={schemaLoading}
                        error={schemaError}
                        selectedTypeName={selectedTypeName}
                        onRefresh={handleRefreshSchema}
                        onTypeClick={handleTypeClick}
                    />
                </div>
                <div className="main-content">
                    <FilterBar
                        schema={schema}
                        onFilterChange={handleFilterChange}
                        initialFilter={activeFilter}
                    />
                    {graphLoading && (
                        <div className="loading-overlay">
                            <div className="spinner"></div>
                            <p>Loading graph...</p>
                        </div>
                    )}
                    {graphError && (
                        <div className="error-message">
                            <p>⚠️ {graphError}</p>
                            <button onClick={() => loadRandomNodes(100)}>Retry</button>
                        </div>
                    )}
                    {!graphLoading && !graphError && graphData.nodes.length > 0 && (
                        <GraphCanvas
                            data={graphData}
                            selectedNodeId={selectedNode?.id || null}
                            highlightedNodeIds={highlightedNodeIds}
                            expandedNodes={expandedNodes}
                            onNodeClick={handleNodeClick}
                            onNodeRightClick={handleNodeRightClick}
                            onLoadMore={handleLoadMore}
                        />
                    )}
                    {!graphLoading && !graphError && graphData.nodes.length === 0 && (
                        <div className="loading-overlay">
                            <p>No nodes to display</p>
                        </div>
                    )}
                </div>
                <div className="sidebar-right">
                    {selectedType ? (
                        <div className="type-details-panel">
                            <div className="detail-header">
                                <div className="detail-title">
                                    <h2>{selectedType.name}</h2>
                                    <span className={`type-badge ${selectedType.name.toLowerCase()}`}>
                                        {selectedType.count} instances
                                    </span>
                                </div>
                                <button className="close-button" onClick={() => { setSelectedType(null); setSelectedTypeName(null); }} title="Close (Escape)">
                                    ✕
                                </button>
                            </div>
                            {detailsLoading ? (
                                <div className="detail-loading">
                                    <div className="spinner-large"></div>
                                    <p>Loading type details...</p>
                                </div>
                            ) : (
                                <div className="detail-content">
                                    <div className="detail-section">
                                        <h3>Properties ({selectedType.properties?.length || 0})</h3>
                                        {selectedType.properties && selectedType.properties.length > 0 ? (
                                            <div className="properties-list">
                                                {selectedType.properties.map((prop, idx) => (
                                                    <div key={idx} className="property-item">
                                                        <div className="property-key">{prop.name}</div>
                                                        <div className="property-meta">
                                                            <span className="property-type">{prop.data_type}</span>
                                                            {prop.sample_value && prop.sample_value.length > 0 && (
                                                                <div className="sample-values">
                                                                    <strong>Samples:</strong>
                                                                    <ul>
                                                                        {prop.sample_value.slice(0, 3).map((val, i) => (
                                                                            <li key={i}>{String(val)}</li>
                                                                        ))}
                                                                    </ul>
                                                                </div>
                                                            )}
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        ) : (
                                            <p className="empty-message">No properties defined</p>
                                        )}
                                    </div>
                                </div>
                            )}
                        </div>
                    ) : (
                        <DetailPanel
                            node={selectedNode}
                            loading={loadingNodeId === selectedNode?.id}
                            expandedNodeState={selectedNode ? expandedNodes.get(selectedNode.id) || null : null}
                            onLoadMore={handleLoadMore}
                            onClose={handleCloseDetail}
                            relatedEntities={relatedEntities}
                            onExpandRelationship={handleExpandRelationship}
                        />
                    )}
                </div>
            </div>
        </div>
    );
}

export default App;
