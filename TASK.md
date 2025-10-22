# Slack Translation Bot - Task List

Based on the architecture document, this task list outlines the MVP implementation plan for the Slack Translation Bot using Clean Architecture in Golang.

---

## High Priority (Core MVP - Week 1-2)

### Foundation & Setup

- [x] **1. Project structure setup**
  - Create Clean Architecture folders (domain, usecase, infrastructure, interface)
  - Set up cmd/bot/main.go entry point

- [x] **2. Environment setup**
  - Initialize Go module
  - Configure .env template with all required variables
  - Add dependencies (chi, zap, viper, testify, slack-go, google.generativeai, lingua-go, mysql driver, go-redis)

- [x] **3. Domain entities and interfaces**
  - Create entity models (message.go, translation.go, channel_config.go)
  - Define repository interfaces (translation.go, channel.go)
  - Create domain service interfaces (translator.go)

### Database & Storage

- [x] **4. MySQL schema and migrations**
  - Create database tables (channel_configs, translations, usage_metrics)
  - Add proper indexes and constraints
  - Set up golang-migrate for database migrations

- [x] **5. Redis caching layer**
  - Implement cache structure for translations (key: translation:{hash}, TTL: 24h)
  - Implement cache for channel configs (key: channel_config:{channel_id}, TTL: 1h)
  - Create cache interface and implementation

### AI & Language Detection

- [x] **6. Language detection integration**
  - Integrate lingua-go library for offline language detection
  - Create detector wrapper in infrastructure/language/detector.go
  - Test accuracy with English/Vietnamese samples

- [x] **7. Google Gemini integration**
  - Create AI provider interface in infrastructure/ai/provider.go
  - Implement Gemini client with prompt template
  - Handle API errors and rate limits

### Core Business Logic

- [x] **8. Translation use case**
  - Implement core translation logic with cache check → AI call → storage flow
  - Handle cache hits/misses
  - Store translations in MySQL
  - Update cache after successful translation

### Slack Integration

- [x] **9. Slack webhook handler**
  - Create HTTP endpoint with chi router
  - Implement signature verification middleware
  - Set up event URL verification

- [x] **10. Event processing**
  - Handle message.channels events
  - Handle reaction_added events (🇻🇳 emoji)
  - Parse and validate event payloads

- [x] **11. Thread reply posting**
  - Implement Slack API integration to post translations as threaded replies
  - Format response with 🇻🇳 emoji prefix
  - Handle API errors and retries

---

## Medium Priority (Essential Features)

### Configuration & Management

- [x] **12. Channel configuration**
  - Create use case for managing channel settings
  - Implement CRUD operations for channel_configs table
  - Add auto_translate, languages, and enabled flags

### Security & Validation

- [x] **13. Basic error handling**
  - Add structured logging with zap
  - Implement error response formatting
  - Add correlation IDs for request tracking

- [x] **14. Rate limiting**
  - Implement per-user rate limit (10 translations/min)
  - Implement per-channel rate limit (30 translations/min)
  - Return appropriate error messages when limits exceeded

- [x] **15. Input validation**
  - Add message sanitization to prevent prompt injection
  - Implement length limits (10KB max)
  - Validate Slack event payloads

### Deployment

- [x] **16. Docker containerization**
  - Create Dockerfile for Go application
  - Create docker-compose.yml with MySQL and Redis services
  - Configure environment variable mapping

### Testing

- [x] **17. Unit tests**
  - Write tests for domain entities and services
  - Write tests for use cases with mocks
  - Target 80%+ code coverage

---

## Low Priority (Quality & Monitoring)

### Testing & Quality Assurance

- [ ] **18. Integration tests**
  - Test database operations with test MySQL instance
  - Test Redis caching behavior
  - Test Gemini API calls with test API key
  - Mock Slack API responses

### Operations

- [ ] **19. Health check endpoint**
  - Add /health endpoint for monitoring
  - Check database connectivity
  - Check Redis connectivity
  - Return status JSON

- [ ] **20. Monitoring & metrics**
  - Track translation requests per channel/user
  - Track API latency and success rate
  - Track cache hit rate
  - Track Gemini API token usage
  - Track error rates by type

---

## Success Criteria

### MVP Success
- ✅ Translates English → Vietnamese accurately
- ✅ Posts in threads, doesn't clutter channels
- ✅ Responds within 3 seconds
- ✅ Handles 10+ concurrent translations
- ✅ No crashes under normal load
- ✅ Proper error handling and logging

### Production Ready (Post-MVP)
- ✅ 99.5% uptime
- ✅ <2s average response time
- ✅ Comprehensive monitoring
- ✅ Automated deployment
- ✅ Documentation complete
- ✅ Security audit passed

---

## Timeline

**Week 1**: Tasks 1-8 (Foundation, Database, AI Integration, Core Logic)  
**Week 2**: Tasks 9-17 (Slack Integration, Security, Deployment, Testing)  
**Week 3+**: Tasks 18-20 (Quality Assurance, Monitoring)

---

## Notes

- Follow Clean Architecture principles strictly
- Write tests alongside implementation
- Review security best practices before deployment
- Use .env.example for configuration template (never commit secrets)
- Test with ngrok during local development
