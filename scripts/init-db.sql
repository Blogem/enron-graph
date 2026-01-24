-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Verify extension is loaded
SELECT * FROM pg_extension WHERE extname = 'vector';
