# SDK & Later Phases Backlog

> Items deferred from Phase 2 — for SDK implementation or future CLI phases.

---

## SDK-Level (Not CLI)

These require `@aerostack/sdk` or runtime logic, not CLI commands.

### Intelligent routing (D1 vs Postgres)

| Field | Value |
|-------|-------|
| **What it does** | SDK automatically routes reads to D1 (edge) and writes to Postgres (primary) |
| **Use case** | Single `app.db.query()` API; SDK picks the right DB per operation |
| **When to build** | When you have both D1 and Postgres and want unified read/write optimization |
| **Effort** | High |

### Edge caching

| Field | Value |
|-------|-------|
| **What it does** | Caches query results at the edge (e.g. Cache API) with TTL |
| **Use case** | High-traffic reads (product catalog, blog) served from cache; fewer DB hits |
| **When to build** | Read-heavy apps needing lower latency and reduced DB load |
| **Effort** | High |

---

## CLI — Later Phase

These are CLI improvements deferred to a future phase.

### Migration rollback

| Field | Value |
|-------|-------|
| **What it does** | `aerostack db migrate rollback` undoes the last applied migration |
| **Use case** | Bad migration in production → roll back instead of writing a fix migration |
| **When to build** | When production migrations need safer rollback |
| **Effort** | Medium |
| **Notes** | D1: wrangler has no built-in rollback. Postgres: use `_aerostack_migrations` + down-migrations |

### Neon region validation

| Field | Value |
|-------|-------|
| **What it does** | Validates `--region` against Neon’s supported regions before API call |
| **Use case** | Catch typos (e.g. `us-wst-2`) and show valid options |
| **When to build** | Quick UX improvement |
| **Effort** | Low |

### parseTableNames JSON output

| Field | Value |
|-------|-------|
| **What it does** | Use `wrangler d1 execute --json` instead of parsing table output |
| **Use case** | Robust D1 introspection if wrangler changes output format |
| **When to build** | When D1 introspection breaks or for more reliable parsing |
| **Effort** | Low |

---

## Suggested Priority

| Order | Item | Effort | Impact |
|-------|------|--------|--------|
| 1 | Neon region validation | Low | Fewer API errors |
| 2 | Migration rollback | Medium | Safer production deploys |
| 3 | parseTableNames JSON | Low | More robust introspection |
| 4 | Intelligent routing | High | SDK feature |
| 5 | Edge caching | High | SDK feature |

---

## Reference

- Phase 2 scope: `docs/planning/AEROSTACK_SDK_CLI_MASTER_VISION.md`
- Phase 2 review: `cli/docs/PHASE_2_REVIEW.md`
