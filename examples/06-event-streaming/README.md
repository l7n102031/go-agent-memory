# Event Streaming Memory Example

Redis Streams for event sourcing, audit trails, and debugging AI agent behavior.

## 🎯 What This Example Shows

- Event sourcing with Redis Streams
- Real-time event processing
- Consumer groups for distributed processing
- Event replay and time travel
- Audit trail generation
- Analytics and debugging

## 📋 Features

- ✅ **Event sourcing** - Immutable event log
- ✅ **Real-time streaming** - Process events as they happen
- ✅ **Consumer groups** - Distributed processing
- ✅ **Event replay** - Time travel through history
- ✅ **Audit trails** - Complete activity log
- ✅ **Analytics** - Event statistics and patterns

## 🚀 Quick Start

### Prerequisites

1. Redis server (6.0+ for Streams)
2. Environment variables:

```bash
# Required
export REDIS_URL="localhost:6379"

# Optional
export REDIS_PASSWORD=""
```

### Run the Example

```bash
go run main.go
```

## 💻 Architecture Overview

```
    Event Occurs
         ↓
    Publish to Stream
         ↓
    ┌────────────┐
    │Redis Stream│
    └─────┬──────┘
          ↓
    ┌─────┴──────┬──────────┬─────────┐
    ↓            ↓          ↓         ↓
Analytics    Audit Log   Debugger   Replay
Consumer     Consumer    Consumer   System
```

## 📡 Event Types

```go
const (
    EventMessageAdded    = "message.added"
    EventSessionStarted  = "session.started"
    EventSessionEnded    = "session.ended"
    EventSummaryCreated  = "summary.created"
    EventSearchPerformed = "search.performed"
    EventMemoryCleared   = "memory.cleared"
    EventErrorOccurred   = "error.occurred"
)
```

## 🔄 Event Flow

### Publishing Events
```go
// Every action generates an event
event := Event{
    Type:      EventMessageAdded,
    SessionID: "session-123",
    UserID:    "user-456",
    Timestamp: time.Now(),
    Data: map[string]interface{}{
        "message_id": "msg-789",
        "role":       "user",
        "tokens":     150,
    },
}

// Publish to stream
redis.XAdd(ctx, &redis.XAddArgs{
    Stream: "memory:events",
    Values: eventToMap(event),
})
```

### Consuming Events
```go
// Real-time consumption
for {
    events := redis.XRead(ctx, &redis.XReadArgs{
        Streams: []string{"memory:events", lastID},
        Block:   1 * time.Second,
    })
    
    for _, event := range events {
        processEvent(event)
    }
}
```

## 👥 Consumer Groups

### Multiple Consumers
```
Stream: memory:events
    │
    ├─→ analytics-group
    │     ├─→ analytics-consumer-1
    │     └─→ analytics-consumer-2
    │
    ├─→ audit-group
    │     └─→ audit-consumer-1
    │
    └─→ debug-group
          └─→ debug-consumer-1
```

### Group Benefits
- **Parallel processing** - Multiple consumers per group
- **At-least-once delivery** - Acknowledgment required
- **Load balancing** - Automatic work distribution
- **Fault tolerance** - Reassign pending messages

## 🕐 Event Replay

### Time-Based Replay
```go
// Replay events from last hour
oneHourAgo := time.Now().Add(-1 * time.Hour)
events := redis.XRange(ctx, "memory:events", 
    fmt.Sprintf("%d", oneHourAgo.UnixMilli()), "+")

// Process historical events
for _, event := range events {
    replayEvent(event)
}
```

### Session Reconstruction
```go
// Rebuild complete session from events
func reconstructSession(sessionID string) Session {
    events := redis.XRange(ctx, "memory:events", "-", "+")
    
    session := Session{}
    for _, event := range events {
        if event.SessionID == sessionID {
            applyEventToSession(&session, event)
        }
    }
    return session
}
```

## 📊 Event Analytics

### Real-time Metrics
```go
type EventMetrics struct {
    TotalEvents        int64
    EventsPerSecond    float64
    AverageLatency     time.Duration
    ErrorRate          float64
    ActiveSessions     int
    TopEventTypes      map[string]int
}
```

### Pattern Detection
- Identify usage patterns
- Detect anomalies
- Track performance trends
- Monitor error rates

## 🔍 Debugging Features

### Event Inspector
```go
// Find all events for a user
func getUserEvents(userID string) []Event {
    // Query stream for user's events
    // Useful for debugging user issues
}

// Find error events
func getErrorEvents(timeRange TimeRange) []Event {
    // Filter for error events
    // Helps identify problems
}
```

### Session Timeline
```
Session: abc-123
10:00:00 session.started
10:00:05 message.added (user)
10:00:07 message.added (assistant)
10:00:45 search.performed
10:01:20 summary.created
10:15:00 session.ended
```

## 🎯 Use Cases

### 1. Audit Trail
- Complete activity log
- Compliance requirements
- Security monitoring

### 2. Debugging
- Reproduce issues
- Trace execution flow
- Performance analysis

### 3. Analytics
- Usage patterns
- User behavior
- System metrics

### 4. Recovery
- Rebuild state from events
- Disaster recovery
- Data migration

## 🔧 Configuration

```go
type EventStreamConfig struct {
    // Stream settings
    StreamName      string
    MaxLength       int64  // Limit stream size
    ApproxMaxLength bool   // Use approximate trimming
    
    // Consumer settings
    ConsumerGroup   string
    ConsumerName    string
    BatchSize       int
    BlockTimeout    time.Duration
    
    // Retention
    RetentionPeriod time.Duration
    CompactOldEvents bool
}
```

## 📈 Performance Considerations

### Stream Size Management
```go
// Trim old events
redis.XTrimMaxLen(ctx, "memory:events", 10000)

// Time-based trimming
redis.XTrimMinID(ctx, "memory:events", oneWeekAgo)
```

### Throughput
| Events/Second | CPU Usage | Memory | Recommendation |
|--------------|-----------|---------|----------------|
| < 100 | Low | < 10MB | Single consumer |
| 100-1000 | Medium | < 100MB | 2-3 consumers |
| 1000-10000 | High | < 1GB | Consumer group |
| > 10000 | Very High | > 1GB | Multiple Redis instances |

## 🚨 Common Patterns

### Event Sourcing Pattern
```go
// State = fold(events)
currentState := initialState
for _, event := range events {
    currentState = applyEvent(currentState, event)
}
```

### CQRS Pattern
```go
// Commands write events
publishEvent(CommandExecuted{...})

// Queries read projections
projection := readProjection()
```

## 💡 Best Practices

1. **Immutable events** - Never modify, only append
2. **Event versioning** - Handle schema evolution
3. **Idempotency** - Handle duplicate events
4. **Compression** - Compact old events
5. **Monitoring** - Track lag and throughput

## 📚 Next Steps

- [07-agent-integration](../07-agent-integration/) - Complete system with events

## 📄 Full Code

See [main.go](./main.go) for the complete implementation.
