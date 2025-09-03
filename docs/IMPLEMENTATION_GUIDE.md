# 🧠 Memory Implementation Guide

A comprehensive guide for integrating configurable memory into your AI agent with multiple modes and feature flags.

## Table of Contents
- [Memory Modes](#memory-modes)
- [Configuration Options](#configuration-options)
- [Implementation Examples](#implementation-examples)
- [Usage Patterns](#usage-patterns)
- [Best Practices](#best-practices)

---

## Memory Modes

The memory system supports multiple operational modes that can be configured based on your needs:

### 1. **Session-Only Mode** (No Persistence)
- Fast, ephemeral memory
- Perfect for stateless applications
- No database required
- Memory cleared on restart

### 2. **Persistent Mode** (Full Features)
- Session memory + long-term storage
- Semantic search capabilities
- Conversation history across restarts
- Requires PostgreSQL + optional Redis

### 3. **Hybrid Mode** (Balanced)
- Redis for sessions
- PostgreSQL for important memories
- Best performance/persistence balance

---

## Configuration Options

### Feature Flags

```go
type MemoryConfig struct {
    // Core Settings
    Mode           MemoryMode    // SESSION_ONLY, PERSISTENT, HYBRID
    
    // Feature Flags
    EnablePersistence    bool   // Store to database
    EnableSemanticSearch bool   // Generate embeddings for search
    EnableAutoSummarize  bool   // Auto-summarize long conversations
    EnableEventStream    bool   // Use Redis Streams for event sourcing
    EnableContextSwitch  bool   // Detect and handle topic changes
    
    // Behavior Settings
    MaxSessionMessages   int    // Messages to keep in session (default: 50)
    SummarizeAfter      int    // Summarize after N messages (default: 100)
    SessionTTL          time.Duration // Session expiry (default: 24h)
    
    // Storage Backends
    RedisAddr           string  // Optional Redis connection
    DatabaseURL         string  // PostgreSQL connection
    OpenAIKey          string  // For embeddings and summarization
}
```

---

## Implementation Examples

> 📁 **Note**: Complete runnable examples with detailed READMEs are available in the [`examples/`](../examples/) directory.

### Example 1: Session-Only Memory (Simplest)
📂 **Full Example**: [`examples/01-session-only/`](../examples/01-session-only/)

```go
// examples/session_only.go
package main

import (
    "context"
    "fmt"
    "os"
    
    memory "github.com/framehood/go-agent-memory"
)

func main() {
    // Session-only configuration - no persistence
    config := memory.Config{
        Mode:               memory.SESSION_ONLY,
        EnablePersistence:  false,
        EnableSemanticSearch: false,
        MaxSessionMessages: 20,
        // No database or Redis needed!
    }
    
    mem, err := memory.NewWithConfig(config)
    if err != nil {
        log.Fatal(err)
    }
    defer mem.Close()
    
    // Use memory for current session only
    sessionID := "session-123"
    
    // Add messages
    mem.AddMessage(context.Background(), memory.Message{
        ID:      "msg-1",
        Role:    "user",
        Content: "Hello!",
        Metadata: memory.Metadata{
            SessionID: sessionID,
        },
    })
    
    // Retrieve recent messages (from memory only)
    messages, _ := mem.GetRecentMessages(context.Background(), sessionID, 10)
    
    // No persistence - everything lost on restart
}
```

### Example 2: Persistent Memory with Auto-Summarization

```go
// examples/persistent_with_summary.go
package main

import (
    "context"
    "log"
    "os"
    "time"
    
    memory "github.com/framehood/go-agent-memory"
)

func main() {
    config := memory.Config{
        Mode:                memory.PERSISTENT,
        EnablePersistence:   true,
        EnableSemanticSearch: true,
        EnableAutoSummarize: true,  // Auto-summarization enabled
        
        // Summarization settings
        SummarizeAfter:     50,      // Summarize after 50 messages
        MaxSessionMessages: 100,     // Keep last 100 in quick access
        
        // Storage
        DatabaseURL: os.Getenv("DATABASE_URL"),
        OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
    }
    
    mem, err := memory.NewWithConfig(config)
    if err != nil {
        log.Fatal(err)
    }
    defer mem.Close()
    
    // Memory will automatically:
    // 1. Store all messages to PostgreSQL
    // 2. Generate embeddings for semantic search
    // 3. Auto-summarize after 50 messages
    
    sessionID := "persistent-session"
    
    // Simulate conversation
    for i := 0; i < 60; i++ {
        mem.AddMessage(context.Background(), memory.Message{
            ID:      fmt.Sprintf("msg-%d", i),
            Role:    "user",
            Content: fmt.Sprintf("Message %d", i),
            Metadata: memory.Metadata{
                SessionID: sessionID,
            },
        })
    }
    
    // After 50 messages, auto-summary was created
    summary, _ := mem.GetSummary(context.Background(), sessionID)
    fmt.Printf("Auto-generated summary: %s\n", summary)
}
```

### Example 3: Hybrid Mode with Custom Settings

```go
// examples/hybrid_custom.go
package main

import (
    "context"
    "time"
    
    memory "github.com/framehood/go-agent-memory"
)

func main() {
    config := memory.Config{
        Mode: memory.HYBRID,
        
        // Features
        EnablePersistence:    true,
        EnableSemanticSearch: true,
        EnableAutoSummarize:  false,  // Manual summarization only
        EnableEventStream:    true,   // Use Redis Streams
        EnableContextSwitch:  true,   // Detect topic changes
        
        // Behavior
        MaxSessionMessages:  30,
        SessionTTL:         12 * time.Hour,
        
        // Storage
        RedisAddr:   "localhost:6379",
        DatabaseURL: os.Getenv("DATABASE_URL"),
        OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
    }
    
    mem, err := memory.NewWithConfig(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // With event streaming enabled
    mem.OnEvent(func(event memory.Event) {
        log.Printf("Event: %s - %v", event.Type, event.Data)
    })
    
    // With context switching detection
    mem.OnContextSwitch(func(old, new string) {
        log.Printf("Context switched from %s to %s", old, new)
    })
}
```

### Example 4: Minimal Integration for Existing Agent

```go
// examples/agent_integration.go
package main

import (
    "context"
    "fmt"
    "os"
    
    memory "github.com/framehood/go-agent-memory"
    openai "github.com/openai/openai-go/v2"
)

type Agent struct {
    client *openai.Client
    memory memory.Memory
}

func NewAgent() *Agent {
    // Optional memory - works without it
    var mem memory.Memory
    
    if os.Getenv("ENABLE_MEMORY") == "true" {
        config := memory.Config{
            Mode: memory.AUTO, // Auto-detect based on available backends
            
            // Let the module decide based on environment
            DatabaseURL: os.Getenv("DATABASE_URL"),
            RedisAddr:   os.Getenv("REDIS_URL"),
            OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
        }
        
        mem, _ = memory.NewWithConfig(config)
    }
    
    return &Agent{
        client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
        memory: mem, // Can be nil
    }
}

func (a *Agent) Chat(sessionID, message string) (string, error) {
    ctx := context.Background()
    
    // Build context with memory if available
    var messages []openai.ChatCompletionMessageParam
    
    if a.memory != nil {
        // Retrieve context from memory
        history, _ := a.memory.GetRecentMessages(ctx, sessionID, 10)
        for _, msg := range history {
            messages = append(messages, openai.ChatCompletionMessageParam{
                Role:    msg.Role,
                Content: msg.Content,
            })
        }
    }
    
    // Add current message
    messages = append(messages, openai.ChatCompletionMessageParam{
        Role:    "user",
        Content: message,
    })
    
    // Store in memory if available
    if a.memory != nil {
        a.memory.AddMessage(ctx, memory.Message{
            Role:    "user",
            Content: message,
            Metadata: memory.Metadata{
                SessionID: sessionID,
            },
        })
    }
    
    // Get response from OpenAI
    resp, err := a.client.CreateChatCompletion(ctx, messages)
    if err != nil {
        return "", err
    }
    
    response := resp.Choices[0].Message.Content
    
    // Store response in memory if available
    if a.memory != nil {
        a.memory.AddMessage(ctx, memory.Message{
            Role:    "assistant",
            Content: response,
            Metadata: memory.Metadata{
                SessionID: sessionID,
                Model:     resp.Model,
            },
        })
    }
    
    return response, nil
}
```

---

## Usage Patterns

### Pattern 1: Development Mode (No External Dependencies)

```go
config := memory.Config{
    Mode: memory.SESSION_ONLY,
    MaxSessionMessages: 10,
}
// Works without any database or Redis!
```

### Pattern 2: Production with All Features

```go
config := memory.Config{
    Mode:                memory.PERSISTENT,
    EnablePersistence:   true,
    EnableSemanticSearch: true,
    EnableAutoSummarize: true,
    EnableEventStream:   true,
    
    DatabaseURL: os.Getenv("DATABASE_URL"),
    RedisAddr:   os.Getenv("REDIS_URL"),
    OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
}
```

### Pattern 3: Custom Feature Selection

```go
config := memory.Config{
    Mode: memory.CUSTOM,
    
    // Pick only what you need
    EnablePersistence:    true,   // Yes, save to DB
    EnableSemanticSearch: false,  // No, don't need search
    EnableAutoSummarize:  false,  // No, manual control
    EnableEventStream:    true,   // Yes, for debugging
}
```

### Pattern 4: Environment-Based Configuration

```go
func getMemoryConfig() memory.Config {
    env := os.Getenv("ENV")
    
    switch env {
    case "development":
        return memory.Config{
            Mode: memory.SESSION_ONLY,
            MaxSessionMessages: 20,
        }
    
    case "staging":
        return memory.Config{
            Mode: memory.PERSISTENT,
            EnablePersistence: true,
            DatabaseURL: os.Getenv("DATABASE_URL"),
        }
    
    case "production":
        return memory.Config{
            Mode: memory.HYBRID,
            EnablePersistence:    true,
            EnableSemanticSearch: true,
            EnableAutoSummarize:  true,
            DatabaseURL: os.Getenv("DATABASE_URL"),
            RedisAddr:   os.Getenv("REDIS_URL"),
        }
    
    default:
        return memory.Config{
            Mode: memory.SESSION_ONLY,
        }
    }
}
```

---

## Best Practices

### 1. Start Simple, Add Features Gradually

```go
// Phase 1: Session-only
config := memory.Config{Mode: memory.SESSION_ONLY}

// Phase 2: Add persistence
config.EnablePersistence = true
config.DatabaseURL = "..."

// Phase 3: Add search
config.EnableSemanticSearch = true

// Phase 4: Add optimization
config.EnableAutoSummarize = true
```

### 2. Handle Memory Gracefully

```go
// Memory should never break your agent
if mem != nil {
    mem.AddMessage(ctx, msg)
} 
// Continue even if memory is nil
```

### 3. Use Appropriate Modes for Different Environments

```go
// Development: SESSION_ONLY (no dependencies)
// Staging: PERSISTENT (test full features)
// Production: HYBRID (optimal performance)
```

### 4. Monitor Memory Usage

```go
// Periodically check stats
stats, _ := mem.GetStats(ctx, sessionID)
if stats.TotalMessages > 1000 {
    // Trigger cleanup or summarization
}
```

### 5. Configure Based on Use Case

| Use Case | Recommended Mode | Features |
|----------|-----------------|----------|
| Chatbot | HYBRID | Persistence, Auto-summary |
| Code Assistant | PERSISTENT | Semantic search, Event stream |
| Customer Service | HYBRID | Persistence, Context switch |
| Development Tool | SESSION_ONLY | Minimal features |
| Research Assistant | PERSISTENT | All features enabled |

---

## Environment Variables

```bash
# Required for persistent modes
DATABASE_URL=postgresql://user:pass@host:5432/db
OPENAI_API_KEY=sk-...

# Optional for hybrid mode
REDIS_URL=localhost:6379

# Feature flags (optional)
MEMORY_MODE=hybrid              # session_only, persistent, hybrid
MEMORY_ENABLE_SEARCH=true       # Enable semantic search
MEMORY_ENABLE_SUMMARY=true      # Enable auto-summarization
MEMORY_ENABLE_EVENTS=true       # Enable event streaming
MEMORY_MAX_MESSAGES=50          # Session message limit
MEMORY_SUMMARIZE_AFTER=100      # Summarize after N messages
```

---

## Testing Different Configurations

```bash
# Test session-only mode
MEMORY_MODE=session_only go run examples/session_only.go

# Test persistent mode
MEMORY_MODE=persistent DATABASE_URL=... go run examples/persistent.go

# Test hybrid mode
MEMORY_MODE=hybrid REDIS_URL=... DATABASE_URL=... go run examples/hybrid.go
```

---

## Migration Path

### From No Memory → Session Memory
```go
// Before: No memory
messages := []Message{}

// After: Session memory
mem, _ := memory.NewWithConfig(memory.Config{
    Mode: memory.SESSION_ONLY,
})
```

### From Session → Persistent
```go
// Just add database
config.EnablePersistence = true
config.DatabaseURL = "..."
```

### From Persistent → Hybrid
```go
// Add Redis for speed
config.Mode = memory.HYBRID
config.RedisAddr = "..."
```

---

## Debugging Memory

```go
// Enable debug mode
config.Debug = true

// Listen to events
mem.OnEvent(func(e memory.Event) {
    log.Printf("[MEMORY] %s: %v", e.Type, e.Data)
})

// Replay events (if event stream enabled)
events, _ := mem.ReplayEvents(ctx, sessionID, time.Hour)
for _, event := range events {
    fmt.Printf("%s: %s\n", event.Timestamp, event.Type)
}
```

This guide provides flexibility to use memory in any configuration that suits your needs!
