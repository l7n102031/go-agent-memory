# Session-Only Memory Example

The simplest memory configuration with no external dependencies. Perfect for development, testing, or stateless applications.

## 🎯 What This Example Shows

- In-memory session management
- No database or Redis required
- Fast message retrieval
- Automatic memory cleanup
- Graceful handling of memory limits

## 📋 Features

- ✅ **Zero dependencies** - Works out of the box
- ✅ **Fast access** - Nanosecond retrieval times
- ✅ **Simple API** - Easy to understand and use
- ❌ **No persistence** - Data lost on restart
- ❌ **No search** - Only chronological retrieval
- ❌ **Single instance** - Not distributed

## 🚀 Quick Start

```bash
# No setup required! Just run:
go run main.go
```

## 💻 Code Overview

```go
// Configuration for session-only mode
config := memory.Config{
    Mode:               memory.SESSION_ONLY,
    MaxSessionMessages: 20,  // Keep last 20 messages
    // No database or Redis configuration needed!
}

// Initialize memory
mem, _ := memory.NewWithConfig(config)

// Use it immediately
mem.AddMessage(ctx, message)
messages, _ := mem.GetRecentMessages(ctx, sessionID, 10)
```

## 📊 Memory Behavior

### Storage Limits
- Default: Last 50 messages per session
- Configurable via `MaxSessionMessages`
- Older messages automatically removed

### Performance
- Write: < 1 microsecond
- Read: < 1 microsecond
- No network calls
- No disk I/O

## 🔄 Lifecycle

```
Application Start
    ↓
Initialize Memory (instant)
    ↓
Store Messages (in-memory map)
    ↓
Retrieve as Needed (direct access)
    ↓
Application Stop → All Data Lost
```

## 📝 Use Cases

### ✅ Good For:
- Development and testing
- Stateless chat applications
- Temporary conversation context
- Prototyping
- Low-latency requirements

### ❌ Not Good For:
- Production systems needing persistence
- Multi-instance deployments
- Long conversation history
- Search capabilities
- Audit requirements

## 🧪 Testing the Example

Run the interactive demo:

```bash
go run main.go

# Output:
# Memory initialized (session-only mode)
# Adding 5 test messages...
# 
# Recent messages (last 3):
# - [user] Message 3
# - [user] Message 4
# - [user] Message 5
# 
# Session stats:
# Total messages: 5
# Memory usage: ~1KB
```

## 🔧 Configuration Options

```go
type SessionOnlyConfig struct {
    MaxSessionMessages   int           // Max messages per session (default: 50)
    MaxSessions         int           // Max concurrent sessions (default: 1000)
    SessionTimeout      time.Duration // Auto-cleanup timeout (default: 1 hour)
    EnableStats         bool          // Track usage statistics (default: false)
}
```

## 💡 Tips

1. **Memory Management**: Set reasonable limits to prevent memory exhaustion
2. **Session Cleanup**: Implement periodic cleanup for inactive sessions
3. **Monitoring**: Track memory usage if running long-term
4. **Upgrade Path**: Easy to upgrade to persistent mode later

## 📈 Scaling Considerations

| Sessions | Memory Usage | Performance |
|----------|-------------|-------------|
| 100 | ~1 MB | Excellent |
| 1,000 | ~10 MB | Excellent |
| 10,000 | ~100 MB | Good |
| 100,000 | ~1 GB | Monitor closely |

## 🚀 Next Steps

Ready for persistence? Check out:
- [02-persistent-basic](../02-persistent-basic/) - Add database storage
- [03-hybrid-mode](../03-hybrid-mode/) - Add Redis caching
- [07-agent-integration](../07-agent-integration/) - Complete agent example

## 📄 Full Code

See [main.go](./main.go) for the complete implementation.
