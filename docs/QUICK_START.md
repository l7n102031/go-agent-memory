# 🚀 Quick Start Guide

Get Go Agent Memory running in your project in 5 minutes.

## 1. Install

```bash
go get github.com/framehood/go-agent-memory
```

## 2. Choose Your Mode

### Option A: Zero Dependencies (Development)
Perfect for development, testing, or stateless applications.

```go
import memory "github.com/framehood/go-agent-memory"

mem, _ := memory.NewWithConfig(memory.Config{
    Mode: memory.SESSION_ONLY,
    MaxSessionMessages: 50, // Optional: default 50
})
defer mem.Close()
```

**✅ Pros:** No setup, instant start, microsecond performance  
**❌ Cons:** Data lost on restart, no semantic search

### Option B: Persistent (Production)
For applications needing conversation history and semantic search.

```bash
# Set up PostgreSQL with pgvector
export DATABASE_URL="postgresql://user:pass@host:5432/dbname"
export OPENAI_API_KEY="sk-..."  # Optional: for semantic search
```

```go
mem, _ := memory.NewWithConfig(memory.Config{
    Mode:                 memory.PERSISTENT,
    DatabaseURL:          os.Getenv("DATABASE_URL"),
    OpenAIKey:            os.Getenv("OPENAI_API_KEY"),
    EnableSemanticSearch: true,  // Optional
    EnableAutoSummarize:  true,  // Optional
})
defer mem.Close()
```

**✅ Pros:** Persistent, semantic search, summarization  
**❌ Cons:** Requires database, 20-50ms latency

### Option C: Hybrid (Best Performance)
Combines Redis cache with PostgreSQL persistence.

```bash
# Additional Redis setup
export REDIS_URL="localhost:6379"
```

```go
mem, _ := memory.NewWithConfig(memory.Config{
    Mode:        memory.HYBRID,
    DatabaseURL: os.Getenv("DATABASE_URL"),
    RedisAddr:   os.Getenv("REDIS_URL"),
    OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
    
    EnableSemanticSearch: true,
    EnableAutoSummarize:  true,
    MaxSessionMessages:   30,        // Redis cache size
    SessionTTL:          2 * time.Hour,
})
defer mem.Close()
```

**✅ Pros:** 2-5ms cache performance + persistence + all features  
**❌ Cons:** Requires Redis + PostgreSQL

## 3. Basic Usage

All modes support the same core API:

```go
ctx := context.Background()
sessionID := "user-123"

// Store messages
mem.AddMessage(ctx, memory.Message{
    ID:      "msg-1",
    Role:    "user",
    Content: "Hello!",
    Metadata: memory.Metadata{
        SessionID: sessionID,
    },
    Timestamp: time.Now(),
})

// Get recent messages
messages, _ := mem.GetRecentMessages(ctx, sessionID, 10)

// Search (persistent/hybrid modes only)
if results, err := mem.Search(ctx, "hello", 5, 0.7); err == nil {
    for _, result := range results {
        fmt.Printf("Found: %s (score: %.2f)\n", 
            result.Message.Content, result.Score)
    }
}

// Get statistics
stats, _ := mem.GetStats(ctx, sessionID)
fmt.Printf("Messages: %d, Tokens: %d\n", 
    stats.SessionMessages, stats.TotalTokens)
```

## 4. Integration with Your Agent

### Simple Integration

```go
package main

import (
    "context"
    "os"
    memory "github.com/framehood/go-agent-memory"
)

var mem memory.Memory

func init() {
    // Initialize based on environment
    if os.Getenv("DATABASE_URL") != "" {
        mem, _ = memory.NewWithConfig(memory.Config{
            Mode:        memory.HYBRID,
            DatabaseURL: os.Getenv("DATABASE_URL"),
            RedisAddr:   os.Getenv("REDIS_URL"),
            OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
            EnableSemanticSearch: true,
        })
    } else {
        // Fallback to session-only
        mem, _ = memory.NewWithConfig(memory.Config{
            Mode: memory.SESSION_ONLY,
        })
    }
}

func handleUserMessage(sessionID, userInput string) {
    ctx := context.Background()
    
    // Store user message
    if mem != nil {
        mem.AddMessage(ctx, memory.Message{
            Role:    "user",
            Content: userInput,
            Metadata: memory.Metadata{SessionID: sessionID},
        })
        
        // Get conversation context
        recent, _ := mem.GetRecentMessages(ctx, sessionID, 10)
        // Use 'recent' to build your prompt...
    }
    
    // ... your agent logic ...
    
    // Store assistant response
    if mem != nil {
        mem.AddMessage(ctx, memory.Message{
            Role:    "assistant", 
            Content: response,
            Metadata: memory.Metadata{SessionID: sessionID},
        })
    }
}
```

## 5. Environment Setup

### For Development (Session-Only)
```bash
# No environment variables needed!
go run main.go
```

### For Persistent Mode
```bash
# PostgreSQL (Supabase recommended)
export DATABASE_URL="postgresql://postgres:password@db.supabase.co:5432/postgres"

# Optional: OpenAI for semantic search
export OPENAI_API_KEY="sk-..."

go run main.go
```

### For Hybrid Mode
```bash
# PostgreSQL + Redis
export DATABASE_URL="postgresql://postgres:password@db.supabase.co:5432/postgres"
export REDIS_URL="localhost:6379"
export OPENAI_API_KEY="sk-..."

# Start Redis locally
redis-server

go run main.go
```

### Docker Setup (Testing)
```bash
# Start local PostgreSQL + Redis
cd deployment
docker-compose up -d

export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/agent_memory"
export REDIS_URL="localhost:6379"
```

## 6. Performance Expectations

| Mode | Setup Time | Read Latency | Write Latency | Dependencies |
|------|------------|--------------|---------------|--------------|
| **Session-Only** | Instant | ~1μs | ~1μs | None |
| **Persistent** | 2-5 min | 20-50ms | 20-30ms | PostgreSQL |
| **Hybrid** | 5-10 min | 2-5ms | 10ms | Redis + PostgreSQL |

## 7. Next Steps

### Detailed Examples
- **[01-session-only](../examples/01-session-only/)** - Zero dependencies example
- **[02-persistent-basic](../examples/02-persistent-basic/)** - PostgreSQL setup
- **[03-hybrid-mode](../examples/03-hybrid-mode/)** - Redis + PostgreSQL
- **[07-agent-integration](../examples/07-agent-integration/)** - Complete AI agent

### Advanced Features
- **[Semantic Search](../examples/04-semantic-search/)** - Vector embeddings
- **[Auto-Summarization](../examples/05-auto-summarization/)** - Token optimization
- **[Event Streaming](../examples/06-event-streaming/)** - Redis Streams

### Troubleshooting
- Memory not working? Check environment variables
- Slow performance? Try hybrid mode
- Need semantic search? Set `OPENAI_API_KEY`
- Want zero dependencies? Use `SESSION_ONLY` mode

## 🎯 That's It!

Your agent now has configurable memory. Start with session-only mode and upgrade as needed!
