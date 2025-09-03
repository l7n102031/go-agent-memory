-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create a test user and grant permissions
CREATE USER agent_user WITH PASSWORD 'agent_password';
GRANT ALL PRIVILEGES ON DATABASE agent_memory TO agent_user;
GRANT ALL ON SCHEMA public TO agent_user;

-- Create initial tables will be handled by the Go code
-- This file just ensures the extension is available
