# Architecture Overview

## System Design

The Go Agent Memory module follows a modular, interface-based design that allows for flexible storage backends while maintaining a simple API surface.

```
┌─────────────────────────────────────┐
│         Your Application            │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      Memory Interface (API)         │
├─────────────────────────────────────┤
│  • AddMessage()                     │
│  • GetRecentMessages()              │
│  • Search()                         │
│  • Summarize()                      │
└──────────┬──────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│    Implementation Layer              │
├──────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐ │
│  │ SupabaseMemory│  │ HybridMemory │ │
│  └──────────────┘  └──────────────┘ │
└──────────────────────────────────────┘
           │                │
           ▼                ▼
┌──────────────────┐  ┌──────────────┐
│  Supabase/PG     │  │    Redis     │
│  + pgvector      │  │   (Cache)    │
└──────────────────┘  └──────────────┘
```

## Core Components

### 1. Memory Interface (`memory.go`)
- Defines the contract that all memory implementations must follow
- Provides a unified API regardless of the underlying storage
- Includes types for Message, Metadata, SearchResult, and Stats

### 2. SupabaseMemory (`supabase.go`)
- Direct PostgreSQL implementation using pgvector
- Handles embedding generation via OpenAI API
- Manages database schema initialization
- Provides semantic search capabilities

### 3. HybridMemory (`hybrid.go`)
- Combines Redis for fast session access with Supabase for persistence
- Implements write-through caching strategy
- Provides automatic fallback to Supabase if Redis is unavailable
- Manages cache invalidation and TTL

## Data Flow

### Message Storage Flow
```
1. Application calls AddMessage()
2. Generate embedding (if not provided)
3. If Hybrid:
   a. Write to Redis (fast cache)
   b. Write to Supabase (persistence)
4. If Supabase-only:
   a. Write directly to PostgreSQL
```

### Message Retrieval Flow
```
1. Application calls GetRecentMessages()
2. If Hybrid:
   a. Check Redis first
   b. If miss, query Supabase
   c. Optionally repopulate Redis cache
3. If Supabase-only:
   a. Query PostgreSQL directly
```

### Semantic Search Flow
```
1. Application calls Search()
2. Generate embedding for query
3. Perform vector similarity search in pgvector
4. Return ranked results with similarity scores
```

## Database Schema

### PostgreSQL Tables

#### agent_messages
- Stores all conversation messages
- Contains embeddings as vector type
- Indexed for fast retrieval by session_id
- HNSW index for efficient similarity search

#### agent_summaries
- Stores conversation summaries
- Linked to sessions
- Contains token counts and time ranges

## Performance Considerations

### Caching Strategy
- Redis stores recent messages (configurable limit)
- Session TTL prevents unbounded growth
- Background cache population after cache misses

### Embedding Generation
- Async generation where possible
- Graceful degradation if embedding fails
- Caching of embeddings to avoid regeneration

### Database Optimization
- HNSW index for fast vector search
- Connection pooling via pgxpool
- Prepared statements for common queries

## Error Handling

The module follows a fail-safe design:
- Redis failures don't block operations
- Embedding failures don't prevent message storage
- Database connection issues are properly propagated
- All errors are returned to the caller for handling

## Extensibility

New storage backends can be added by:
1. Implementing the Memory interface
2. Adding initialization logic to New() function
3. Following the same patterns as existing implementations

## Security Considerations

- Database credentials via environment variables
- No hardcoded secrets
- Prepared statements to prevent SQL injection
- Proper context handling for cancellation
