# Slack Translation Bot

A high-performance Slack bot that automatically translates English messages to Vietnamese using Google Gemini AI. Built with Go, Clean Architecture, and containerized for easy deployment.

## Features

- **Automatic Translation**: Translates 🇬🇧 English ⇄ 🇻🇳 Vietnamese automatically
- **Offline Language Detection**: Fast language detection using lingua-go (75+ languages)
- **Redis Caching**: 24-hour cache to reduce API calls and improve response time
- **Secure & Observable**: Request signature verification, health checks, metrics, and structured logging

## Tech Stack

- **Language**: Go 1.25+
- **Framework**: Gin (HTTP routing)
- **AI**: Google Gemini API (gemini-1.5-flash)
- **Database**: MySQL
- **Cache**: Redis
- **Language Detection**: lingua-go
- **Architecture**: Clean Architecture with layered structure
- **Logging**: Uber's zap (structured logging)
- **Testing**: testify + stretchr for assertions

See [ARCHITECTURE.md](./docs/ARCHITECTURE.md) for detailed design and system flow.

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose (or standalone MySQL + Redis)
- Slack workspace admin access
- Google Gemini API key (free tier)

### Setup

1. **Create Slack App**
   - Go to [api.slack.com/apps](https://api.slack.com/apps)
   - Create new app from scratch
   - Add bot scopes: `channels:history`, `channels:read`, `chat:write`, `groups:history`, `groups:read`, `im:history`, `users:read`, `reactions:write`, `reactions:read`, `app_mentions:read`
   - Enable Event Subscriptions and subscribe to:
     - `message.channels`
     - `reaction_added`
     - `app_mention`
   - Install to workspace and copy Bot Token

2. **Get Gemini API Key**
   - Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
   - Generate free API key

3. **Configure Environment**
   ```bash
   cp .env.example .env
   ```
   Edit `.env` with:
   ```bash
   SLACK_BOT_TOKEN=xoxb-...
   SLACK_SIGNING_SECRET=...
   SLACK_APP_TOKEN=xapp-...
   GEMINI_API_KEY=AIza...
   DATABASE_HOST=localhost
   REDIS_HOST=localhost
   ```

4. **Run with Docker Compose**
   ```bash
   make docker-up           # Start MySQL & Redis
   make migrate-up          # Run database migrations
   go run cmd/api/main.go   # Start the bot
   ```

## API Endpoints

- `POST /slack/events` - Slack webhook for events (requires signature verification)
- `GET /health` - Health check endpoint (returns database and Redis status)
- `GET /metrics` - Metrics endpoint (translation stats, cache hit rate, etc.)

## Demo


## License

MIT - see [LICENSE](./LICENSE)
