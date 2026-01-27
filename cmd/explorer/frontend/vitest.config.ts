import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
    plugins: [react({
        jsxRuntime: 'automatic',
        fastRefresh: false // Disable Fast Refresh for tests
    })],
    test: {
        globals: true,
        environment: 'jsdom',
        setupFiles: './src/test/setup.ts',
    },
});
