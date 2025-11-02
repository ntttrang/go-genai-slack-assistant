# Slack Translation Bot

A Slack bot that automatically translates English messages to Vietnamese using Google Gemini AI. Built with Go, Clean Architecture, and containerized for easy deployment.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-Framework-00ADD8?style=flat&logo=go&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-Database-4479A1?style=flat&logo=mysql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?style=flat&logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Containerized-2496ED?style=flat&logo=docker&logoColor=white)
![Gemini](https://img.shields.io/badge/Google-Gemini_AI-4285F4?style=flat&logo=google&logoColor=white)

[![Demo Video](https://blog.n8n.io/content/images/size/w1200/2024/05/post-slack-bot3--1-.png)](https://www.youtube.com/@trang-nguyen-thi-thuy)

## Features

- **Automatic Translation**: Translates messages between English and Vietnamese in Slack channels using Google Gemini AI
- **Smart Language Detection**: Offline language detection with lingua-go supporting 75+ languages for fast, accurate identification
- **Clean Architecture**: Follows Go standard layout with layered architecture for maintainability and testability

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
- **Containerization**: Docker & Docker Compose

## Architecture

```text
┌─────────────┐         ┌─────────────┐
│   Slack     │────────▶│  Gin HTTP   │
│  Workspace  │         │   Server    │
└─────────────┘         └──────┬──────┘
                               │
                    ┌──────────▼──────────┐
                    │   Controllers       │
                    │  (Event Handlers)   │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │    Use Cases        │
                    │  (Business Logic)   │
                    └──────────┬──────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
┌───────▼────────┐    ┌────────▼────────┐   ┌────────▼────────┐
│  Repositories  │    │  Gemini AI API  │   │  Redis Cache    │
│    (MySQL)     │    │   (Translation) │   │  (24h TTL)      │
└────────────────┘    └─────────────────┘   └─────────────────┘
```

## Project Structure

```text
.
├── cmd/
│   └── api/                 # Application entry point
├── internal/
│   ├── controller/          # HTTP handlers (Slack events, metrics, health)
│   ├── service/             # Business logic (translation, channel, message)
│   ├── repository/          # Data access layer (MySQL)
│   ├── model/               # Domain models
│   ├── dto/                 # Data transfer objects
│   ├── middleware/          # HTTP middleware
│   └── translator/          # Gemini AI client
├── pkg/
│   ├── ai/                  # AI utilities
│   ├── cache/               # Redis cache client
│   ├── config/              # Configuration management
│   ├── database/            # Database connection
│   ├── language/            # lingua-go language detection
│   ├── logger/              # Zap logger setup
│   ├── metrics/             # Metrics collection
│   └── ratelimit/           # Rate limiting
├── database/                # Database migrations
├── tests/                   # Integration tests only
├── docs/                    # Documentation
├── scripts/                 # Utility scripts
├── docker-compose.yml       # Docker services
├── Dockerfile               # Container image
└── Makefile                 # Build and deployment commands
```

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose (or standalone MySQL + Redis)
- Slack workspace admin access
- Google Gemini API key (free tier available at [Google AI Studio](https://makersuite.google.com/app/apikey))

### Setup

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd go-genai-slack-assistant
   ```

2. **Create Slack App**
   - Go to [api.slack.com/apps](https://api.slack.com/apps)
   - Create new app from scratch
   - Add bot scopes: `channels:history`, `channels:read`, `chat:write`, `groups:history`, `groups:read`, `im:history`, `users:read`, `reactions:write`, `reactions:read`, `app_mentions:read`
   - Enable Event Subscriptions and subscribe to:
     - `message.channels`
     - `reaction_added`
     - `app_mention`
   - Install to workspace and copy Bot Token
   - See detailed setup guide: [SLACK_SETUP.md](./docs/SLACK_SETUP.md)

3. **Get Gemini API Key**

   - Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
   - Generate free API key

4. **Configure Environment**

   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your credentials:

   ```bash
   SLACK_BOT_TOKEN=xoxb-...
   SLACK_SIGNING_SECRET=...
   SLACK_APP_TOKEN=xapp-...
   GEMINI_API_KEY=AIza...
   DATABASE_HOST=localhost
   DATABASE_PORT=3306
   REDIS_HOST=localhost
   REDIS_PORT=6379
   ```

5. **Start Services and Run**

   ```bash
   make docker-up           # Start MySQL & Redis containers
   make migrate-up          # Run database migrations
   go run cmd/api/main.go   # Start the bot
   ```

   **Available Make Commands:**

   - `make docker-up` - Start Docker services
   - `make docker-down` - Stop Docker services
   - `make migrate-up` - Run database migrations
   - `make migrate-down` - Rollback migrations
   - `make test` - Run tests
   - `make build` - Build the application

## API Documentation

Postman Collection - see [go-genai-slack-assistant_postman.json](./docs/go-genai-slack-assistant_postman.json)

**Key Endpoints:**

- `POST /slack/events` - Slack webhook for events (requires signature verification)
- `GET /health` - Health check endpoint (returns database and Redis status)
- `GET /metrics` - Metrics endpoint (translation stats, cache hit rate, etc.)

## License

MIT - see [LICENSE](./LICENSE)

## Acknowledgments

- [Google Gemini API](https://ai.google.dev/) - AI-powered translation
- [lingua-go](https://github.com/pemistahl/lingua-go) - Fast offline language detection
- [Slack API](https://api.slack.com/) - Slack bot integration
- [Gin Web Framework](https://gin-gonic.com/) - High-performance HTTP routing
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Software design principles

Built with ❤️ using Go, Gin, Google Gemini AI, MySQL, Redis, and Docker