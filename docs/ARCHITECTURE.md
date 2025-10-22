# Slack Translation Bot - Architecture Document

## Overview
A Slack bot that translates English messages to Vietnamese using Clean Architecture principles in Golang, with generative AI for high-quality translation.

## Architecture Decisions

### 1. AI Provider Strategy
- **Phase 1 (MVP)**: Google Gemini API (Free tier)
  - Model: gemini-1.5-flash
  - Free tier: 15 req/min, 1M tokens/day
  - No cost, good quality
- **Phase 2**: Self-hosted Ollama + Llama 3
  - Complete privacy
  - No API costs
  - Requires GPU infrastructure

### 2. Translation Trigger
- **Opt-in Channels**: Admins configure which channels to monitor
- **Emoji Reactions**: Users can trigger with 🇻🇳 or 🇬🇧 reactions
- **Channel Settings**: Store per-channel configuration in database

### 3. Response Format
- **Thread Replies**: Post translations as threaded replies
- **Format**: `🇻🇳 {translated_text}` (Vietnamese) or `🇬🇧 {translated_text}` (English)
- **Non-intrusive**: Keeps main channel clean

### 4. Data Storage
- **MySQL**: Store translations, channel configs, user preferences (Railway free tier)
- **Redis**: Cache recent translations (24h TTL)
- **Metrics**: Track usage patterns and analytics

### 5. Scalability
- **Target**: Small team (<100 users)
- **Deployment**: Single Docker container
- **Database**: MySQL on Railway (free tier)
- **Hosting**: Railway free tier or ~$5-10/month

### 6. Security
- **Slack Verification**: Verify signing secret on all requests
- **Environment Variables**: All secrets in .env
- **Rate Limiting**: Per-channel and per-user limits
- **Input Validation**: Sanitize all inputs

### 7. Development Approach
- **MVP First**: 2-week sprint
- **Core Features**: Translation working end-to-end
- **Clean Interfaces**: Easy to extend later
- **Iterative**: Get feedback and improve

### 8. Language Detection
- **Library**: lingua-go
- **Offline**: No API calls needed
- **Fast**: Millisecond detection
- **Accurate**: Supports 75+ languages

---

## Clean Architecture Layers

### Domain Layer (Core Business Logic)
```
internal/domain/
├── entity/
│   ├── message.go          # Message entity
│   ├── translation.go      # Translation entity
│   └── channel_config.go   # Channel configuration
├── repository/             # Repository interfaces
│   ├── translation.go
│   └── channel.go
└── service/                # Domain services
    └── translator.go       # Translation business rules
```

### Use Case Layer (Application Logic)
```
internal/usecase/
├── translate_message.go    # Core translation logic
├── handle_slack_event.go   # Process Slack events
└── manage_channel.go       # Channel configuration
```

### Infrastructure Layer (External Dependencies)
```
internal/infrastructure/
├── ai/
│   ├── provider.go         # AI provider interface
│   ├── gemini.go          # Google Gemini client
│   └── ollama.go          # Ollama client (Phase 2)
├── slack/
│   ├── client.go          # Slack API wrapper
│   └── event_handler.go   # Event processing
├── persistence/
│   ├── mysql.go           # MySQL repositories
│   └── redis.go           # Redis cache
└── language/
    └── detector.go        # lingua-go wrapper
```

### Interface Layer (Presentation)
```
internal/interface/
├── http/
│   ├── handler/
│   │   ├── slack_events.go   # Slack webhook handler
│   │   ├── slack_commands.go # Slash command handler
│   │   └── health.go         # Health check
│   └── middleware/
│       ├── verify_slack.go   # Slack signature verification
│       └── logging.go        # Request logging
└── presenter/
    └── translation.go        # Format responses
```

---

## System Flow

### Message Translation Flow
```
1. User posts English message in monitored channel
2. Slack sends event to webhook endpoint
3. Verify Slack signature
4. Detect language using lingua-go
5. If not English, skip
6. Check Redis cache for translation
7. If cache miss, call Gemini API
8. Store translation in PostgreSQL
9. Cache in Redis (24h)
10. Post as thread reply with 🇻🇳 emoji
```

