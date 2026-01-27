// Wails API service - wraps auto-generated bindings
import {
    GetSchema,
    GetTypeDetails,
    RefreshSchema,
    GetRandomNodes,
    GetRelationships,
    GetNodeDetails,
    GetNodes
} from '../wailsjs/go/main/App';
import type { explorer } from '../wailsjs/go/models';
import type { NodeFilter } from '../types/graph';

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
    async getRandomNodes(limit: number): Promise<explorer.GraphResponse> {
        return await GetRandomNodes(limit);
    },

    async getRelationships(nodeId: string, offset: number, limit: number): Promise<explorer.RelationshipsResponse> {
        return await GetRelationships(nodeId, offset, limit);
    },

    async getNodeDetails(nodeId: string): Promise<explorer.GraphNode> {
        return await GetNodeDetails(nodeId);
    },

    async getNodes(filter: NodeFilter): Promise<explorer.GraphResponse> {
        // Convert TypeScript filter to Go-compatible format
        const goFilter: any = {
            types: filter.types || [],
            category: filter.category || '',
            search_query: filter.search_query || '',
            limit: filter.limit || 100
        };
        return await GetNodes(goFilter);
    },
};
