# Changelog

All notable changes to the Aerostack CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.3] - 2026-02-24

### Fixed
- **CI Reliability**: Bumped version to v1.5.3 to resolve GitHub release asset conflicts and GoReleaser config errors.

## [1.5.2] - 2026-02-24

### Added
- **Feature Boilerplates**: Added comprehensive DB (D1), Cache (KV), AI Proxy, and Queue examples to all project templates (`blank`, `api`, `express`, `neon`).

### Fixed
- **AI Bindings**: Fixed a bug where `aerostack dev` was missing AI bindings in the generated `wrangler.toml`.
- **Template Reliability**: Improved SDK initialization and Queue support in starter templates.
- **Version Synchronization**: Synchronized CLI version across Go source, NPM package, and VERSION tracker.

## [1.5.1] - 2026-02-18

### Added
- **AI-First Foundation**: Introduced Project-Aware Knowledge Graph (PKG) for deep code understanding.
- **Agentic Core**: Native AI Agent with multi-turn tool use (read/write files, search symbols).
- **Recursive Self-Healing**: Automated error interception and TUI fix proposals for command failures.
- **Domain Modules**:
    - `aerostack auth doctor`: AI-powered authentication diagnostics.
    - `aerostack store schema`: Natural language SQL migration generation.
    - `aerostack ui sync`: Theme-aware context syncing for UI development.
- **Intelligent Deployment**: Added pre-flight checks and automated AI failure analysis for `aerostack deploy`.
- **Legacy Migration**: `aerostack migrate` command for seamless transition from `wrangler.toml` to `aerostack.toml`.

## [1.2.9] - 2026-02-16

### Added
- Auto-injection of `CACHE` (KV) and `QUEUE` bindings in `aerostack dev` if missing.
- Default `CACHE` and `QUEUE` bindings to all project templates (api, multi-func, neon, etc.).

### Fixed
- Monacopilot `KV cache not configured` error by enabling CLI support for KV namespaces.
- Logic Lab UI reliability: added error boundaries and retry logic for lazy-loaded modules.
- Logic Lab UI: improved "Custom API Endpoint" card with premium aesthetics and subdomain-based URLs.

## [1.2.8] - 2026-02-15
### Changed
- Improved CLI deployment stability.
- Core CLI scaffolding with Cobra
- `aerostack init` — Initialize new project from templates
- `aerostack dev` — Local development server (workerd)
- `aerostack deploy` — Deploy to Aerostack
- `aerostack login` — Authenticate with Aerostack
- `aerostack db` — Database commands (create, migrate, pull)
- Blank and API starter templates

### Changed
- (none yet)

### Fixed
- (none yet)

### Security
- (none yet)

---

## [0.1.0] - TBD

Initial public release.
