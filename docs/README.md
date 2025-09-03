# 📚 Go Agent Memory Documentation

Complete documentation for the Go Agent Memory module.

## 📖 Documentation Structure

### **🚀 [Quick Start Guide](QUICK_START.md)**
- Zero to running in 5 minutes
- All memory modes with examples
- Environment setup and integration
- Performance expectations

### **🏗️ [Architecture Guide](ARCHITECTURE.md)**  
- System design and data flow
- Component overview and implementation details
- Database schema and performance characteristics

---

## 📋 Quick Reference

### Memory Modes
```go
memory.SESSION_ONLY  // In-memory, zero dependencies
memory.PERSISTENT    // PostgreSQL + pgvector
memory.HYBRID        // Redis + PostgreSQL (best performance)
```

### Basic Usage
```go
// Zero dependencies
mem, _ := memory.NewWithConfig(memory.Config{
    Mode: memory.SESSION_ONLY,
})

// Full features  
mem, _ := memory.NewWithConfig(memory.Config{
    Mode:        memory.HYBRID,
    DatabaseURL: "postgresql://...",
    RedisAddr:   "localhost:6379",
    OpenAIKey:   "sk-...",
    EnableSemanticSearch: true,
})
```

## 🎯 Examples by Use Case

| Use Case | Example Directory | Description |
|----------|------------------|-------------|
| **Development/Testing** | [01-session-only](../examples/01-session-only/) | Zero dependencies |
| **Simple Persistence** | [02-persistent-basic](../examples/02-persistent-basic/) | PostgreSQL only |
| **Production Ready** | [03-hybrid-mode](../examples/03-hybrid-mode/) | Redis + PostgreSQL |
| **AI-Powered Search** | [04-semantic-search](../examples/04-semantic-search/) | Vector embeddings |
| **Cost Optimization** | [05-auto-summarization](../examples/05-auto-summarization/) | Token compression |
| **Event Tracking** | [06-event-streaming](../examples/06-event-streaming/) | Redis Streams |
| **Complete Agent** | [07-agent-integration](../examples/07-agent-integration/) | Full integration |

## 🔗 External Links

- **[Main README](../README.md)** - Project overview and installation
- **[Examples](../examples/)** - Complete runnable examples
- **[GitHub Repository](https://github.com/framehood/go-agent-memory)** - Source code
