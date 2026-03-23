# Aerostack CLI

[![Go Version](https://img.shields.io/github/go-mod/go-version/aerostackdev/sdks)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

The official CLI for building, testing, and deploying Aerostack applications. Scaffold projects from templates, run a full local development environment, manage databases, and deploy to Aerostack Cloud — all from a single command-line tool.

## Features

- **Zero-Config Local Dev** — Start building immediately with `aerostack dev`, powered by an embedded workerd runtime
- **Multi-DB Orchestration** — Unified interface for D1, Neon Postgres, and external databases
- **Built-in Testing** — Test framework with service mocking and isolated environments
- **AI-Powered Fixes** — Automatic error detection and fix suggestions via integrated LLM agent
- **20+ Starter Templates** — Production-ready templates for APIs, SaaS, AI agents, and more
- **One-Command Deploy** — Ship to staging or production with `aerostack deploy`

## Installation

**curl (recommended)**

```bash
curl -fsSL https://get.aerostack.dev | sh
```

**npm / pnpm / yarn / bun**

```bash
npm install -g @aerostack/cli
# or
pnpm add -g @aerostack/cli
# or
yarn global add @aerostack/cli
# or
bun install -g @aerostack/cli
```

**npx (no install)**

```bash
npx aerostack init my-app
```

## Quick Start

```bash
# 1. Create a new project from a template
aerostack init my-app --template=blog

# 2. Start local development
cd my-app
aerostack dev

# 3. Deploy to production
aerostack deploy
```

## Commands

### Project Management

| Command | Description |
|---------|-------------|
| `aerostack init [name]` | Create a new project (interactive template picker) |
| `aerostack dev` | Start local dev server with embedded workerd, D1, and hot reload |
| `aerostack deploy` | Deploy to Aerostack Cloud (staging or production) |
| `aerostack link` | Link an existing local project to an Aerostack remote project |

### Database

| Command | Description |
|---------|-------------|
| `aerostack db create [name]` | Create a new D1 or Postgres database |
| `aerostack db migrate` | Run pending migrations |
| `aerostack db pull` | Introspect database and generate TypeScript types |

### Authentication

| Command | Description |
|---------|-------------|
| `aerostack login` | Authenticate with your Aerostack account |
| `aerostack whoami` | Display the currently logged-in user |

### Resources & Services

| Command | Description |
|---------|-------------|
| `aerostack functions` | Manage serverless functions |
| `aerostack secrets` | Manage project secrets and environment variables |
| `aerostack resources` | List and manage provisioned resources |
| `aerostack store` | Initialize and manage data stores |

### Advanced

| Command | Description |
|---------|-------------|
| `aerostack ai` | Interactive AI assistant for troubleshooting |
| `aerostack generate` | Code generation from templates and schemas |
| `aerostack migrate` | Migrate from Wrangler/Workers projects to Aerostack |
| `aerostack skill` | Run predefined project skills |

## Configuration

Projects are configured via `aerostack.toml` in the project root:

```toml
name = "my-app"
main = "src/index.ts"
compatibility_date = "2024-12-01"

[vars]
AEROSTACK_API_URL = "https://api.aerostack.dev"
DATABASE_NAME = "my-db"

[d1_databases]
binding = "DB"
database_name = "my-db"
```

## Architecture

| Component | Technology | Purpose |
|-----------|-----------|---------|
| CLI Framework | Go + Cobra | Fast, cross-platform binary |
| Local Runtime | workerd (embedded) | Cloudflare Workers compatibility |
| TUI | Charm (bubbletea, huh, lipgloss) | Rich terminal interface |
| Build | GoReleaser | Automated multi-platform releases |

## Development

### Prerequisites

- Go 1.24+
- Node.js (for workerd)

### Build & Run

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/aerostack/main.go

# Build binary
go build -o bin/aerostack cmd/aerostack/main.go

# Run tests
go test ./...

# Test release build (no publish)
goreleaser release --snapshot
```

### Project Structure

```
cli/
├── cmd/aerostack/          # CLI entry point
├── internal/
│   ├── commands/           # All command implementations
│   ├── agent/              # AI-powered diagnostics agent
│   ├── devserver/          # Local dev server (workerd + Miniflare)
│   ├── credentials/        # Credential storage and encryption
│   ├── modules/
│   │   ├── auth/           # Auth diagnostics
│   │   ├── deploy/         # Deployment agent
│   │   ├── mcpconvert/     # MCP conversion pipeline
│   │   ├── migration/      # Wrangler-to-Aerostack migration
│   │   ├── store/          # Data store initialization
│   │   └── ui/             # UI synchronization
│   ├── neon/               # Neon Postgres client
│   ├── pkg/                # Shared utilities
│   ├── provision/          # Resource provisioning
│   ├── selfheal/           # Self-healing diagnostics TUI
│   └── templates/          # Project template management
└── scripts/                # Build scripts
```

## Release

Releases are automated via GoReleaser in CI (`.github/workflows/release-cli.yml`). The version is read from the `VERSION_CLI` file. Do not manually publish binaries.

## Links

- **Repository:** [aerostackdev/sdks](https://github.com/aerostackdev/sdks)
- **Releases:** [GitHub Releases](https://github.com/aerostackdev/sdks/releases)
- **Contributing:** See [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Release Policy:** See [RELEASE.md](./RELEASE.md)

## License

MIT
