// TypeScript type definitions for Graph Explorer
// These mirror the Go structs in internal/explorer/models.go

export interface PropertyDefinition {
    name: string;
    type: string;
    sample_values: string[];
}

export interface SchemaType {
    name: string;
    category: string;
    count: number;
    properties: PropertyDefinition[];
    is_promoted: boolean;
    relationships: string[];
}

export interface SchemaResponse {
    promoted_types: SchemaType[];
    discovered_types: SchemaType[];
    total_entities: number;
}

export interface GraphNode {
    id: string;
    label: string;
    type: string;
    properties: Record<string, any>;
}

export interface GraphEdge {
    id: string;
    source: string;
    target: string;
    type: string;
    properties: Record<string, any>;
}

export interface GraphResponse {
    nodes: GraphNode[];
    edges: GraphEdge[];
    total_nodes: number;
    total_edges: number;
    has_more: boolean;
}

export interface RelationshipsResponse {
    nodes: GraphNode[];
    edges: GraphEdge[];
    total_count: number;
    has_more: boolean;
    offset: number;
}

export interface NodeFilter {
    types?: string[];
    category?: string;
    search_query?: string;
    limit?: number;
}

// Types for force-directed graph visualization
export interface GraphNodeWithPosition extends GraphNode {
    x?: number;
    y?: number;
    vx?: number;
    vy?: number;
    fx?: number;
    fy?: number;
}

export interface GraphData {
    nodes: GraphNodeWithPosition[];
    links: GraphEdge[];
}

// State tracking for expanded nodes with batched loading
export interface ExpandedNodeState {
    nodeId: string;
    offset: number;
    hasMore: boolean;
    totalRelationships: number;
}
