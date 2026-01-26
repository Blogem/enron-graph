// Wails API service - wraps auto-generated bindings
import { GetSchema, GetTypeDetails, RefreshSchema } from '../wailsjs/go/main/App';
import type { explorer } from '../wailsjs/go/models';

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
};
