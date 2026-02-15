# Phase 3: Modular Workspace & Tracing — Plan

> **Goal:** Scaling to microservices. Make Aerostack the best platform for building multi-service edge apps.

---

## Vision Summary

Phase 3 turns Aerostack from "single Worker + DB" into **multi-service with shared logic, tracing, and production-ready deploy**. This is where Aerostack becomes meaningfully better than raw Wrangler.

---

## Phase 3 Scope (from Master Vision)

| Item | Description |
|------|-------------|
| **Shared Workspace** | `@shared/*` logic with automated tree-shaking |
| **Distributed Tracing** | Cross-service log stitching, trace IDs |
| **Framework Adapters** | Magic Wrapper for Express, Fastify, NestJS |
| **Multi-Environment** | `deploy --env staging` |
| **Testing Framework** | `aerostack test` with service mocking |
| **Secret Management** | Production-grade (Cloudflare encrypted vault) |

---

## Proposed Priorities (What Makes the Platform Best)

### Tier 1: Foundation (Must Have)

| # | Feature | Why It Matters | Effort |
|---|---------|----------------|--------|
| 1 | **Multi-Environment Deploy** | `deploy --env staging` / `--env production`. No staging = no safe production path. | Medium |
| 2 | **Shared Workspace** | `@shared/*` alias + tree-shaking. Without it, multi-service = copy-paste hell. Core to "monolith-to-functions". | High |
| 3 | **Distributed Tracing** | One terminal, all service logs, color-coded + trace IDs. Debugging 5 services without this is painful. | Medium |

### Tier 2: DX Multipliers

| # | Feature | Why It Matters | Effort |
|---|---------|----------------|--------|
| 4 | **Multi-Worker Dev** | `aerostack dev` spins up N workers when project has N services. Today: single worker only. | High |
| 5 | **Testing Framework** | `aerostack test` — run tests with D1/KV/Postgres mocked. Confidence before deploy. | Medium |
| 6 | **Secret Management** | `.env` local, encrypted vault for staging/prod. No secrets in code. | Medium |

### Tier 3: Reach & Polish

| # | Feature | Why It Matters | Effort |
|---|---------|----------------|--------|
| 7 | **Framework Adapters** | Express/Fastify/NestJS users can adopt without rewriting. Expands audience. | Medium |
| 8 | **Workspace Commands** | `aerostack add function <name>`, `aerostack add lib <name>`. Scaffolding for multi-service. | Low |

---

## Detailed Breakdown

### 1. Multi-Environment Deploy

**What:** `aerostack deploy --env staging` deploys to staging; `--env production` to prod.

**Current state:** `deploy` exists but likely single-target.

**Tasks:**
- [ ] Parse `[env.staging]`, `[env.production]` from aerostack.toml
- [ ] Map env to wrangler `--env` or separate wrangler configs
- [ ] Deploy to correct D1/KV/R2 IDs per environment
- [ ] Output deploy URL per env

**Dependencies:** None. Can ship first.

---

### 2. Shared Workspace (`@shared/*`)

**What:** `import { db } from "@shared/db"` resolves to `shared/db.ts`. Each deployed Worker gets only the shared code it imports (tree-shaking).

**Current state:** `bundler.go` uses esbuild; no `@shared` alias.

**Tasks:**
- [ ] Add esbuild alias: `@shared` → `./shared`
- [ ] Ensure tree-shaking works (esbuild does this by default)
- [ ] Document `shared/` in init templates
- [ ] `aerostack add lib <name>` creates `shared/<name>.ts`

**Dependencies:** None. Single-service projects benefit immediately.

---

### 3. Distributed Tracing

**What:** When Service A calls Service B, both logs show the same trace ID. One terminal, color-coded by service.

**Current state:** `aerostack dev` runs single wrangler; no multi-service, no trace stitching.

**Tasks:**
- [ ] Inject trace ID into request context (e.g. `X-Trace-ID` header or env)
- [ ] SDK/CLI: generate and propagate trace ID
- [ ] When multi-worker dev exists: aggregate logs, prefix with `[service-name]`, include trace ID
- [ ] Optional: structured JSON logs for tooling

**Dependencies:** Multi-Worker Dev (for full value). Can start with trace ID injection for single-worker.

---

### 4. Multi-Worker Dev

