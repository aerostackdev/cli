# Changelog

All notable changes to the Aerostack CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.13] - 2026-02-25

### Fixed
- **Release Automation**: Triggered a clean release build for NPM and GitHub Actions to distribute the ES module and socket dispatcher fixes.

## [1.5.12] - 2026-02-25

### Fixed
- **ES Module Build**: Fixed an issue in `deploy.go` where spaces in esbuild arguments (specifically the `--banner` flag) caused the deployment bundle to be invalid or produce a 10021 "Main module must be an ES module" error. Also ensured the `dist` directory is cleaned before every build.

## [1.5.11] - 2026-02-25

### Fixed
- **Socket Error**: Resolved "Service dispatcher not configured" error by supporting `AEROSTACK_API_URL` and missing bindings.
- **Deployment**: Fixed "Queue not found" error during deployment by filtering out local stub bindings.

## [1.5.8] - 2026-02-25

### Added
- **Node.js Compatibility Bridge**: Introduced `.cjs` mock extension and `createRequire` shim for reliable CommonJS support in Workers.
- **Global Prototype Fix**: Added automatic `esbuild` banner to patch missing methods (like `hasOwnProperty`) on native Worker modules.

### Changed
- **Project Templates**: Updated all project templates to use `nodejs_compat_v2` for modern Node.js support.
- **Build Optimization**: Switched to relative `./` paths for alias resolution, ensuring robust builds across different environments.

## [1.5.7] - 2026-02-25

### Fixed
- **CI/CD**: Enabled artifact replacement in GoReleaser to prevent release failures due to asset conflicts.

## [1.5.6] - 2026-02-25

### Added
- **CLI Error Details**: Added ability to view detailed CLI error logs and telemetry in the Admin UI.

## [1.5.5] - 2026-02-24

### Fixed
- **Express Deployment**: Fixed issue where `nodejs_compat_v2` flag was not correctly passed during deployment.
- **Express Bundling**: Resolved bundling errors in Express templates by optimizing build configuration.

## [1.5.4] - 2026-02-24

### Fixed
- **Windows Build/Run**: Resolved `syscall.Kill` errors on Windows by abstracting process termination logic.

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
