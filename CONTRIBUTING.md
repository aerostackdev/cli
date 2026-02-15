# Contributing to Aerostack CLI

Thank you for contributing to the Aerostack CLI. This document outlines our workflow and conventions.

---

## Development Setup

```bash
# Clone the repo
git clone https://github.com/aerostackdev/cli.git
cd cli

# Install dependencies
go mod download

# Run locally
go run cmd/aerostack/main.go --help

# Build
go build -o bin/aerostack cmd/aerostack/main.go
```

---

## Commit Phases (Feature-by-Feature)

We commit in **small, logical phases**. Each phase should be a single PR when possible.

### Phase 1: Core scaffolding ✅
- CLI entry point, Cobra setup
- `init`, `dev`, `deploy`, `login`, `db` commands (stubs or basic impl)

### Phase 2: Local development
- `aerostack dev` with workerd
- Template engine, bundler integration
- Hot reload (if applicable)

### Phase 3: Deployment
- `aerostack deploy` integration
- Cloudflare Workers / Aerostack platform API

### Phase 4: Database tooling
- `aerostack db create`, `db migrate`, `db pull`
- D1, Neon, external DB support

### Phase 5: Polish
- Error handling, logging
- Install script, Homebrew
- Documentation

---

## Commit Message Convention

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(scope): add hot reload
fix(dev): resolve port conflict
docs: update install instructions
```

See [RELEASE.md](./RELEASE.md) for full details.

---

## Pull Request Process

1. **Branch** from `main`: `git checkout -b feature/your-feature`
2. **Commit** with conventional messages
3. **Push** and open a PR
4. **CI** must pass (lint, test, build)
5. **Review** — at least one approval
6. **Merge** — squash or merge, per repo settings

---

## Code Style

- **Go:** Follow [Effective Go](https://go.dev/doc/effective_go) and `gofmt`
- **Tests:** Place tests next to code (`*_test.go`)
- **Comments:** Document exported functions and types

---

## Questions?

Open an [issue](https://github.com/aerostackdev/cli/issues) or reach out to the team.
