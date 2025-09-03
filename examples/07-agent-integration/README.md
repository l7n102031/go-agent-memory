# Complete Agent Integration Example

A production-ready AI agent with configurable memory system, demonstrating all features working together.

## 🎯 What This Example Shows

- Complete OpenAI agent with memory
- Configurable memory modes via environment
- Graceful degradation without memory
- Context management for conversations
- Tool calling with memory
- Production best practices

## 📋 Features

- ✅ **Full Integration** - Agent + Memory working together
- ✅ **Configurable** - Switch modes via environment
- ✅ **Resilient** - Works even if memory fails
- ✅ **Context-Aware** - Smart context window management
- ✅ **Tool Memory** - Remembers tool interactions
- ✅ **Production Ready** - Error handling, logging, monitoring

## 🚀 Quick Start

### Option 1: No Memory (Development)
```bash
go run main.go
```

### Option 2: Session Memory Only
```bash
MEMORY_MODE=session_only go run main.go
```

### Option 3: Full Persistent Memory
```bash
DATABASE_URL="postgresql://..." \
OPENAI_API_KEY="sk-..." \
MEMORY_MODE=persistent \
go run main.go
```

### Option 4: Hybrid Mode (Recommended for Production)
```bash
DATABASE_URL="postgresql://..." \
REDIS_URL="localhost:6379" \
OPENAI_API_KEY="sk-..." \
MEMORY_MODE=hybrid \
go run main.go
```

## 💻 Architecture Overview

```
User Input
    ↓
Session Manager (get/create session)
    ↓
Memory Retrieval (if enabled)
    ├── Recent Messages (Redis/Memory)
    └── Semantic Search (PostgreSQL)
    ↓
Context Builder
    ├── System Prompt
    ├── Conversation History
    └── Relevant Context
    ↓
OpenAI API Call
    ├── Chat Completion
    └── Tool Calls (if needed)
    ↓
Memory Storage (if enabled)
    ├── User Message
    ├── Assistant Response
    └── Tool Results
    ↓
Response to User
```

## 🛠️ Agent Configuration

```go
type AgentConfig struct {
    // OpenAI Settings
    Model           string  // "gpt-4", "gpt-3.5-turbo", etc.
    Temperature     float64 // 0.0 to 2.0
    MaxTokens       int     // Response limit
    
    // Memory Settings
    MemoryMode      MemoryMode // NONE, SESSION_ONLY, PERSISTENT, HYBRID
    MemoryConfig    memory.Config
    
    // Context Management
    MaxContextTokens int    // Total context window
    RecentMessages   int    // Recent messages to include
    RelevantContext  int    // Semantic search results
    
    // Features
    EnableTools     bool    // Tool calling support
    EnableStreaming bool    // Stream responses
}
```

## 📝 Code Structure

```go
// Main agent structure
type Agent struct {
    client   *openai.Client
    memory   memory.Memory  // Can be nil
    config   AgentConfig
    sessions map[string]*Session
}

// Core methods
func (a *Agent) Chat(sessionID, message string) (string, error)
func (a *Agent) ChatWithTools(sessionID, message string) (string, error)
func (a *Agent) StreamChat(sessionID, message string) (<-chan string, error)
func (a *Agent) GetContext(sessionID string) ([]Message, error)
func (a *Agent) ClearSession(sessionID string) error
```

## 🔄 Memory Integration Points

### 1. Initialization
```go
func NewAgent(config AgentConfig) *Agent {
    // Initialize OpenAI client
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
    
    // Initialize memory (optional)
    var mem memory.Memory
    if config.MemoryMode != NONE {
        mem, err := initializeMemory(config)
        // Continue even if memory fails
    }
    
    return &Agent{
        client: client,
        memory: mem,
        config: config,
    }
}
```

### 2. Context Retrieval
```go
func (a *Agent) buildContext(sessionID, currentMessage string) []Message {
    var messages []Message
    
    // System prompt
    messages = append(messages, systemPrompt)
    
    if a.memory != nil {
        // Get recent conversation
        recent, _ := a.memory.GetRecentMessages(ctx, sessionID, 10)
        messages = append(messages, recent...)
        
        // Get relevant context
        if a.config.MemoryConfig.EnableSemanticSearch {
            relevant, _ := a.memory.Search(ctx, currentMessage, 5, 0.75)
            messages = appendRelevant(messages, relevant)
        }
    }
    
    // Add current message
    messages = append(messages, currentMessage)
    
    // Manage token budget
    return trimToTokenLimit(messages, a.config.MaxContextTokens)
}
```

