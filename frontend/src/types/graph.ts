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
    relationships: GraphEdge[];
    total: number;
}

export interface NodeFilter {
    types?: string[];
    category?: string;
    search_query?: string;
    limit?: number;
}
