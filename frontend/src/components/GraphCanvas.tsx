import React, { useRef, useCallback, useEffect, useState, useMemo } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import type { GraphData, GraphNodeWithPosition, GraphEdge, ExpandedNodeState } from '../types/graph';
import './GraphCanvas.css';

interface GraphCanvasProps {
    data: GraphData;
    selectedNodeId: string | null;
    highlightedNodeIds?: Set<string>;
    expandedNodes: Map<string, ExpandedNodeState>;
    onNodeClick: (node: GraphNodeWithPosition) => void;
    onNodeRightClick: (node: GraphNodeWithPosition) => void;
    onLoadMore?: (nodeId: string) => void;
}

const GraphCanvas: React.FC<GraphCanvasProps> = ({
    data,
    selectedNodeId,
    highlightedNodeIds = new Set(),
    expandedNodes,
    onNodeClick,
    onNodeRightClick,
    onLoadMore
}) => {
    const graphRef = useRef<any>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [dimensions, setDimensions] = useState({ width: 800, height: 600 });

    // Performance optimization: Enable particle rendering for large graphs (T105)
    const isLargeGraph = data.nodes.length > 500;

    // Performance optimization: Memoize node color lookup map (T105)
    const nodeColorMap = useMemo(() => {
        const map = new Map<string, string>();

        data.nodes.forEach(node => {
            // Check if node is a ghost node (is_ghost property)
            const isGhost = node.properties?.is_ghost === true;

            if (isGhost) {
                // Ghost nodes: greyed out, semi-transparent (FR-007a)
                map.set(node.id, 'rgba(128, 128, 128, 0.3)');
            } else if (node.id === selectedNodeId) {
                map.set(node.id, '#ff6b6b'); // Highlight selected node in red
            } else if (highlightedNodeIds.has(node.id)) {
                // Highlighted search result nodes (T092)
                map.set(node.id, '#fbbf24'); // Bright amber/yellow for search matches
            } else {
                // Color by type
                const typeColors: Record<string, string> = {
                    'Email': '#4ecdc4',
                    'Person': '#95e1d3',
                    'Organization': '#f38181',
                    'Relationship': '#aa96da',
                    'DiscoveredEntity': '#fcbad3'
                };
                map.set(node.id, typeColors[node.type] || '#a8dadc');
            }
        });

        return map;
    }, [data.nodes, selectedNodeId, highlightedNodeIds]);

    // Update dimensions on mount and resize
    useEffect(() => {
        const updateDimensions = () => {
            if (containerRef.current) {
                const rect = containerRef.current.getBoundingClientRect();
                setDimensions({
                    width: rect.width,
                    height: rect.height
                });
            }
        };

        // Initial update with a slight delay to ensure layout is ready
        const timer = setTimeout(updateDimensions, 100);
        updateDimensions();

        window.addEventListener('resize', updateDimensions);
        return () => {
            clearTimeout(timer);
            window.removeEventListener('resize', updateDimensions);
        };
    }, []);

    // Node color differentiation by entity type (optimized with memoized map - T105)
    const getNodeColor = useCallback((node: GraphNodeWithPosition) => {
        return nodeColorMap.get(node.id) || '#a8dadc';
    }, [nodeColorMap]);

    // Node size based on degree (relationship count)
    const getNodeSize = useCallback((node: GraphNodeWithPosition) => {
        const degree = node.properties?.degree || 0;
        // Scale size from 4 to 12 based on degree
        return Math.min(12, 4 + Math.log(degree + 1) * 2);
    }, []);

    // Node label with relationship count indicator
    const getNodeLabel = useCallback((node: GraphNodeWithPosition) => {
        const isGhost = node.properties?.is_ghost === true;
        const degree = node.properties?.degree || 0;
        const label = node.label || node.id;

        // Ghost nodes have minimal label
        if (isGhost) {
            return `${label} (ghost)`;
        }

        // Show relationship count for nodes with relationships
        if (degree > 0) {
            const expandedState = expandedNodes.get(node.id);
            const loadedCount = expandedState ? expandedState.offset : 0;
            return `${label}\n(${loadedCount}/${degree} edges)`;
        }

        return label;
    }, [expandedNodes]);

    // Handle node click
    const handleNodeClick = useCallback((node: any) => {
        onNodeClick(node as GraphNodeWithPosition);
    }, [onNodeClick]);

    // Handle node right-click for expansion
    const handleNodeRightClick = useCallback((node: any, event: MouseEvent) => {
        event.preventDefault();
        onNodeRightClick(node as GraphNodeWithPosition);
    }, [onNodeRightClick]);

    // Recenter graph view
    const handleRecenter = useCallback(() => {
        if (graphRef.current) {
            graphRef.current.zoomToFit(400, 20);
        }
    }, []);

    // Auto-zoom to highlighted search results
    useEffect(() => {
        if (highlightedNodeIds.size > 0 && graphRef.current) {
            // Zoom to fit highlighted nodes after a brief delay
            setTimeout(() => {
                if (graphRef.current) {
                    // Get the bounding box of highlighted nodes
                    const highlightedNodes = data.nodes.filter(n => highlightedNodeIds.has(n.id));

                    if (highlightedNodes.length > 0) {
                        // Center on the highlighted nodes
                        const avgX = highlightedNodes.reduce((sum, n) => sum + (n.x || 0), 0) / highlightedNodes.length;
                        const avgY = highlightedNodes.reduce((sum, n) => sum + (n.y || 0), 0) / highlightedNodes.length;

                        graphRef.current.centerAt(avgX, avgY, 400);
                        graphRef.current.zoom(3, 400);
                    }
                }
            }, 500);
        }
    }, [highlightedNodeIds, data.nodes]);

    return (
        <div className="graph-canvas-container" ref={containerRef}>
            <div className="graph-controls">
                <button onClick={handleRecenter} title="Recenter view (Space)">
                    ⌖ Recenter
                </button>
                <div className="graph-stats">
                    {data.nodes.length} nodes, {data.links.length} edges
                </div>
            </div>
            <ForceGraph2D
                ref={graphRef}
                width={dimensions.width}
                height={dimensions.height}
                graphData={data}
                nodeId="id"
                nodeLabel={getNodeLabel}
                nodeColor={getNodeColor}
                nodeVal={(node: any) => {
                    const degree = node.properties?.degree || 0;
                    return Math.min(12, 4 + Math.log(degree + 1) * 2);
                }}
                nodeCanvasObjectMode={() => 'after'}
                nodeCanvasObject={(node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
                    const isGhost = node.properties?.is_ghost === true;

                    // Skip badges for ghost nodes
                    if (isGhost) {
                        return;
                    }

                    // Performance optimization: Skip custom rendering at low zoom levels for large graphs (T105)
                    if (isLargeGraph && globalScale < 1.5) {
                        return;
                    }

                    // Draw search highlight ring
                    if (highlightedNodeIds.has(node.id)) {
                        const nodeSize = Math.min(12, 4 + Math.log((node.properties?.degree || 0) + 1) * 2);
                        ctx.beginPath();
                        ctx.arc(node.x!, node.y!, nodeSize * 1.8, 0, 2 * Math.PI);
                        ctx.strokeStyle = '#fbbf24';
                        ctx.lineWidth = 3 / globalScale;
                        ctx.stroke();
                    }

                    // Draw relationship count badge for nodes with relationships
                    const degree = node.properties?.degree || 0;
                    if (degree > 0) {
                        const nodeSize = Math.min(12, 4 + Math.log(degree + 1) * 2);
                        const badgeSize = 6 / globalScale;
                        ctx.beginPath();
                        ctx.arc(node.x! + nodeSize * 0.7, node.y! - nodeSize * 0.7, badgeSize, 0, 2 * Math.PI);
                        ctx.fillStyle = '#ff6b6b';
                        ctx.fill();
                        ctx.fillStyle = '#fff';
                        ctx.font = `bold ${badgeSize * 1.2}px Sans-Serif`;
                        ctx.textAlign = 'center';
                        ctx.textBaseline = 'middle';
                        ctx.fillText(String(degree > 99 ? '99+' : degree), node.x! + nodeSize * 0.7, node.y! - nodeSize * 0.7);
                    }
                }}
                linkDirectionalArrowLength={6}
                linkDirectionalArrowRelPos={1}
                linkColor={() => 'rgba(0,0,0,0.2)'}
                linkWidth={1}
                onNodeClick={handleNodeClick}
                onNodeRightClick={handleNodeRightClick}
                enableNodeDrag={true}
                enableZoomInteraction={true}
                enablePanInteraction={true}
                cooldownTicks={isLargeGraph ? 50 : 100}
                warmupTicks={isLargeGraph ? 50 : 100}
                d3AlphaDecay={isLargeGraph ? 0.05 : 0.02}
                d3VelocityDecay={isLargeGraph ? 0.5 : 0.3}
                nodePointerAreaPaint={(node: any, color: string, ctx: CanvasRenderingContext2D) => {
                    // Increase clickable area for better UX
                    ctx.fillStyle = color;
                    const degree = node.properties?.degree || 0;
                    const nodeSize = Math.min(12, 4 + Math.log(degree + 1) * 2);
                    ctx.beginPath();
                    ctx.arc(node.x!, node.y!, nodeSize * 1.5, 0, 2 * Math.PI);
                    ctx.fill();
                }}
            />
        </div>
    );
};

export default GraphCanvas;
