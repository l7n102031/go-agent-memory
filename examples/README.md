# 📚 Memory System Examples

A collection of practical examples demonstrating different memory configurations and use cases for AI agents.

## 🗂️ Example Categories

### Basic Configurations

| Example | Description | Dependencies | Difficulty |
|---------|-------------|--------------|------------|
| [01-session-only](./01-session-only/) | In-memory only, no persistence | None | ⭐ Beginner |
| [02-persistent-basic](./02-persistent-basic/) | Basic database persistence | PostgreSQL | ⭐⭐ Easy |
| [03-hybrid-mode](./03-hybrid-mode/) | Redis + PostgreSQL combination | Redis, PostgreSQL | ⭐⭐⭐ Intermediate |

### Advanced Features

| Example | Description | Dependencies | Difficulty |
|---------|-------------|--------------|------------|
| [04-semantic-search](./04-semantic-search/) | Vector search with pgvector | PostgreSQL, OpenAI | ⭐⭐⭐ Intermediate |
| [05-auto-summarization](./05-auto-summarization/) | Automatic conversation summaries | PostgreSQL, OpenAI | ⭐⭐⭐ Intermediate |
| [06-event-streaming](./06-event-streaming/) | Redis Streams for event sourcing | Redis | ⭐⭐⭐⭐ Advanced |

### Integration Examples

| Example | Description | Dependencies | Difficulty |
|---------|-------------|--------------|------------|
| [07-agent-integration](./07-agent-integration/) | Complete agent with memory | All optional | ⭐⭐⭐⭐ Advanced |

## 🚀 Quick Start

### 1. Choose Your Configuration

Start with the example that matches your needs:

- **No external dependencies?** → Start with [01-session-only](./01-session-only/)
- **Need persistence?** → Try [02-persistent-basic](./02-persistent-basic/)
- **Want best performance?** → Check [03-hybrid-mode](./03-hybrid-mode/)
- **Building an AI agent?** → See [07-agent-integration](./07-agent-integration/)

### 2. Set Up Environment

Each example folder contains its own `README.md` with specific setup instructions. Generally:

```bash
# For session-only (no setup needed)
cd 01-session-only
go run main.go

# For persistent modes
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
export OPENAI_API_KEY="sk-..."
cd 02-persistent-basic
go run main.go

# For hybrid mode
export REDIS_URL="localhost:6379"
cd 03-hybrid-mode
go run main.go
```

### 3. Understand the Progression

Examples are numbered to show a learning progression:

```
01-session-only (simplest)
    ↓
02-persistent-basic (add database)
    ↓
03-hybrid-mode (add caching)
    ↓
04-semantic-search (add AI search)
    ↓
05-auto-summarization (add smart compression)
    ↓
06-event-streaming (add event sourcing)
    ↓
07-agent-integration (combine everything)
```

## 📊 Configuration Matrix

| Feature | Session-Only | Persistent | Hybrid | Semantic | Auto-Summary | Event Stream |
|---------|-------------|------------|--------|----------|--------------|--------------|
| No Dependencies | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Persistence | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Fast Cache | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| Vector Search | ❌ | ❌ | ❌ | ✅ | ✅ | Optional |
| Summarization | ❌ | ❌ | ❌ | ❌ | ✅ | Optional |
| Event Replay | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |

## 🛠️ Development Workflow

### Testing Different Modes

```bash
# Run examples with different configurations
make run-examples

# Or individually
go run examples/01-session-only/main.go
go run examples/02-persistent-basic/main.go
# ... etc
```

### Docker Setup for Dependencies

```bash
# Start all dependencies for testing
cd deployment
docker-compose up -d

# Now all examples will work
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/agent_memory"
export REDIS_URL="localhost:6379"
```

## 📖 Learning Path

### For Beginners
1. Start with **01-session-only** to understand basic concepts
2. Move to **02-persistent-basic** to add database storage
3. Try **07-agent-integration** to see a complete implementation

### For Production Users
1. Review **03-hybrid-mode** for optimal performance
2. Implement **04-semantic-search** for intelligent retrieval
3. Add **05-auto-summarization** for token optimization
4. Consider **06-event-streaming** for debugging/audit

## 💡 Tips

- Each example is self-contained and runnable
- Examples progressively build on each other
- All examples include error handling and comments
- Check individual READMEs for specific details

## 🤝 Contributing

Have a useful example? Please contribute!

1. Create a new folder following the naming pattern
2. Include a clear README.md
3. Add comprehensive comments in code
4. Submit a pull request

## 📝 License

All examples are MIT licensed and free to use in your projects.