### 3. Message Storage
```go
func (a *Agent) storeExchange(sessionID string, userMsg, assistantMsg string) {
    if a.memory == nil {
        return // Memory is optional
    }
    
    // Store user message
    a.memory.AddMessage(ctx, memory.Message{
        Role:    "user",
        Content: userMsg,
        Metadata: memory.Metadata{
            SessionID: sessionID,
        },
    })
    
    // Store assistant response
    a.memory.AddMessage(ctx, memory.Message{
        Role:    "assistant",
        Content: assistantMsg,
        Metadata: memory.Metadata{
            SessionID: sessionID,
            Model:     a.config.Model,
        },
    })
}
```

## 🎮 Interactive Features

### Chat Interface
```bash
$ go run main.go

🤖 AI Agent with Memory
========================
Memory Mode: HYBRID
Model: gpt-4
Session: auto-generated-id

You: Hello! Can you help me with Go?
Assistant: Of course! I'd be happy to help with Go programming...

You: /memory stats
Memory: 2 messages, 150 tokens used

You: /memory search "error handling"
Found 3 relevant messages...

You: /clear
Session cleared.
```

### Available Commands
- `/memory stats` - Show memory statistics
- `/memory search <query>` - Search conversation history
- `/session new` - Start new session
- `/session list` - List active sessions
- `/clear` - Clear current session
- `/help` - Show available commands
- `/exit` - Exit the agent

## 📊 Monitoring & Metrics

```go
// Track metrics
type AgentMetrics struct {
    TotalMessages    int64
    TotalTokens      int64
    AverageLatency   time.Duration
    MemoryHits       int64
    MemoryMisses     int64
    ErrorCount       int64
}

// Export metrics (Prometheus example)
func (a *Agent) ExportMetrics() {
    prometheus.CounterValue("agent_messages_total", a.metrics.TotalMessages)
    prometheus.GaugeValue("agent_tokens_used", a.metrics.TotalTokens)
    prometheus.HistogramValue("agent_response_time", a.metrics.AverageLatency)
}
```

## 🚨 Error Handling

```go
// Graceful degradation example
func (a *Agent) Chat(sessionID, message string) (string, error) {
    // Try to get context from memory
    context := a.getContextSafely(sessionID)
    
    // Continue even if memory fails
    if context == nil {
        log.Warn("Memory unavailable, using minimal context")
        context = []Message{{Role: "user", Content: message}}
    }
    
    // Make API call with available context
    response, err := a.callOpenAI(context)
    
    // Try to store, but don't fail if memory is down
    a.storeSafely(sessionID, message, response)
    
    return response, err
}
```

## 🌍 Environment Variables

```bash
# Required
OPENAI_API_KEY=sk-...

# Memory Configuration
MEMORY_MODE=hybrid              # none, session_only, persistent, hybrid
DATABASE_URL=postgresql://...   # For persistent/hybrid modes
REDIS_URL=localhost:6379        # For hybrid mode

# Agent Settings
AGENT_MODEL=gpt-4
AGENT_TEMPERATURE=0.7
AGENT_MAX_TOKENS=2000

# Feature Flags
ENABLE_STREAMING=true
ENABLE_TOOLS=true
ENABLE_SEMANTIC_SEARCH=true
ENABLE_AUTO_SUMMARY=true

# Performance
MAX_CONTEXT_TOKENS=8000
RECENT_MESSAGES=20
RELEVANT_CONTEXT=5
```

## 📈 Performance Optimization

### Token Management
```go
// Smart context window management
func optimizeContext(messages []Message, limit int) []Message {
    tokens := 0
    optimized := []Message{}
    
    // Always include system prompt
    optimized = append(optimized, systemPrompt)
    tokens += countTokens(systemPrompt)
    
    // Add recent messages (priority 1)
    for i := len(messages) - 1; i >= 0 && tokens < limit*0.7; i-- {
        optimized = append([]Message{messages[i]}, optimized...)
        tokens += countTokens(messages[i])
    }
    
    // Add relevant context (priority 2)
    // ... semantic search results
    
    return optimized
}
```

### Caching Strategy
- Cache embeddings for repeated content
- Cache summaries for long sessions
- Cache tool results when appropriate

## 🔒 Production Considerations

1. **Session Management**: Implement proper session lifecycle
2. **Rate Limiting**: Protect against abuse
3. **Monitoring**: Track usage and performance
4. **Security**: Validate and sanitize inputs
5. **Scaling**: Use connection pooling, implement sharding
6. **Backup**: Regular memory backups
7. **Privacy**: Implement data retention policies

## 🧪 Testing

```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run with different modes
./test_all_modes.sh
```

## 📚 Next Steps

1. Deploy to production with proper monitoring
2. Implement custom tools for your use case
3. Fine-tune memory configuration for optimal performance
4. Add authentication and authorization
5. Implement data retention policies

## 📄 Full Code

- [main.go](./main.go) - Complete agent implementation
- [memory_integration.go](./memory_integration.go) - Memory integration logic
- [context_manager.go](./context_manager.go) - Context optimization
- [tools.go](./tools.go) - Tool calling with memory
- [config.yaml](./config.yaml) - Configuration example
