import React, { useRef, useCallback, useEffect, useState } from 'react';
import ForceGraph2D from 'react-force-graph-2d';
import type { GraphData, GraphNodeWithPosition, GraphEdge, ExpandedNodeState } from '../types/graph';
import './GraphCanvas.css';

interface GraphCanvasProps {
    data: GraphData;
    selectedNodeId: string | null;
    expandedNodes: Map<string, ExpandedNodeState>;
    onNodeClick: (node: GraphNodeWithPosition) => void;
    onNodeRightClick: (node: GraphNodeWithPosition) => void;
    onLoadMore?: (nodeId: string) => void;
}

const GraphCanvas: React.FC<GraphCanvasProps> = ({
    data,
    selectedNodeId,
    expandedNodes,
    onNodeClick,
    onNodeRightClick,
    onLoadMore
}) => {
    const graphRef = useRef<any>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const [dimensions, setDimensions] = useState({ width: 800, height: 600 });

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

    // Node color differentiation by entity type
    const getNodeColor = useCallback((node: GraphNodeWithPosition) => {
        if (node.id === selectedNodeId) {
            return '#ff6b6b'; // Highlight selected node in red
        }

        // Color by type
        const typeColors: Record<string, string> = {
            'Email': '#4ecdc4',
            'Person': '#95e1d3',
            'Organization': '#f38181',
            'Relationship': '#aa96da',
            'DiscoveredEntity': '#fcbad3'
        };

        return typeColors[node.type] || '#a8dadc'; // Default color
    }, [selectedNodeId]);

    // Node size based on degree (relationship count)
    const getNodeSize = useCallback((node: GraphNodeWithPosition) => {
        const degree = node.properties?.degree || 0;
        // Scale size from 4 to 12 based on degree
        return Math.min(12, 4 + Math.log(degree + 1) * 2);
    }, []);

    // Node label with relationship count indicator
    const getNodeLabel = useCallback((node: GraphNodeWithPosition) => {
        const degree = node.properties?.degree || 0;
        const label = node.label || node.id;

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
                cooldownTicks={100}
                warmupTicks={100}
                d3AlphaDecay={0.02}
                d3VelocityDecay={0.3}
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
