# Persistent Basic Memory Example

Database-backed memory with PostgreSQL for production-ready persistent storage.

## 🎯 What This Example Shows

- PostgreSQL persistence
- Data survives application restarts
- Multiple concurrent sessions
- Cross-session data access
- Production-ready configuration

## 📋 Features

- ✅ **Full persistence** - Data stored in PostgreSQL
- ✅ **Unlimited storage** - Database-limited only
- ✅ **Multi-session** - Handle many users simultaneously
- ✅ **Queryable** - SQL access to all history
- ⚠️ **Network latency** - Database round-trips
- ❌ **No caching** - Every read hits database

## 🚀 Quick Start

### Prerequisites

1. PostgreSQL database (local or cloud)
2. Environment variables:

```bash
# Required
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"

# Optional (for semantic search)
export OPENAI_API_KEY="sk-..."
```

### Run the Example

```bash
go run main.go
```

## 💻 Code Overview

```go
// Configuration for persistent mode
config := memory.Config{
    Mode:              memory.PERSISTENT,
    EnablePersistence: true,
    
    DatabaseURL: os.Getenv("DATABASE_URL"),
    
    // Optional for embeddings
    OpenAIKey: os.Getenv("OPENAI_API_KEY"),
    
    MaxSessionMessages: 100, // Can be higher with DB
}

// Data persists across restarts
mem.AddMessage(ctx, message)
mem.Close()

// ... application restarts ...

mem = memory.NewWithConfig(config)
messages := mem.GetRecentMessages(ctx, sessionID, 10)
// Messages are still there!
```

## 🗄️ Database Setup

### Option 1: Local PostgreSQL

```bash
# Install PostgreSQL
brew install postgresql

# Start service
brew services start postgresql

# Create database
createdb agent_memory

# Set environment
export DATABASE_URL="postgresql://localhost:5432/agent_memory"
```

### Option 2: Supabase (Recommended)

```bash
# Create free account at supabase.com
# Get connection string from Settings > Database
export DATABASE_URL="postgresql://postgres:[password]@[project].supabase.co:5432/postgres"
```

### Option 3: Docker

```bash
# Use provided docker-compose
cd ../../deployment
docker-compose up -d postgres

export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/agent_memory"
```

## 📊 Storage Behavior

### Message Persistence
```
Application Start
    ↓
Connect to PostgreSQL
    ↓
Store Messages → Database (permanent)
    ↓
Application Stop
    ↓
Data Remains in Database ✅
    ↓
Application Restart
    ↓
Retrieve Previous Messages ✅
```

### Performance Characteristics

| Operation | Latency | Throughput |
|-----------|---------|------------|
| Write Message | 20-30ms | ~50/sec |
| Read Recent | 20-50ms | ~100/sec |
| Batch Read | 30-100ms | ~20/sec |
| Search | 50-200ms | ~10/sec |

## 🔄 Migration from Session-Only

```go
// Before (Session-Only)
config := memory.Config{
    Mode: memory.SESSION_ONLY,
}

// After (Persistent)
config := memory.Config{
    Mode: memory.PERSISTENT,
    DatabaseURL: os.Getenv("DATABASE_URL"),
}
// That's it! Same API, now with persistence
```

## 📈 Scaling Considerations

### Database Sizing

| Users | Messages/Day | Storage | Recommended |
|-------|-------------|---------|-------------|
| 100 | 10,000 | ~10 MB/day | Hobby tier |
| 1,000 | 100,000 | ~100 MB/day | Small instance |
| 10,000 | 1,000,000 | ~1 GB/day | Medium instance |
| 100,000 | 10,000,000 | ~10 GB/day | Large instance |

### Connection Pooling

```go
// The library handles connection pooling automatically
config := memory.Config{
    DatabaseURL: dbURL,
    // Pool settings handled internally:
    // - MaxConnections: 25
    // - MinConnections: 5
    // - MaxIdleTime: 30min
}
```

## 🚨 Common Issues

### Issue: "connection refused"
**Solution**: Ensure PostgreSQL is running and accessible

### Issue: "database does not exist"
**Solution**: Create the database first:
```bash
createdb agent_memory
```

### Issue: Slow queries
**Solution**: The library creates indexes automatically, but verify:
```sql
-- Check indexes
\di agent_messages
```

## 💡 Best Practices

1. **Connection String Security**: Never hardcode, use environment variables
2. **Backup Strategy**: Regular pg_dump for production
3. **Monitoring**: Track slow queries and connection pool usage
4. **Retention**: Implement data cleanup for old messages
5. **Indexing**: Additional indexes for custom queries

## 🔧 Advanced Configuration

```go
// With more features enabled
config := memory.Config{
    Mode:              memory.PERSISTENT,
    EnablePersistence: true,
    
    // Database
    DatabaseURL: dbURL,
    
    // Add semantic search
    EnableSemanticSearch: true,
    OpenAIKey:           apiKey,
    EmbeddingModel:      "text-embedding-3-small",
    
    // Add auto-summarization
    EnableAutoSummarize:  true,
    SummarizeThreshold:   50, // Summarize after 50 messages
    
    // Connection tuning
    MaxRetries:          3,
    RetryDelay:         time.Second,
}
```

## 📚 Next Steps

Ready for better performance? Check out:
- [03-hybrid-mode](../03-hybrid-mode/) - Add Redis caching layer
- [04-semantic-search](../04-semantic-search/) - Enable vector search
- [05-auto-summarization](../05-auto-summarization/) - Compress old conversations

## 📄 Full Code

See [main.go](./main.go) for the complete implementation.
