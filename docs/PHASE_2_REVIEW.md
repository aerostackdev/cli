# Phase 2 Code Review: Gaps, Bugs & Improvements

> Review of the Multi-DB Super Bridge implementation (Neon, migrations, typegen).  
> Based on `AEROSTACK_SDK_CLI_MASTER_VISION.md` Phase 2 scope.

---

## Phase 2 Scope (from Master Vision)

- **Neon Integration** — Connect to Neon API for Postgres-on-demand
- **Intelligent Routing** — Route queries between D1 and Postgres
- **Unified TypeGen** — `aerostack db:pull` for ALL connected databases
- **Edge Caching** — Automatic caching for DB client
- **Database Migrations** — Schema versioning, migration runner, rollback

---

## Current Implementation Summary

| Feature | Status | Location |
|---------|--------|----------|
| `db neon create <name>` | ✅ Implemented | `commands/db.go`, `neon/client.go` |
| `db migrate new <name>` | ✅ Implemented | `commands/db.go` |
| `db migrate new <name> --postgres` | ✅ Implemented | `commands/db.go` |
| `db migrate apply` | ✅ Implemented (D1 + Postgres) | `commands/db.go`, `postgres_migrate.go` |
| `db pull` | ✅ Implemented | `commands/db.go` |
| `generate types` | ✅ Implemented | `commands/generate.go`, `devserver/introspect.go` |
| Postgres/Hyperdrive in wrangler.toml | ✅ Implemented | `devserver/miniflare.go` |
| Intelligent routing | SDK-level (Phase 3+) | — |
| Edge caching | SDK-level (Phase 3+) | — |

---

## Bugs & Issues

### 1. Neon API Response Parsing (High)

**File:** `cli/internal/neon/client.go`

**Issue:** `CreateProjectResponse` expects `connection_uri` as a `ConnectionDetails` object, but the Neon API v2 returns `connection_uri` as a **string** (full URI), and connection details are nested under `connection_uris` or `branches[0].connection_uri`.

```go
// Current (likely wrong):
type CreateProjectResponse struct {
    Project    Project           `json:"project"`
    Connection ConnectionDetails `json:"connection_uri"`  // API returns string, not object
    Password   string            `json:"password"`
}
```

**Impact:** `CreateProject` will fail to decode the response; `BuildConnectionString` will receive empty/zero values.