### Emoji Reaction Flow
```
1. User adds 🇻🇳 reaction to any message
2. Slack sends reaction_added event
3. Bot fetches message text
4. Translate (check cache → AI → store)
5. Post translation in thread
```

---

## Database Schema

### MySQL Tables

```sql
-- Channel configurations
CREATE TABLE channel_configs (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    channel_id VARCHAR(50) UNIQUE NOT NULL,
    channel_name VARCHAR(100),
    auto_translate BOOLEAN DEFAULT false,
    source_lang VARCHAR(10) DEFAULT 'en',
    target_lang VARCHAR(10) DEFAULT 'vi',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_channel_id (channel_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Translation history
CREATE TABLE translations (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    original_text TEXT NOT NULL,
    translated_text TEXT NOT NULL,
    source_lang VARCHAR(10) DEFAULT 'en',
    target_lang VARCHAR(10) DEFAULT 'vi',
    slack_user_id VARCHAR(50),
    slack_channel_id VARCHAR(50),
    slack_message_ts VARCHAR(50),
    slack_thread_ts VARCHAR(50),
    ai_provider VARCHAR(50),
    tokens_used INT,
    latency_ms INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_translations_channel (slack_channel_id, created_at),
    INDEX idx_translations_message (slack_message_ts),
    INDEX idx_translations_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Usage metrics
CREATE TABLE usage_metrics (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    channel_id VARCHAR(50),
    user_id VARCHAR(50),
    translations_count INT DEFAULT 0,
    date DATE DEFAULT (CURRENT_DATE),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_channel_user_date (channel_id, user_id, date),
    INDEX idx_metrics_date (date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### Redis Cache Structure
```
Key: translation:{hash(original_text)}
Value: {translated_text}
TTL: 24 hours

Key: channel_config:{channel_id}
Value: JSON config
TTL: 1 hour
```

---

## API Integration

### Google Gemini API
```go
// Prompt template
System: You are a professional English to Vietnamese translator.
Provide natural, accurate translations maintaining tone and context.

User: Translate to Vietnamese: "{text}"

Response: Only output the Vietnamese translation, no explanations.
```

### Slack API Events
- `message.channels` - New messages in channels
- `reaction_added` - User adds emoji reaction
- `app_mention` - Bot is mentioned

---

## Environment Variables

```bash
# Slack Configuration
SLACK_BOT_TOKEN=xoxb-...
SLACK_SIGNING_SECRET=...
SLACK_APP_TOKEN=xapp-...

# AI Provider
AI_PROVIDER=gemini          # gemini, ollama, groq
GEMINI_API_KEY=AIza...
OLLAMA_HOST=http://localhost:11434

# Database
DATABASE_URL=mysql://user:pass@localhost:3306/slack_translator
REDIS_URL=redis://localhost:6379/0

