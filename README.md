# Slack Translation Bot

A Slack bot that automatically translates English messages to Vietnamese using Google Gemini AI. Built with Go, Clean Architecture, and containerized for easy deployment.

[![Release](https://img.shields.io/github/v/release/ntttrang/go-genai-slack-assistant?style=flat&logo=github&color=success)](https://github.com/ntttrang/go-genai-slack-assistant/releases)
[![Build Status](https://img.shields.io/jenkins/build?jobUrl=https://3dd68e143f08.ngrok-free.app/job/go-genai-slack-assistant&style=flat&logo=jenkins&label=Jenkins)](https://3dd68e143f08.ngrok-free.app/job/go-genai-slack-assistant/)
[![Docker Image](https://img.shields.io/docker/v/minhtrang2106/slack-bot?style=flat&logo=docker&label=Docker&color=2496ED)](https://hub.docker.com/r/minhtrang2106/slack-bot)
[![Docker Pulls](https://img.shields.io/docker/pulls/minhtrang2106/slack-bot?style=flat&logo=docker)](https://hub.docker.com/r/minhtrang2106/slack-bot)


![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-Framework-00ADD8?style=flat&logo=go&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-Database-4479A1?style=flat&logo=mysql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?style=flat&logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Containerized-2496ED?style=flat&logo=docker&logoColor=white)
[![Gemini](https://img.shields.io/badge/Google-Gemini_AI-4285F4?style=flat&logo=google&logoColor=white)](https://makersuite.google.com/app/apikey)

[![Demo Video](https://blog.n8n.io/content/images/size/w1200/2024/05/post-slack-bot3--1-.png)](https://www.youtube.com/watch?v=ihzPYzZtxRY)

## Features

- **Automatic Translation**: Translates messages between English and Vietnamese in Slack channels using Google Gemini AI
- **Smart Language Detection**: Offline language detection with lingua-go supporting 75+ languages for fast, accurate identification
- **Clean Architecture**: Follows Go standard layout with layered architecture for maintainability and testability

## Tech Stack

- **Language**: Go 1.24.0
- **Framework**: Gin (HTTP routing)
- **AI**: Google Generative AI (Gemini API)
- **Database**: MySQL with GORM ORM
- **Cache**: Redis
- **Language Detection**: lingua-go
- **Architecture**: Clean Architecture with layered structure
- **Logging**: Uber's zap (structured logging)
- **Testing**: testify + stretchr for assertions
- **Containerization**: Docker & Docker Compose
- **CI/CD**: Jenkins with automated releases to GitHub

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
├── tests/                   # Integration tests only
├── docs/                    # Documentation (*)
├── scripts/                 # Utility scripts (*)
├── database/                # Database migrations (*)
├── docker-compose.yml       # Docker services
├── Dockerfile               # Container image
└── Makefile                 # Build and deployment commands
```
(*): Optional

## Getting Started

### Prerequisites

- Go 1.24.0 (or compatible version)
- Docker & Docker Compose (or standalone MySQL + Redis)
- Slack workspace admin access
- Google Gemini API key (free tier available at [Google AI Studio](https://makersuite.google.com/app/apikey))
- (Optional) Jenkins for CI/CD and automated releases

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

   Create a `.env` file in the project root with your credentials:

   ```bash
   SLACK_BOT_TOKEN=xoxb-...
   SLACK_SIGNING_SECRET=...
   SLACK_APP_TOKEN=xapp-...
   GEMINI_API_KEY=AIza...
   DATABASE_HOST=localhost
   DATABASE_PORT=3306
   DATABASE_USER=root
   DATABASE_PASSWORD=...
   DATABASE_NAME=slack_bot
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

## Git Pre-Commit Hooks

This project includes automated pre-commit hooks to ensure code quality before commits. The hooks perform comprehensive validation checks to catch issues early.

### Installation

Install the git hooks by running:

```bash
./scripts/install-hooks.sh
```

This will install two hooks:
- **pre-commit**: Runs code quality and security checks on staged files
- **commit-msg**: Validates commit message format

### What Gets Checked

**Pre-Commit Hook (runs before each commit):**

1. ✅ **Debugging Statements** - Warns about `fmt.Println` and `println()` usage
2. ✅ **TODO/FIXME Comments** - Warns about unlinked TODO/FIXME comments
3. ✅ **Sensitive Data** - Blocks commits containing passwords, API keys, tokens, secrets
4. ✅ **Code Formatting** - Enforces `go fmt` standard
5. ✅ **Import Organization** - Checks `goimports` formatting (if installed)
6. ✅ **Static Analysis** - Runs `go vet` for potential issues
7. ✅ **Linting** - Runs `golangci-lint` for comprehensive checks (if installed)
8. ✅ **Module Tidiness** - Ensures `go.mod` and `go.sum` are tidy
9. ⚠️ **Unit Tests** - Optional (disabled by default for speed)

**Commit-Msg Hook (validates commit message format):**

Enforces [Conventional Commits](https://www.conventionalcommits.org/) format:

```
type(scope): message

Examples:
  feat(auth): add login functionality
  fix(api): handle nil pointer error
  docs(readme): update installation steps
```

**Allowed types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks
- `perf` - Performance improvements
- `ci` - CI/CD changes
- `build` - Build system changes
- `revert` - Revert previous commit

### Optional Tools

For full functionality, install these optional tools:

```bash
# Import organization
go install golang.org/x/tools/cmd/goimports@latest

# Comprehensive linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Customization

The hooks are located in:
- `scripts/hooks/pre-commit` - Pre-commit validation script
- `scripts/hooks/commit-msg` - Commit message validation script

To enable unit tests on every commit, edit `scripts/hooks/pre-commit` and uncomment lines 147-152.

### Bypassing Hooks (Not Recommended)

If you absolutely need to bypass the hooks (emergency situations only):

```bash
git commit --no-verify -m "your message"
```

**Note:** This skips all validation checks and should only be used in exceptional cases.

## API Documentation

Postman Collection - see [go-genai-slack-assistant_postman.json](./docs/go-genai-slack-assistant_postman.json)

**Key Endpoints:**

- `POST /slack/events` - Slack webhook for events (requires signature verification)
- `GET /health` - Health check endpoint (returns database and Redis status)
- `GET /metrics` - Metrics endpoint (translation stats, cache hit rate, etc.)

## CI/CD & Deployment

The project includes automated CI/CD pipeline using Jenkins with separated CI and CD stages:

**CI Pipeline (Runs on every Pull Request):**
1. **Code Quality**: Linting, testing, and code coverage
2. **Security Scanning**: Gosec, Govulncheck, and container vulnerability scanning (Trivy)
3. **Build**: Compile Go binary to verify buildability

**CD Pipeline (Runs after PR merge to develop/main):**
4. **Registry**: Build Docker image and push to Docker Hub
5. **Deployment**: 
   - Deploy to staging environment on Render (develop branch)
   - Deploy to production on Render (main branch)
   - Health checks and validation
6. **Release**: Automatic GitHub releases with version tags

**Key Features:**
- Pull requests must pass all CI stages before merge is allowed
- CD stages only execute after successful merge to protected branches
- GitHub branch protection rules enforce quality gates

see the [JENKINS PIPELINE](./docs/jenkins.jpg)

**Release Information:**

- Releases are automatically created on pushes to `main` branch
- Release version format: `v1.0.{BUILD_NUMBER}` (e.g., `v1.0.42`)
- Each release includes comprehensive notes with deployment URLs and health check information
- Docker images are available at [Docker Hub - minhtrang2106/slack-bot](https://hub.docker.com/r/minhtrang2106/slack-bot)

For detailed CI/CD setup, see [docs/QUALITY_GATE_SETUP.md](./docs/QUALITY_GATE_SETUP.md)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Google Gemini API](https://ai.google.dev/) - AI-powered translation
- [lingua-go](https://github.com/pemistahl/lingua-go) - Fast offline language detection
- [Slack API](https://api.slack.com/) - Slack bot integration
- [Gin Web Framework](https://gin-gonic.com/) - High-performance HTTP routing
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Software design principles

Built with ❤️ using Go, Gin, Google Gemini AI, MySQL, Redis, and Docker