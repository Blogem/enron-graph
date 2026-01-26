// Wails API service - wraps auto-generated bindings
import {
    GetSchema,
    GetTypeDetails,
    RefreshSchema,
    GetRandomNodes,
    GetRelationships,
    GetNodeDetails
} from '../wailsjs/go/main/App';
import type { explorer } from '../wailsjs/go/models';
import type { GraphResponse, RelationshipsResponse, GraphNode } from '../types/graph';

export const wailsAPI = {
    // Schema operations
    async getSchema(): Promise<explorer.SchemaResponse> {
        return await GetSchema();
    },

    async getTypeDetails(typeName: string): Promise<explorer.SchemaType> {
        return await GetTypeDetails(typeName);
    },

    async refreshSchema(): Promise<void> {
        return await RefreshSchema();
    },

    // Graph operations
    async getRandomNodes(limit: number): Promise<GraphResponse> {
        return await GetRandomNodes(limit);
    },

    async getRelationships(nodeId: string, offset: number, limit: number): Promise<RelationshipsResponse> {
        return await GetRelationships(nodeId, offset, limit);
    },

    async getNodeDetails(nodeId: string): Promise<GraphNode> {
        return await GetNodeDetails(nodeId);
    },
};