**What:** Project with 5 services → `aerostack dev` runs 5 wrangler instances. Services find each other via localhost RPC.

**Current state:** Single main entry, single wrangler dev.

**Tasks:**
- [ ] Define multi-service structure in aerostack.toml (e.g. `[[services]]` or `main` as array)
- [ ] CLI spawns N wrangler processes (or use wrangler's multi-worker mode if available)
- [ ] Local service bindings: Service A → localhost:8788, Service B → localhost:8789
- [ ] Unified log aggregation (combine stdout from all workers)

**Dependencies:** Shared Workspace (so services can share code). Complex.

---

### 5. Testing Framework

**What:** `aerostack test` runs Vitest/Jest with D1, KV, Postgres available as test fixtures.

**Tasks:**
- [ ] Add `aerostack test` command
- [ ] Integrate Vitest
- [ ] Provide test helpers: `getTestDB()`, `getTestKV()` — in-memory or local SQLite
- [ ] Optional: `--coverage`

**Dependencies:** None. High value standalone.

---

### 6. Secret Management

**What:** Local: `.env` / `.dev.vars`. Staging/Prod: Cloudflare Workers Secrets or encrypted vault.

**Tasks:**
- [ ] `aerostack secrets set KEY value` — set secret for env
- [ ] `aerostack secrets list` — list (names only, not values)
- [ ] Load `.dev.vars` in dev; inject into wrangler
- [ ] Deploy: push secrets to Cloudflare before deploy

**Dependencies:** Multi-Environment Deploy (secrets are per-env).

---

### 7. Framework Adapters

**What:** Wrap Express/Fastify/NestJS app so it runs on Workers. User writes `app.listen()`-style code; CLI adapts to `fetch`.

**Priority order:** Express → Fastify → NestJS (if possible)

**Tasks:**
- [ ] **Express:** Detect in package.json, use Worker adapter (e.g. `@cloudflare/workers-express` or similar), wrap in `export default { fetch }`
- [ ] **Fastify:** Same pattern; Fastify is edge-friendly, may be easier
- [ ] **NestJS:** If feasible — more complex (DI, modules); evaluate effort
- [ ] Document migration path for each

**Dependencies:** None. Expands audience.

---

### 8. Workspace Commands

**What:** `aerostack add function api-gateway`, `aerostack add lib auth`.

**Tasks:**
- [ ] `aerostack add function <name>` — create `services/<name>/index.ts`, update aerostack.toml
- [ ] `aerostack add lib <name>` — create `shared/<name>.ts`
- [ ] Templates for each

**Dependencies:** None. Quick win.

---

## Recommended Implementation Order

```
Week 1: Foundation
├── 1. Multi-Environment Deploy
├── 2. Shared Workspace (@shared + tree-shaking)
└── 8. Workspace Commands (add function, add lib)

Week 2: Observability & Quality
├── 3. Distributed Tracing (trace ID injection first)
├── 5. Testing Framework
└── 6. Secret Management

Week 3: Scale
├── 4. Multi-Worker Dev
└── 7. Framework Adapters (if time)
```

---

## Success Metrics

| Metric | Target |
|--------|--------|
| **Multi-service project** | User can run 3+ services locally with shared code |
| **Deploy** | User can deploy to staging and production with one command each |
| **Debugging** | User can trace a request across services via trace ID |
| **Testing** | User can run `aerostack test` with DB/KV available |
| **Secrets** | User never commits secrets; local and prod both work |

---

## Decisions

| Decision | Choice | Notes |
|----------|--------|-------|
| **Deploy target** | Aerostack-managed infra (for now) | Later: add option for user's own Cloudflare account. Both paths supported. |
| **Testing framework** | Vitest | ESM-native, fast, TypeScript-friendly, good fit for Workers/edge. |
| **Framework adapters** | Express first, then Fastify, NestJS if possible | Express = largest install base; Fastify = edge-friendly; NestJS = enterprise adoption. |

---

## Open Questions

- [ ] **Multi-worker:** Wrangler supports multiple workers in one process? Or spawn N processes?

---

## References

- Master Vision: `docs/planning/AEROSTACK_SDK_CLI_MASTER_VISION.md`
- Phase 2: `cli/docs/PHASE_2_REVIEW.md`
- DEV_DIFFERENTIATION: `cli/docs/DEV_DIFFERENTIATION.md`
