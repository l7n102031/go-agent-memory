# Auto-Summarization Memory Example

Intelligent conversation compression to optimize token usage and costs.

## 🎯 What This Example Shows

- Automatic conversation summarization
- Token usage optimization
- Cost reduction strategies
- Summary preservation of key details
- Configurable compression thresholds

## 📋 Features

- ✅ **Auto-compression** - Summarize long conversations
- ✅ **Token optimization** - Reduce context size
- ✅ **Cost savings** - Lower API usage costs
- ✅ **Smart triggers** - Configurable thresholds
- ✅ **Detail preservation** - Keep important info
- ✅ **Background processing** - Non-blocking

## 🚀 Quick Start

### Prerequisites

1. PostgreSQL database
2. OpenAI API key (for summarization)
3. Environment variables:

```bash
# Required
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
export OPENAI_API_KEY="sk-..."

# Optional
export REDIS_URL="localhost:6379"
```

### Run the Example

```bash
go run main.go
```

## 💻 Code Overview

```go
// Configuration with auto-summarization
config := memory.Config{
    // Enable summarization
    EnableAutoSummarize: true,
    SummarizeThreshold:  10,    // Trigger after 10 messages
    SummarizeMaxTokens:  500,   // Target summary size
    SummarizeModel:      "gpt-3.5-turbo",
    
    // Archive before summarizing
    ArchiveOldMessages: true,
    
    DatabaseURL: dbURL,
    OpenAIKey:   apiKey,
}
```

## 📊 How Summarization Works

### Trigger Process
```
Conversation grows
    ↓
Reaches threshold (e.g., 10 messages)
    ↓
Trigger summarization
    ↓
Generate summary with LLM
    ↓
Archive original messages
    ↓
Store summary
    ↓
Continue with summary + recent messages
```

### Token Reduction
```
Original: 50 messages × 100 tokens = 5,000 tokens
    ↓
Summarized: 1 summary × 500 tokens = 500 tokens
    ↓
Savings: 4,500 tokens (90% reduction)
```

## 💰 Cost Impact

### Without Summarization
| Messages | Tokens/Message | Total Tokens | Cost (@$0.002/1K) |
|----------|---------------|--------------|-------------------|
| 10 | 100 | 1,000 | $0.002 |
| 50 | 100 | 5,000 | $0.010 |
| 100 | 100 | 10,000 | $0.020 |
| 500 | 100 | 50,000 | $0.100 |

### With Summarization
| Messages | Summary Tokens | Recent Messages | Total Tokens | Cost | Savings |
|----------|---------------|-----------------|--------------|------|---------|
| 10 | 0 | 1,000 | 1,000 | $0.002 | 0% |
| 50 | 500 | 1,000 | 1,500 | $0.003 | 70% |
| 100 | 1,000 | 1,000 | 2,000 | $0.004 | 80% |
| 500 | 2,500 | 1,000 | 3,500 | $0.007 | 93% |

## 🔧 Configuration Options

```go
type SummarizationConfig struct {
    // When to summarize
    SummarizeThreshold   int    // Message count trigger
    SummarizeTokenLimit  int    // Token count trigger
    
    // How to summarize
    SummarizeModel       string // LLM model to use
    SummarizeMaxTokens   int    // Target summary length
    SummarizePrompt      string // Custom prompt
    
    // What to preserve
    PreserveEntities     bool   // Keep names, dates, numbers
    PreserveDecisions    bool   // Keep action items
    PreserveQuestions    bool   // Keep unanswered questions
    
    // Storage
    ArchiveOldMessages   bool   // Keep originals in archive
    CompressArchive      bool   // Compress archived messages
}
```

## 📈 Summarization Strategies

### 1. Progressive Summarization
```go
// Summarize in stages as conversation grows
if messageCount == 10 {
    // First summary: detailed
    Summarize(maxTokens: 1000)
} else if messageCount == 50 {
    // Re-summarize: more concise
    Summarize(maxTokens: 500)
} else if messageCount == 100 {
    // Final summary: highly compressed
    Summarize(maxTokens: 250)
}
```

### 2. Importance-Based
```go
// Preserve important messages, summarize routine ones
if message.HasActionItem() || message.HasDecision() {
    KeepOriginal()
} else {
    IncludeInSummary()
}
```

### 3. Time-Based
```go
// Summarize older conversations more aggressively
if age > 24*time.Hour {
    AggressiveSummary(maxTokens: 200)
} else if age > 1*time.Hour {
    ModerateSummary(maxTokens: 500)
} else {
    KeepOriginal()
}
```

## 🎯 What Gets Preserved

### Always Preserved
- User and assistant identities
- Key decisions and conclusions
- Action items and todos
- Important numbers and dates
- Technical specifications
- Error messages and solutions

### Condensed
- Small talk and greetings
- Repetitive explanations
- Thinking process (keep conclusions)
- Examples (keep patterns)

### Example Summary
```
Original (50 messages, 5000 tokens):
"User Alice from TechCorp discussed database performance issues.
After upgrading PostgreSQL from v12 to v14, query latency increased
from 100ms to 500ms on their orders table with 50M rows.
Investigated: query plans, indexes, statistics.
Solution: Partition table by date, rebuild indexes, analyze tables.
Result: Latency reduced to 80ms. Alice will implement in staging first."

Summary (500 tokens):
Captures all key information in 10% of original space.
```

## 💡 Best Practices

1. **Threshold Tuning**: Balance between context and cost
2. **Model Selection**: Use cheaper models for routine summaries
3. **Selective Summarization**: Don't summarize critical conversations
4. **Archive Strategy**: Keep originals for compliance/debugging
5. **Summary Review**: Periodically check summary quality

## 🚨 Common Issues

### Issue: Important details lost
**Solution**: Adjust prompt to emphasize detail preservation

### Issue: Summaries too long
**Solution**: Reduce `SummarizeMaxTokens` or use more aggressive model

### Issue: Slow summarization
**Solution**: Use async processing or batch summaries

## 📊 Monitoring

Track these metrics:
- Compression ratio (target: >80%)
- Summary quality scores
- Token savings per day
- Cost reduction percentage
- User satisfaction with context

## 🔄 Integration with Agent

```go
// Agent automatically uses summaries for context
agent.Chat(ctx, sessionID, "What did we discuss about the database?")

// Memory system returns:
// 1. Summary of old conversation (500 tokens)
// 2. Recent messages (500 tokens)
// Total context: 1000 tokens instead of 5000
```

## 📚 Next Steps

- [06-event-streaming](../06-event-streaming/) - Add event tracking
- [07-agent-integration](../07-agent-integration/) - Complete system

## 📄 Full Code

See [main.go](./main.go) for the complete implementation.