# Application
PORT=8080
LOG_LEVEL=info
ENVIRONMENT=development
CACHE_TTL_HOURS=24
```

---

## MVP Features (Week 1-2)

### Week 1: Foundation
- [ ] Project structure setup
- [ ] Google Gemini integration
- [ ] Language detection (lingua-go)
- [ ] MySQL schema and migrations
- [ ] Redis caching layer
- [ ] Domain entities and interfaces

### Week 2: Slack Integration
- [ ] Slack webhook handler
- [ ] Event processing (messages)
- [ ] Thread reply posting
- [ ] Channel configuration
- [ ] Basic error handling
- [ ] Docker containerization

---

## Future Enhancements (Post-MVP)

### Phase 2: Advanced Features
- [ ] Self-hosted Ollama integration
- [ ] Bi-directional translation (VI → EN)
- [ ] `/translate` slash command
- [ ] User preferences (opt-out)
- [ ] Translation quality feedback
- [ ] Web dashboard for analytics

### Phase 3: Scale & Polish
- [ ] Kubernetes deployment
- [ ] Multiple AI provider fallback
- [ ] Advanced caching strategies
- [ ] Comprehensive monitoring
- [ ] Load testing
- [ ] Security audit

---

## Technology Stack

### Core
- **Language**: Go 1.21+
- **Framework**: Standard library + chi router
- **Slack**: slack-go/slack

### AI Providers
- **Gemini**: google.generativeai
- **Ollama**: ollama/ollama-go (Phase 2)

### Storage
- **Database**: MySQL (go-sql-driver/mysql)
- **Cache**: Redis (go-redis)
- **Migrations**: golang-migrate/migrate (MySQL)

### Utilities
- **Config**: viper
- **Logging**: zap (structured logging)
- **Testing**: testify + mockery
- **Language Detection**: lingua-go

### DevOps
- **Container**: Docker
- **Orchestration**: docker-compose
- **CI/CD**: Jenkins

---

## Cost Analysis (MVP)

### Free Tier Resources
- **Gemini API**: FREE (1M tokens/day)
- **MySQL**: Railway free tier (500MB storage)
- **Redis**: Railway/Upstash free tier or local

### Paid Resources
- **Hosting**: $5-10/month (Fly.io, Railway, DigitalOcean)
- **Domain** (optional): $10/year

**Total MVP Cost**: $5-10/month

### Phase 2 (Self-hosted)
- **GPU Server**: $50-100/month for decent GPU
- **Or**: Use existing hardware with GPU

---

## Monitoring & Observability

### Metrics to Track
- Translation requests per channel/user
- API latency and success rate
- Cache hit rate
- Gemini API token usage
- Error rates by type

### Logging
- Structured JSON logs (zap)
- Request/response correlation IDs
- Error stack traces
- User actions (privacy-conscious)

### Alerts
- API failures (> 5% error rate)
- High latency (> 3s)
- Rate limit approaching
- Database connection issues

---

## Security Best Practices

1. **Slack Request Verification**
   - Validate signing secret on every request
   - Check timestamp to prevent replay attacks

2. **Environment Variables**
   - Never commit secrets to git
   - Use .env.example as template

3. **Rate Limiting**
   - Per-user: 10 translations/minute
   - Per-channel: 30 translations/minute

4. **Input Validation**
   - Sanitize message content
   - Limit message length (10KB max)
   - Prevent prompt injection

5. **Error Handling**
   - Never expose API keys in errors
   - Log errors without sensitive data
   - Generic error messages to users

---

## Testing Strategy

### Unit Tests
- Domain logic (entities, services)
- Use case implementations
- Mock external dependencies
- Target: 80%+ coverage

### Integration Tests
- Database operations
- Redis caching
- Gemini API calls (with test API key)
- Slack API mocks

### End-to-End Tests
- Full flow from webhook to translation
- Use Slack testing workspace
- Verify thread replies

---

## Deployment

### Local Development
```bash
docker-compose up -d  # Start MySQL + Redis
go run cmd/bot/main.go
ngrok http 8080  # Expose webhook endpoint
```

### Production (Railway)
```bash
# Connect to Railway
railway login

# Create new project
railway init

# Add MySQL and Redis
railway add --mysql
railway add --redis

# Set environment variables
railway variables set SLACK_BOT_TOKEN=...
railway variables set GEMINI_API_KEY=...

# Deploy
railway up
```

---

## Success Criteria

### MVP Success
- ✅ Translates English → Vietnamese accurately
- ✅ Posts in threads, doesn't clutter channels
- ✅ Responds within 3 seconds
- ✅ Handles 10+ concurrent translations
- ✅ No crashes under normal load
- ✅ Proper error handling and logging

### Production Ready
- ✅ 99.5% uptime
- ✅ <2s average response time
- ✅ Comprehensive monitoring
- ✅ Automated deployment
- ✅ Documentation complete
- ✅ Security audit passed
