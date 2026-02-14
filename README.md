# Aerostack CLI

> **Status**: Early Development - Core scaffolding in progress

The official CLI for building, testing, and deploying Aerostack applications.

## Features

- 🚀 **Zero-Config Local Dev** - Start building immediately with `aerostack dev`
- 📦 **Multi-DB Orchestration** - Unified interface for D1, Neon, and external databases
- 🧪 **Built-in Testing** - Test framework with service mocking
- 🤖 **AI-Powered Fixes** - Automatic error detection and fix suggestions
- 🎨 **20+ Starter Templates** - Production-ready templates for common use cases

## Installation

**curl (recommended)**
```bash
curl -fsSL https://get.aerostack.dev | sh
```

**npm / pnpm / yarn / bun**
```bash
npm install -g aerostack
# or: pnpm add -g aerostack
# or: yarn global add aerostack
# or: bun install -g aerostack
```

**npx (no install)**
```bash
npx aerostack init my-app
```

## Quick Start

```bash
# Initialize a new project
aerostack init my-app --template=blog

# Start local development
cd my-app
aerostack dev

# Deploy to staging
aerostack deploy --env staging
```

## Commands

### Project Management
- `aerostack init [name]` - Initialize a new project
- `aerostack dev` - Start local development server (D1, one config, see [DEV_DIFFERENTIATION.md](docs/DEV_DIFFERENTIATION.md))
- `aerostack deploy` - Deploy to Aerostack cloud

### Database
- `aerostack db create [name]` - Create a new database
- `aerostack db migrate` - Run migrations
- `aerostack db pull` - Generate TypeScript types

### Authentication
- `aerostack login` - Authenticate with Aerostack

## Architecture

This CLI is built with:
- **Go** for fast, cross-platform binary distribution
- **Cobra** for command-line interface
- **workerd** (embedded) for local Cloudflare Workers runtime

## Development

```bash
# Install dependencies
go mod download

# Run locally
go run cmd/aerostack/main.go

# Build binary
go build -o bin/aerostack cmd/aerostack/main.go
```

## Project Structure

```
cli/
├── cmd/aerostack/          # CLI entry point
├── internal/
│   ├── commands/           # Command implementations
│   ├── devserver/          # Local dev server (workerd wrapper)
│   ├── templates/          # Template engine
│   └── config/             # Configuration management
├── pkg/api/                # Aerostack platform API client
├── templates/              # Starter templates
└── scripts/                # Build scripts
```

## Repository

- **GitHub:** [aerostackdev/cli](https://github.com/aerostackdev/cli)
- **Releases:** [Releases](https://github.com/aerostackdev/cli/releases)
- **Contributing:** See [CONTRIBUTING.md](./CONTRIBUTING.md)
- **Release policy:** See [RELEASE.md](./RELEASE.md)

## License

TBD
