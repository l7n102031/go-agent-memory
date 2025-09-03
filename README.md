# Go Agent Memory 🧠

[![Go Reference](https://pkg.go.dev/badge/github.com/framehood/go-agent-memory.svg)](https://pkg.go.dev/github.com/framehood/go-agent-memory)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/framehood/go-agent-memory)](https://goreportcard.com/report/github.com/framehood/go-agent-memory)

A modular, production-ready memory system for AI agents with semantic search capabilities. Uses **Supabase** (PostgreSQL + pgvector) for long-term semantic memory and optional **Redis** for blazing-fast session caching.

## Features ✨

- **📝 Session Memory**: Fast retrieval of recent conversation history
- **🔍 Semantic Search**: Find relevant past conversations using vector similarity
- **💾 Hybrid Storage**: Redis for speed + Supabase for persistence
- **🎯 Auto-Summarization**: Compress old conversations to save tokens
- **🔌 Modular Design**: Use as a simple import, doesn't affect your app if not used
- **⚡ Production Ready**: Connection pooling, error handling, and graceful degradation

## Documentation 📚

- [**Architecture Overview**](docs/ARCHITECTURE.md) - System design and data flow
- [**Integration Guide**](docs/INTEGRATE_WITH_AGENT.md) - Step-by-step agent integration
- [**Summary**](docs/SUMMARY.md) - Quick overview and key features

## Repository Structure 📁

```
go-agent-memory/
├── memory.go           # Core interfaces and types
├── supabase.go        # Supabase/pgvector implementation
├── hybrid.go          # Redis + Supabase hybrid
├── docs/              # Documentation
│   ├── ARCHITECTURE.md
│   ├── INTEGRATE_WITH_AGENT.md
│   └── SUMMARY.md
├── tests/             # Test files
│   └── memory_test.go
├── examples/          # Usage examples
│   └── integration.go
├── scripts/           # Utility scripts
│   └── quickstart.sh
└── deployment/        # Deployment configs
    ├── docker-compose.yml
    └── init.sql
```

## Quick Start 🚀

### 1. Install the Package

```bash
go get github.com/framehood/go-agent-memory
```

### 2. Set Up Supabase

Ensure pgvector is enabled in your Supabase instance:

```sql
-- Run this in Supabase SQL Editor
CREATE EXTENSION IF NOT EXISTS vector;
```

### 3. Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    memory "github.com/framehood/go-agent-memory"
)

func main() {
    // Initialize with just Supabase (no Redis)
    mem, err := memory.New(memory.Config{
        DatabaseURL:    "postgresql://user:pass@host:5432/dbname",
        OpenAIKey:      "your-openai-key",
        EmbeddingModel: "text-embedding-3-small",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer mem.Close()
    
    // Add a message
    err = mem.AddMessage(context.Background(), memory.Message{
        ID:      "msg-123",
        Role:    "user",
        Content: "What's the weather like?",
        Metadata: memory.Metadata{
            SessionID: "session-456",
            UserID:    "user-789",
        },
    })
    
    // Get recent messages
    messages, _ := mem.GetRecentMessages(context.Background(), "session-456", 10)
    
    // Semantic search
    results, _ := mem.Search(context.Background(), 
        "weather forecast", 5, 0.7)
    
    for _, result := range results {
        fmt.Printf("Found: %s (score: %.2f)\n", 
            result.Message.Content, result.Score)
    }
}
```

## Integration with Your Agent 🤖

Here's how to add memory to your existing agent:

```go
// In your agent/main.go

package main

import (
    "context"
    "os"
    
    memory "github.com/framehood/go-agent-memory"
    "github.com/openai/openai-go/v2"
)

var mem memory.Memory // Optional memory instance

func initMemory() {
    // Only initialize if environment variables are set
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        fmt.Println("Memory disabled: DATABASE_URL not set")
        return
    }
    
    var err error
    mem, err = memory.New(memory.Config{
        DatabaseURL:    dbURL,
        OpenAIKey:      os.Getenv("OPENAI_API_KEY"),
        
        // Optional Redis for faster session access
        RedisAddr:      os.Getenv("REDIS_URL"), // e.g., "localhost:6379"
        RedisPassword:  os.Getenv("REDIS_PASSWORD"),
        
        // Settings
        MaxSessionMessages: 50,
        SessionTTL:        24 * time.Hour,
        AutoSummarize:     true,
    })
    
    if err != nil {
        fmt.Printf("Warning: Memory initialization failed: %v\n", err)
        mem = nil
    }
}

func main() {
    // Initialize memory (optional)
    initMemory()
    if mem != nil {
        defer mem.Close()
    }
    
    // Your existing agent code...
    
    // When processing messages:
    handleUserMessage := func(sessionID, content string) {
        // Store user message if memory is enabled
        if mem != nil {
            mem.AddMessage(context.Background(), memory.Message{
                ID:      generateID(),
                Role:    "user",
                Content: content,
                Metadata: memory.Metadata{
                    SessionID: sessionID,
                },
            })
        }
        
        // Get context from memory
        var contextMessages []openai.ChatCompletionMessageParamUnion
        if mem != nil {
            // Get recent conversation
            recent, _ := mem.GetRecentMessages(context.Background(), sessionID, 10)
            
            // Search for relevant past conversations
            similar, _ := mem.Search(context.Background(), content, 3, 0.75)
            
            // Add to context (you'd format this appropriately)
            for _, msg := range recent {
                // Add to contextMessages...
            }
        }
        
        // Continue with OpenAI call...
    }
}
```

## Configuration Options 🔧

### Minimal Config (Supabase Only)
```go
memory.Config{
    DatabaseURL: "postgresql://...",  // Required
    OpenAIKey:   "sk-...",           // Required for embeddings
}
```

### Full Config (Hybrid Mode)
```go
memory.Config{
    // Supabase (Required)
    SupabaseURL:  "https://xxx.supabase.co",
    SupabaseKey:  "your-anon-key",
    DatabaseURL:  "postgresql://...",
    
    // Redis (Optional - enables fast session cache)
    RedisAddr:     "localhost:6379",
    RedisPassword: "optional-password",
    RedisDB:       0,
    
    // OpenAI (Required for embeddings)
    OpenAIKey:      "sk-...",
    EmbeddingModel: "text-embedding-3-small", // or text-embedding-3-large
    
    // Memory Settings
    MaxSessionMessages: 50,           // Keep last N messages in fast cache
    SessionTTL:        24 * time.Hour, // Redis cache expiry
    AutoSummarize:     true,          // Auto-summarize old conversations
    VectorDimension:   1536,          // Match your embedding model
}
```

## Environment Variables 🌍

```bash
# Required
export DATABASE_URL="postgresql://postgres:password@db.supabase.co:5432/postgres"
export OPENAI_API_KEY="sk-..."

# Optional (for Redis caching)
export REDIS_URL="localhost:6379"
export REDIS_PASSWORD=""

# Optional Supabase (if using REST APIs)
export SUPABASE_URL="https://xxx.supabase.co"
export SUPABASE_ANON_KEY="eyJ..."
```

## Features in Detail 📚

### 1. **Session Memory**
- Recent messages cached in Redis (if available)
- Falls back to PostgreSQL if Redis is unavailable
- Automatic TTL and size limits

### 2. **Semantic Search**
- Uses OpenAI embeddings (text-embedding-3-small/large)
- HNSW index for fast similarity search
- Configurable similarity threshold

### 3. **Auto-Summarization**
- Compress old conversations to save context tokens
- Summaries stored in separate table
- Cached in Redis for fast access

### 4. **Graceful Degradation**
- Works without Redis (Supabase only)
- Continues if embedding generation fails
- Non-blocking background operations

## Performance 🏎️

### With Redis + Supabase (Hybrid)
- Session retrieval: **~2-5ms**
- Message storage: **~10ms**
- Semantic search: **~50-100ms**

### With Supabase Only
- Session retrieval: **~20-50ms**
- Message storage: **~20-30ms**
- Semantic search: **~50-100ms**

## Database Schema 📊

The package automatically creates these tables:

```sql
-- Messages table
CREATE TABLE agent_messages (
    id SERIAL PRIMARY KEY,
    message_id TEXT UNIQUE,
    session_id TEXT,
    user_id TEXT,
    role TEXT,
    content TEXT,
    metadata JSONB,
    embedding vector(1536),
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

-- Summaries table  
CREATE TABLE agent_summaries (
    id SERIAL PRIMARY KEY,
    session_id TEXT,
    summary TEXT,
    message_count INT,
    token_count INT,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ
);
```

## Advanced Usage 🎓

### Custom Embedding Models
```go
mem, _ := memory.New(memory.Config{
    DatabaseURL:     dbURL,
    OpenAIKey:       apiKey,
    EmbeddingModel:  "text-embedding-3-large",
    VectorDimension: 3072, // Large model dimension
})
```

### Search with Pre-computed Embeddings
```go
// If you already have an embedding
embedding := []float32{0.1, 0.2, ...} // 1536 dimensions
results, _ := mem.SearchWithEmbedding(ctx, embedding, 10, 0.8)
```

### Get Memory Statistics
```go
stats, _ := mem.GetStats(ctx, "session-123")
fmt.Printf("Total messages: %d\n", stats.TotalMessages)
fmt.Printf("Total tokens: %d\n", stats.TotalTokens)
```

### Summarize Long Conversations
```go
summary, _ := mem.Summarize(ctx, "session-123", 4000) // Max 4000 tokens
fmt.Println("Summary:", summary)
```

## Testing 🧪

```bash
# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

## Contributing 🤝

PRs welcome! Please ensure:
1. Tests pass
2. Code is formatted (`go fmt`)
3. No linting issues (`golangci-lint run`)

## License 📄

MIT License - feel free to use in your projects!

## Support 💬

- Issues: [GitHub Issues](https://github.com/framehood/go-agent-memory/issues)
- Discussions: [GitHub Discussions](https://github.com/framehood/go-agent-memory/discussions)
