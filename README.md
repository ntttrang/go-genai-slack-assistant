# Slack Translation Bot

A Slack bot that automatically translates English messages to Vietnamese using AI.

## What It Does

The bot listens to Slack channels and translates English messages to Vietnamese, posting translations as thread replies. It uses Google Gemini API for translation and detects languages offline for fast processing.

## Tech Stack

- **Language**: Go 1.21+
- **AI**: Google Gemini API (free tier)
- **Database**: MySQL
- **Cache**: Redis
- **Architecture**: Clean Architecture pattern

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed design.

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Slack workspace admin access
- Google Gemini API key

### Setup

1. **Create Slack App**
   - Go to [api.slack.com/apps](https://api.slack.com/apps)
   - Add bot scopes: `channels:history`, `channels:read`, `chat:write`, `reactions:read`
   - Enable Socket Mode
   - Install to workspace

2. **Get Gemini API Key**
   - Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
   - Generate free API key

3. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your tokens
   ```

4. **Run**
   ```bash
   docker-compose up -d        # Start MySQL & Redis
   make migrate-up             # Setup database
   go run cmd/api/main.go      # Start bot
   ```

## Usage

- **Auto-translate**: Enable for specific channels via configuration
- **Manual**: React with 🇻🇳 emoji to any message

## Development

```bash
make test           # Run tests
make lint           # Run linter
make build          # Build binary
```

Follow [GOLANG_BEST_PRACTICES.md](./GOLANG_BEST_PRACTICES.md) for code style.

## Roadmap

- ✅ Architecture design
- ⏳ Core translation (MVP)
- ⏳ Slack integration
- ⏳ Self-hosted AI option (Ollama)

## License

MIT - see [LICENSE](./LICENSE)
