# Semantic Search Example

Advanced memory configuration with vector embeddings and semantic search using pgvector.

## 🎯 What This Example Shows

- Vector embedding generation with OpenAI
- Semantic similarity search
- Hybrid search (semantic + metadata)
- Relevance scoring
- Context retrieval for RAG

## 📋 Features

- ✅ **Semantic search** - Find similar conversations
- ✅ **Vector embeddings** - OpenAI text-embedding-3-small
- ✅ **Similarity scoring** - Cosine similarity rankings
- ✅ **Metadata filtering** - Combine vector + SQL queries
- ✅ **Persistence** - Store embeddings in pgvector
- ✅ **Scalable** - HNSW index for fast search

## 🚀 Quick Start

### Prerequisites

1. PostgreSQL with pgvector extension:
```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

2. Environment variables:
```bash
export DATABASE_URL="postgresql://user:pass@localhost:5432/agent_memory"
export OPENAI_API_KEY="sk-..."
```

3. Run the example:
```bash
go run main.go
```

## 💻 Code Overview

```go
// Configuration for semantic search
config := memory.Config{
    Mode:                 memory.PERSISTENT,
    EnableSemanticSearch: true,  // Enable embeddings
    
    DatabaseURL: os.Getenv("DATABASE_URL"),
    OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
    
    // Embedding configuration
    EmbeddingModel:  "text-embedding-3-small",
    VectorDimension: 1536,
}

// Search for similar messages
results, _ := mem.Search(ctx, "How to handle errors in Go?", 5, 0.7)
for _, result := range results {
    fmt.Printf("Score: %.2f - %s\n", result.Score, result.Content)
}
```

## 🔍 Search Capabilities

### Basic Semantic Search
```go
// Find messages similar to a query
results := mem.Search(ctx, "database connection issues", limit=5, threshold=0.7)
```

### Search with Pre-computed Embedding
```go
// If you already have an embedding
embedding := generateEmbedding("your query")
results := mem.SearchWithEmbedding(ctx, embedding, limit=5, threshold=0.8)
```

### Hybrid Search (Semantic + Filters)
```go
// Combine vector search with metadata filters
results := mem.HybridSearch(ctx, HybridQuery{
    Query:     "error handling",
    UserID:    "user-123",
    TimeRange: TimeRange{Start: yesterday, End: today},
    Limit:     10,
})
```

## 📊 How It Works

### Embedding Generation
1. User message received
2. Generate embedding via OpenAI API
3. Store message + embedding in PostgreSQL
4. Index with HNSW for fast retrieval

### Search Process
```
Query: "How to handle errors?"
    ↓
Generate Query Embedding (OpenAI)
    ↓
Vector Similarity Search (pgvector)
    ↓
Apply Filters (SQL WHERE)
    ↓
Return Ranked Results
```

### Distance Metrics
- **Cosine Similarity** (default): Best for normalized embeddings
- **L2 Distance**: Euclidean distance
- **Inner Product**: For specific use cases

## 🎛️ Configuration Options

```go
type SemanticSearchConfig struct {
    // Embedding settings
    EmbeddingModel   string  // "text-embedding-3-small" or "text-embedding-3-large"
    VectorDimension  int     // 1536 for small, 3072 for large
    
    // Search settings
    DefaultThreshold float32 // Minimum similarity score (0.0 to 1.0)
    MaxResults      int      // Maximum results to return
    
    // Performance
    UseHNSWIndex    bool     // Use HNSW index (recommended)
    HNSWParams      HNSWConfig{
        M:              16,   // Number of connections
        EfConstruction: 64,   // Build-time accuracy
    }
}
```

## 🧪 Example Scenarios

### Scenario 1: Finding Related Conversations
```go
// User asks about a previous topic
query := "What did we discuss about authentication last week?"
results := mem.Search(ctx, query, 10, 0.75)
```

### Scenario 2: Context Retrieval for RAG
```go
// Get relevant context for answering a question
query := userQuestion
context := mem.Search(ctx, query, 5, 0.8)

// Build prompt with context
prompt := buildPromptWithContext(userQuestion, context)
```

### Scenario 3: Duplicate Detection
```go
// Check if similar question was already asked
newMessage := "How do I connect to PostgreSQL?"
duplicates := mem.Search(ctx, newMessage, 1, 0.95) // High threshold

if len(duplicates) > 0 {
    fmt.Println("Similar question found:", duplicates[0].Content)
}
```

## 📈 Performance Metrics

| Dataset Size | Index Build Time | Search Time | Memory Usage |
|-------------|------------------|-------------|--------------|
| 1K vectors | < 1s | ~5ms | 10 MB |
| 10K vectors | ~5s | ~10ms | 100 MB |
| 100K vectors | ~30s | ~20ms | 1 GB |
| 1M vectors | ~5min | ~50ms | 10 GB |

## 💰 Cost Analysis

### OpenAI Embeddings
- text-embedding-3-small: $0.02 / 1M tokens
- text-embedding-3-large: $0.13 / 1M tokens

### Example Costs
| Messages/Day | Avg Length | Model | Daily Cost |
|-------------|------------|-------|------------|
| 1,000 | 100 tokens | Small | $0.002 |
| 10,000 | 100 tokens | Small | $0.02 |
| 100,000 | 100 tokens | Small | $0.20 |

## 🔧 Optimization Tips

1. **Batch Embeddings**: Generate embeddings in batches to reduce API calls
2. **Cache Embeddings**: Store and reuse embeddings for identical content
3. **Async Generation**: Don't block on embedding generation
4. **Index Tuning**: Adjust HNSW parameters for your dataset size
5. **Dimension Reduction**: Consider smaller models if cost is a concern

## 🚨 Common Issues

### Issue: Slow Search Performance
**Solution**: Create HNSW index:
```sql
CREATE INDEX ON agent_messages USING hnsw (embedding vector_cosine_ops);
```

### Issue: Poor Search Results
**Solution**: Adjust threshold or try different embedding model

### Issue: High Costs
**Solution**: Use text-embedding-3-small or implement caching

## 📚 Next Steps

- [05-auto-summarization](../05-auto-summarization/) - Add intelligent summarization
- [06-event-streaming](../06-event-streaming/) - Add event sourcing
- [07-agent-integration](../07-agent-integration/) - Complete RAG implementation

## 📄 Full Code

See [main.go](./main.go) for the complete implementation.
