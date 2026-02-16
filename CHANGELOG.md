# Changelog

All notable changes to the Aerostack CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
