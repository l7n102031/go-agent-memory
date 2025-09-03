# Go Agent Memory 🧠

[![Go Reference](https://pkg.go.dev/badge/github.com/framehood/go-agent-memory.svg)](https://pkg.go.dev/github.com/framehood/go-agent-memory)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/framehood/go-agent-memory)](https://goreportcard.com/report/github.com/framehood/go-agent-memory)

A modular, production-ready memory system for AI agents with multiple deployment modes. Supports **session-only** (zero dependencies), **persistent** (PostgreSQL + pgvector), and **hybrid** (Redis + PostgreSQL) configurations.

## Features ✨

- **🚀 Multiple Modes**: Session-only, Persistent, or Hybrid deployment
- **📝 Session Memory**: Fast retrieval of recent conversation history
- **🔍 Semantic Search**: Find relevant past conversations using vector similarity
- **💾 Flexible Storage**: In-memory, PostgreSQL, or Redis + PostgreSQL
- **🎯 Auto-Summarization**: Compress old conversations to save tokens
- **🔧 Feature Flags**: Enable/disable features as needed
- **🔌 Zero Dependencies**: Start with session-only mode, add features later
- **⚡ Production Ready**: Connection pooling, error handling, and graceful degradation

## Documentation 📚

- [**Implementation Guide**](docs/IMPLEMENTATION_GUIDE.md) - Complete integration guide
- [**Architecture Overview**](docs/ARCHITECTURE.md) - System design and data flow
- [**Examples**](examples/) - 7 complete examples from simple to production
- [**Integration Guide**](docs/INTEGRATE_WITH_AGENT.md) - Step-by-step agent integration
- [**Summary**](docs/SUMMARY.md) - Quick overview and key features

## Repository Structure 📁

```
go-agent-memory/
├── memory.go           # Core interfaces and types
├── session_only.go     # In-memory implementation (zero deps)
├── supabase.go        # PostgreSQL + pgvector implementation
├── hybrid.go          # Redis + PostgreSQL hybrid
├── docs/              # Documentation
│   ├── IMPLEMENTATION_GUIDE.md
│   ├── ARCHITECTURE.md
│   ├── INTEGRATE_WITH_AGENT.md
│   └── SUMMARY.md
├── tests/             # Test files
│   └── memory_test.go
├── examples/          # Complete examples (7 modes)
│   ├── 01-session-only/
│   ├── 02-persistent-basic/
│   ├── 03-hybrid-mode/
│   ├── 04-semantic-search/
│   ├── 05-auto-summarization/
│   ├── 06-event-streaming/
│   ├── 07-agent-integration/
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

### 2. Choose Your Mode

**Option A: Zero Dependencies (Session-Only)**
```go
mem, _ := memory.NewWithConfig(memory.Config{
    Mode: memory.SESSION_ONLY,
})
// No database or Redis needed!
```

**Option B: Persistent (PostgreSQL)**
```go
mem, _ := memory.NewWithConfig(memory.Config{
    Mode:        memory.PERSISTENT,
    DatabaseURL: "postgresql://user:pass@host:5432/dbname",
    OpenAIKey:   "sk-...", // For semantic search
})
```

**Option C: Hybrid (Redis + PostgreSQL)**
```go
mem, _ := memory.NewWithConfig(memory.Config{
    Mode:        memory.HYBRID,
    DatabaseURL: "postgresql://user:pass@host:5432/dbname",
    RedisAddr:   "localhost:6379",
    OpenAIKey:   "sk-...",
})
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
    // Initialize memory (auto-detects mode from config)
    mem, err := memory.NewWithConfig(memory.Config{
        Mode:           memory.PERSISTENT,
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
    mem, err = memory.NewWithConfig(memory.Config{
        Mode:        memory.HYBRID, // Or PERSISTENT, SESSION_ONLY
        DatabaseURL: dbURL,
        OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
        
        // Optional Redis for faster session access
        RedisAddr:     os.Getenv("REDIS_URL"), // e.g., "localhost:6379"
        RedisPassword: os.Getenv("REDIS_PASSWORD"),
        
        // Feature flags
        EnableSemanticSearch: true,
        EnableAutoSummarize:  true,
        
        // Settings
        MaxSessionMessages: 50,
        SessionTTL:        24 * time.Hour,
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

### Session-Only Mode (Zero Dependencies)
```go
memory.Config{
    Mode: memory.SESSION_ONLY,
    MaxSessionMessages: 50, // Optional: default 50
}
```

### Persistent Mode (PostgreSQL)
```go
memory.Config{
    Mode:        memory.PERSISTENT,
    DatabaseURL: "postgresql://...",  // Required
    OpenAIKey:   "sk-...",           // Optional: for semantic search
    
    // Feature flags
    EnableSemanticSearch: true,  // Enable vector search
    EnableAutoSummarize:  true,  // Enable summarization
}
```

### Hybrid Mode (Redis + PostgreSQL)
```go
memory.Config{
    Mode:        memory.HYBRID,
    DatabaseURL: "postgresql://...",  // Required
    RedisAddr:   "localhost:6379",   // Required
    OpenAIKey:   "sk-...",           // Optional: for semantic search
    
    // Feature flags
    EnableSemanticSearch: true,
    EnableAutoSummarize:  true,
    
    // Performance settings
    MaxSessionMessages: 30,           // Messages in Redis cache
    SessionTTL:        2 * time.Hour, // Redis cache expiry
    
    // Summarization settings
    SummarizeThreshold: 50,          // Messages before summarization
    SummarizeMaxTokens: 500,         // Target summary length
    SummarizeModel:     "gpt-3.5-turbo",
    
    // Search settings
    DefaultSearchLimit:     5,
    DefaultSearchThreshold: 0.7,
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

### Session-Only Mode
- Session retrieval: **~1μs** (in-memory)
- Message storage: **~1μs** (in-memory)
- Search: **~1ms** (basic text matching)

### Persistent Mode (PostgreSQL)
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
mem, _ := memory.NewWithConfig(memory.Config{
    Mode:            memory.PERSISTENT,
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
fmt.Printf("Has summary: %v\n", stats.HasSummary)
fmt.Printf("Active tokens: %d\n", stats.ActiveTokens)
```

### Get Conversation Summary
```go
summary, _ := mem.GetSummary(ctx, "session-123")
fmt.Printf("Summary: %s\n", summary.Content)
fmt.Printf("Covers %d messages\n", summary.MessageCount)
```

### Cache Management (Hybrid Mode)
```go
if hybridMem, ok := mem.(*memory.HybridMemory); ok {
    // Clear cache for session
    hybridMem.ClearCache(ctx, "session-123")
    
    // Get cache statistics
    stats, _ := hybridMem.GetCacheStats(ctx)
    fmt.Printf("Cache hit rate: %.1f%%\n", 
        float64(stats.Hits)/(stats.Hits+stats.Misses)*100)
}
```

## Examples 📖

The repository includes 7 complete examples showing different configurations:

| Example | Description | Dependencies | Difficulty |
|---------|-------------|--------------|------------|
| [01-session-only](examples/01-session-only/) | In-memory only, no persistence | None | ⭐ Beginner |
| [02-persistent-basic](examples/02-persistent-basic/) | PostgreSQL persistence | PostgreSQL | ⭐⭐ Easy |
| [03-hybrid-mode](examples/03-hybrid-mode/) | Redis + PostgreSQL | Redis, PostgreSQL | ⭐⭐⭐ Intermediate |
| [04-semantic-search](examples/04-semantic-search/) | Vector search with pgvector | PostgreSQL, OpenAI | ⭐⭐⭐ Intermediate |
| [05-auto-summarization](examples/05-auto-summarization/) | Automatic conversation summaries | PostgreSQL, OpenAI | ⭐⭐⭐ Intermediate |
| [06-event-streaming](examples/06-event-streaming/) | Redis Streams for event sourcing | Redis | ⭐⭐⭐⭐ Advanced |
| [07-agent-integration](examples/07-agent-integration/) | Complete AI agent with memory | All optional | ⭐⭐⭐⭐ Advanced |

Each example includes:
- Complete runnable code
- Detailed README with setup instructions
- Performance benchmarks and usage patterns
- Migration guides to more advanced configurations

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
