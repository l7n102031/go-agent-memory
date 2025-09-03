# Hybrid Mode Memory Example

The best of both worlds: Redis for blazing-fast cache + PostgreSQL for reliable persistence.

## 🎯 What This Example Shows

- Redis caching for recent messages
- PostgreSQL for long-term storage
- Automatic cache warming
- Failover handling
- Performance optimization

## 📋 Features

- ✅ **Ultra-fast reads** - Redis cache (1-2ms)
- ✅ **Persistent storage** - PostgreSQL backup
- ✅ **Auto-failover** - Works if Redis is down
- ✅ **Smart caching** - TTL and size limits
- ✅ **Write-through** - Async persistence
- ✅ **Production ready** - Battle-tested pattern

## 🚀 Quick Start

### Prerequisites

1. PostgreSQL database
2. Redis server
3. Environment variables:

```bash
# Required
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
export REDIS_URL="localhost:6379"

# Optional
export REDIS_PASSWORD=""
export OPENAI_API_KEY="sk-..."
```

### Quick Setup with Docker

```bash
# Start both services
cd ../../deployment
docker-compose up -d

# Set environment
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/agent_memory"
export REDIS_URL="localhost:6379"

# Run example
cd ../examples/03-hybrid-mode
go run main.go
```

## 💻 Architecture Overview

```
        User Request
             ↓
    ┌─────────────────┐
    │  Memory Layer   │
    └────────┬────────┘
             ↓
       Check Redis ←──── 1-2ms
             ↓
         Found?
        ↙      ↘
      Yes       No
       ↓         ↓
    Return    Check PostgreSQL ←── 20-50ms
              ↓     ↓
           Store   Return
          in Redis
```

## 🔄 Data Flow

### Write Path (Optimized)
```go
AddMessage()
    ↓
Write to Redis (fast) ──→ Return immediately
    ↓
Async write to PostgreSQL (background)
```

### Read Path (Cached)
```go
GetRecentMessages()
    ↓
Check Redis cache ──→ HIT (1-2ms) ──→ Return
    ↓ MISS
Query PostgreSQL (20-50ms)
    ↓
Populate Redis cache
    ↓
Return messages
```

## 📊 Performance Comparison

| Operation | Session-Only | Persistent | **Hybrid** |
|-----------|-------------|------------|------------|
| Write | <1ms | 20-30ms | **5-10ms** |
| Read (hot) | <1ms | 20-50ms | **1-2ms** |
| Read (cold) | N/A | 20-50ms | **20-50ms** |
| Persistence | ❌ | ✅ | **✅** |
| Scalability | Limited | Good | **Excellent** |

## 🛠️ Configuration

```go
config := memory.Config{
    Mode: memory.HYBRID,
    
    // PostgreSQL (persistence)
    DatabaseURL: "postgresql://...",
    
    // Redis (cache)
    RedisAddr:     "localhost:6379",
    RedisPassword: "optional",
    RedisDB:       0,
    
    // Cache settings
    MaxSessionMessages: 30,        // Messages to keep in Redis
    SessionTTL:        2*time.Hour, // Redis expiry
    
    // Performance
    EnableAutoSummarize: true,      // Compress old convos
    AsyncWrites:        true,       // Non-blocking persistence
}
```

## 🔥 Cache Management

### Automatic Cache Warming
```go
// First access - loads from DB into cache
messages := mem.GetRecentMessages(ctx, sessionID, 10) // 30ms

// Subsequent accesses - served from cache
messages = mem.GetRecentMessages(ctx, sessionID, 10)  // 1ms
messages = mem.GetRecentMessages(ctx, sessionID, 10)  // 1ms
```

### TTL-Based Expiry
```go
// Messages expire from cache after TTL
config.SessionTTL = 2 * time.Hour

// After 2 hours, cache is refreshed from DB
// This prevents stale data and manages memory
```

### Cache Statistics
```go
stats := mem.GetCacheStats()
fmt.Printf("Hit rate: %.1f%%\n", stats.HitRate)
fmt.Printf("Memory usage: %d MB\n", stats.MemoryMB)
fmt.Printf("Cached sessions: %d\n", stats.Sessions)
```

## 🛡️ Failover Behavior

### Redis Available
```
Speed: ⚡ Optimal
Cache: ✅ Active
Writes: Async to both
Reads: From cache
```

### Redis Down
```
Speed: 🐢 Degraded
Cache: ❌ Bypassed
Writes: Direct to PostgreSQL
Reads: Direct from PostgreSQL
System: ✅ Still functional
```

### Recovery
```
Redis comes back online
    ↓
Auto-detected by health check
    ↓
Cache gradually warms up
    ↓
Performance restored
```

## 📈 Scaling Guidelines

### Redis Sizing

| Users | Concurrent Sessions | Redis Memory | Config |
|-------|-------------------|--------------|--------|
| 100 | 50 | 50 MB | 1 Redis instance |
| 1K | 500 | 500 MB | 1 Redis instance |
| 10K | 5,000 | 5 GB | Redis cluster |
| 100K | 50,000 | 50 GB | Sharded Redis |

### PostgreSQL Sizing

Same as persistent mode - Redis doesn't change storage needs, only improves access speed.

## 🔧 Advanced Features

### 1. Intelligent Prefetching
```go
// Prefetch likely next messages
mem.PrefetchContext(ctx, sessionID, userPattern)
```

### 2. Cache Invalidation
```go
// Selective cache clearing
mem.InvalidateSession(ctx, sessionID)
mem.InvalidateUser(ctx, userID)
```

### 3. Write-Behind Caching
```go
config.WriteBehindDelay = 5 * time.Second
// Batch writes to PostgreSQL for efficiency
```

## 💡 Best Practices

1. **Session Affinity**: Route users to same Redis node
2. **Cache Warmup**: Pre-load active sessions on startup
3. **Monitoring**: Track cache hit rates (aim for >90%)
4. **TTL Tuning**: Balance memory usage vs hit rate
5. **Connection Pools**: Properly size Redis and PG pools

## 🚨 Common Issues

### Issue: Low cache hit rate
**Solution**: Increase TTL or cache size

### Issue: Redis memory full
**Solution**: Enable eviction policy or increase memory

### Issue: Slow cold reads
**Solution**: Implement prefetching for predictable access

## 📊 Monitoring Metrics

Track these for optimal performance:
- Cache hit rate (target: >90%)
- Average latency (target: <5ms)
- Redis memory usage
- PostgreSQL connection pool usage
- Background write queue depth

## 🎯 When to Use Hybrid Mode

### ✅ Perfect For:
- Production systems
- High-traffic applications
- Low-latency requirements
- Global scale
- Cost optimization (cache reduces DB load)

### ❌ Overkill For:
- Development/testing
- Low traffic (<100 requests/hour)
- Simple prototypes

## 📚 Next Steps

- [04-semantic-search](../04-semantic-search/) - Add AI-powered search
- [05-auto-summarization](../05-auto-summarization/) - Compress conversations
- [07-agent-integration](../07-agent-integration/) - Production deployment

## 📄 Full Code

See [main.go](./main.go) for the complete implementation.
