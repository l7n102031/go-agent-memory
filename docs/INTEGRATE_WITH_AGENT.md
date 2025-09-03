# Quick Integration Guide for Your Agent

## 1. Add Memory Package to Your Agent

First, update your agent's `go.mod`:

```bash
cd /Users/kirillshidenko/framehood/agent
go get github.com/framehood/go-agent-memory
```

## 2. Update Your main.go

Add these changes to your existing `/Users/kirillshidenko/framehood/agent/main.go`:

```go
// Add import at the top
import (
    // ... existing imports ...
    memory "github.com/framehood/go-agent-memory"
)

// Add after your configuration constants
var mem memory.Memory // Optional memory instance

// Add this function before main()
func initMemory() {
    // Option 1: Session-only mode (no dependencies)
    if os.Getenv("DATABASE_URL") == "" {
        mem, _ = memory.NewWithConfig(memory.Config{
            Mode: memory.SESSION_ONLY,
            MaxSessionMessages: 30,
        })
        return
    }
    
    // Option 2: Full persistent mode
    var err error
    mem, err = memory.NewWithConfig(memory.Config{
        Mode:        memory.HYBRID, // or PERSISTENT
        DatabaseURL: os.Getenv("DATABASE_URL"),
        RedisAddr:   os.Getenv("REDIS_URL"),
        OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
        
        // Feature flags
        EnableSemanticSearch: true,
        EnableAutoSummarize:  true,
        
        // Settings
        MaxSessionMessages: 30,
        SessionTTL:        2 * time.Hour,
    })
    
    if err != nil {
        fmt.Printf("Memory disabled: %v\n", err)
        mem = nil
    }
}

// In your main() function, add at the beginning:
func main() {
    initMemory()
    if mem != nil {
        defer mem.Close()
    }
    
    // ... rest of your existing code ...
}

// In your conversation loop, after receiving user input:
if mem != nil {
    // Store user message
    mem.AddMessage(context.Background(), memory.Message{
        ID:      fmt.Sprintf("msg-%d", time.Now().Unix()),
        Role:    "user",
        Content: input,
        Metadata: memory.Metadata{
            SessionID: sessionID, // You'll need to track this
        },
        Timestamp: time.Now(),
    })
    
    // Get context from memory
    recent, _ := mem.GetRecentMessages(context.Background(), sessionID, 10)
    
    // Add to your messages array
    for _, msg := range recent[max(0, len(recent)-5):] { // Last 5 messages
        messages = append(messages, openai.ChatCompletionMessageParamUnion{
            openai.UserMessage(params)
            // Convert based on msg.Role
        })
    }
}

// After getting assistant response:
if mem != nil {
    mem.AddMessage(context.Background(), memory.Message{
        ID:      fmt.Sprintf("msg-%d", time.Now().Unix()),
        Role:    "assistant",
        Content: assistantMessage,
        Metadata: memory.Metadata{
            SessionID:  sessionID,
            Model:     MODEL,
            Temperature: TEMPERATURE,
        },
        Timestamp: time.Now(),
    })
}
```

## 3. Environment Variables

Create a `.env` file in your agent directory:

```bash
# Your existing OpenAI key
OPENAI_API_KEY=sk-...

# Your Supabase connection
DATABASE_URL=postgresql://postgres:password@db.yourdomain.supabase.co:5432/postgres

# Optional Redis (for faster access)
REDIS_URL=localhost:6379
```

## 4. Run Your Agent with Memory

```bash
# Load environment variables and run
source .env && go run main.go
```

## Features You Get

### All Modes:
1. **Conversation History**: Automatically stores all conversations
2. **Context Retrieval**: Recent messages are loaded automatically  
3. **Session Management**: Track different conversation sessions
4. **Statistics**: Get memory usage stats

### Persistent/Hybrid Modes Only:
5. **Semantic Search**: Find relevant past conversations:
   ```go
   if mem != nil {
       results, _ := mem.Search(ctx, "weather in Tokyo", 3, 0.7)
       // Add relevant context to your prompt
   }
   ```

6. **Auto-Summarization**: Long conversations are automatically summarized
7. **Conversation Summaries**: Get structured summaries:
   ```go
   if mem != nil {
       summary, _ := mem.GetSummary(ctx, sessionID)
       fmt.Printf("Summary: %s\n", summary.Content)
   }
   ```

### Hybrid Mode Only:
8. **Cache Management**: Clear and monitor cache performance

## Optional: Add Memory Commands

You can add special commands to your agent:

```go
// In your input processing
switch input {
case "/memory stats":
    if mem != nil {
        stats, _ := mem.GetStats(context.Background(), sessionID)
        fmt.Printf("Messages: %d, Tokens: %d\n", 
            stats.SessionMessages, stats.TotalTokens)
    }
case "/memory search":
    fmt.Print("Search query: ")
    query := getUserInput()
    if mem != nil {
        results, _ := mem.Search(context.Background(), query, 5, 0.7)
        for _, r := range results {
            fmt.Printf("[%.2f] %s\n", r.Score, r.Message.Content)
        }
    }
case "/memory clear":
    if mem != nil {
        mem.ClearSession(context.Background(), sessionID)
        fmt.Println("Session memory cleared")
    }
}
```

## That's It! 🎉

Your agent now has:
- ✅ Persistent conversation memory
- ✅ Fast session caching (if Redis is configured)
- ✅ Semantic search across all conversations
- ✅ Automatic embeddings generation
- ✅ No changes needed if DATABASE_URL is not set (memory is optional)

The memory module won't affect your agent if not configured - it's completely optional!