**Fix:** Verify against [Neon API Create Project](https://api-docs.neon.tech/reference/createproject) response. The API may:
- Return `connection_uri` as a string (not an object)
- Require a separate GET `/projects/{project_id}/connection_uri` call
- Nest connection details under `branches` or `connection_uris`

Parse the actual response and either decode the URI string or use the correct nested structure.

---

### 2. D1 Introspection Requires wrangler.toml (High)

**File:** `cli/internal/devserver/introspect.go`, `cli/internal/commands/generate.go`

**Issue:** `IntrospectD1Local(dbName)` runs `wrangler d1 execute <dbName> --local` without ensuring `wrangler.toml` exists. Wrangler needs:
1. `wrangler.toml` in the current directory
2. A `[[d1_databases]]` block with matching `database_name`

`generate types` does **not** generate `wrangler.toml` before D1 introspection. If the user runs `aerostack generate types` without ever running `aerostack dev`, D1 introspection will fail.

**Fix:** In `generateTypes`, generate `wrangler.toml` from `aerostack.toml` before calling `IntrospectD1Local`, and run wrangler from the project root (`cmd.Dir`).

---

### 3. wrangler d1 execute Argument (Medium)

**File:** `cli/internal/devserver/introspect.go`

**Issue:** `wrangler d1 execute` expects the **database name** from the config. We pass `cfg.D1Databases[0].DatabaseName` (e.g. `"api-db"`). Wrangler uses this to look up the database in `wrangler.toml`. The exec is run without `cmd.Dir` set, so it may run from the wrong directory.

**Fix:** Set `cmd.Dir` to the project root (where `aerostack.toml` and `wrangler.toml` live).

---

### 4. Postgres connection_string Not Interpolated for generate types (Medium)

**File:** `cli/internal/devserver/miniflare.go`, `cli/internal/commands/generate.go`

**Issue:** `parsePostgresDatabases` calls `interpolateEnvVars(connStr)` when reading from `aerostack.toml`. If `connection_string = "$MY_DATABASE_URL"` and `MY_DATABASE_URL` is unset, the value stays as `"$MY_DATABASE_URL"`. `IntrospectPostgres` will then fail to connect. `ValidatePostgresConnectionString` correctly rejects strings containing `$`.

**Impact:** `generate types` will fail for Postgres when env vars are not loaded. No clear error message.

**Fix:** In `generateTypes`, explicitly check for unresolved `$` in connection strings and return a clear error: e.g. `"Postgres binding 'X' requires MY_DATABASE_URL to be set"`.

---

### 5. addPostgresToConfig Panic on Empty Name (Low)

**File:** `cli/internal/commands/db.go`

**Issue:** `addPostgresToConfig` uses `name[:1]` and `name[:2]`. If `name` is empty (edge case), this panics. Cobra requires one arg, but defensive checks are still useful.

```go
binding := strings.ToUpper(name[:1]) + strings.ToLower(name[1:])  // panic if name == ""
```

**Fix:** Add `if name == "" { return fmt.Errorf("name cannot be empty") }` at the start of `addPostgresToConfig`.

---

### 6. db migrate apply — FIXED ✅

**File:** `cli/internal/commands/db.go`

**Flow:** User runs `aerostack db migrate apply` → CLI ensures wrangler.toml exists → runs `npx wrangler d1 migrations apply <db> --local` (or `--remote`).

**Implementation:**
- Generates wrangler.toml from aerostack.toml if needed
- Uses `wrangler d1 migrations apply` (not `d1 execute --file`) — wrangler handles tracking, order, and rollback
- Supports `--remote` for staging/production
- Works with `migrations/*.sql` (timestamp or sequential naming)

**Remaining:** Postgres migration application.

---

### 7. Postgres Not in wrangler.toml (Medium)

**File:** `cli/internal/devserver/miniflare.go`

**Issue:** `GenerateWranglerToml` only emits D1 bindings. Postgres is configured in `aerostack.toml` but never written to `wrangler.toml`. Wrangler uses **Hyperdrive** for Postgres, not `connection_string` directly.

**Impact:** `aerostack dev` generates a wrangler.toml without Postgres. Workers cannot access Postgres bindings unless we add Hyperdrive config.

**Fix:** Add Hyperdrive bindings to `GenerateWranglerToml` when `PostgresDatabases` exist, or document that Postgres is handled separately (e.g. direct connection from Worker code).

---

### 8. db:pull vs generate types Naming (Low)

**File:** Vision doc vs `commands/generate.go`

**Issue:** Vision says `aerostack db:pull` for unified typegen. CLI implements `aerostack generate types`. No `db pull` subcommand.

**Fix:** Add `aerostack db pull` as an alias or primary command that delegates to the same logic as `generate types`, and/or update the vision to match.

---

### 9. Table Name Collisions Across D1 and Postgres (Medium)

**File:** `cli/internal/commands/generate.go`

**Issue:** `generateTypes` appends schemas from D1 and Postgres into `allSchemas` without namespacing. If both have a `users` table with different schemas, the last one wins and types will be wrong.

**Fix:** Prefix table names with the binding/database name (e.g. `DB_users`, `NeonDb_users`) or merge/union types with clear naming.

---

### 10. D1 parseTableNames Fragile (Low)

**File:** `cli/internal/devserver/introspect.go`

**Issue:** `parseTableNames` parses wrangler’s table output by splitting on `│` and assuming column positions. Wrangler’s output format can change.

**Fix:** Prefer `--json` if wrangler supports it, or add tests and fallbacks for format changes.

---

## Deferred to SDK / Later Phases

Items below are documented in **`SDK_AND_LATER_PHASES_BACKLOG.md`** with use cases and suggested priority.

| Item | Owner | Notes |
|------|-------|-------|
| **Intelligent routing** | SDK | Route reads→D1, writes→Postgres |
| **Edge caching** | SDK | Cache query results at edge |
| **Migration rollback** | CLI | `db migrate rollback` |
| **Neon region validation** | CLI | Validate `--region` before API call |
| **parseTableNames JSON** | CLI | Use wrangler `--json` for robust parsing |

*(Migration tracking and Hyperdrive are done. Postgres migrations implemented.)*

---

## Recommendations

### Priority 1 (Must Fix)

1. **Neon API response** — Fix `CreateProjectResponse` to match Neon API v2.
2. **generate types + D1** — Generate `wrangler.toml` before D1 introspection and set `cmd.Dir`.
3. **db migrate apply** — Implement real migration application for D1 (and later Postgres).

### Priority 2 (Should Fix)

4. **Postgres env vars** — Clear error when connection string env vars are missing.
5. **Postgres in wrangler** — Add Hyperdrive bindings or document the current limitation.
6. **Table name collisions** — Namespace or merge schemas from multiple DBs.

### Priority 3 (Nice to Have)

7. **addPostgresToConfig** — Add empty-name guard.
8. **db pull** — Alias for `generate types` to match vision.
9. **parseTableNames** — Use JSON output or harden parsing.

---

## Phase 2 Completion Status (Updated)

| Item | Status |
|------|--------|
| Neon API response parsing | ✅ Fixed |
| generate types + D1 (wrangler.toml, cmd.Dir) | ✅ Fixed |
| db migrate apply (D1 via wrangler migrations apply) | ✅ Implemented |
| db migrate apply (Postgres) | ✅ Implemented (migrations_postgres/) |
| Postgres/Hyperdrive in wrangler.toml | ✅ Implemented |
| Table name collisions (namespace) | ✅ Fixed |
| db pull alias | ✅ Added |
| Postgres env var error | ✅ Fixed |
| addPostgresToConfig empty guard | ✅ Fixed |

## Files Touched by Phase 2

| File | Purpose |
|-----|---------|
| `cli/internal/commands/db.go` | `db neon create`, `db migrate new/apply`, `db pull` |
| `cli/internal/commands/generate.go` | `generate types` |
| `cli/internal/commands/dev.go` | Hyperdrive env injection |
| `cli/internal/neon/client.go` | Neon API client |
| `cli/internal/devserver/introspect.go` | D1/Postgres introspection, TypeScript generation (namespaced) |
| `cli/internal/devserver/miniflare.go` | Parse aerostack.toml, generate wrangler.toml (with Hyperdrive) |
| `cli/internal/devserver/postgres.go` | Postgres connection validation |
| `cli/internal/devserver/postgres_migrate.go` | Postgres migration application |
