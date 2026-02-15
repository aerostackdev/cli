# Phase 2 Test Plan

Run the automated test script:
```bash
cd cli && chmod +x scripts/phase2-test.sh && ./scripts/phase2-test.sh
```

## Manual Test Checklist

### 1. D1 Migrations
- [ ] `aerostack init` creates aerostack.toml
- [ ] `aerostack db migrate new add_users` creates migrations/TIMESTAMP_add_users.sql
- [ ] `aerostack db migrate new add_posts --postgres` creates migrations_postgres/TIMESTAMP_add_posts.sql
- [ ] `aerostack db migrate apply` runs wrangler d1 migrations apply (local)
- [ ] `aerostack db migrate apply --remote staging` runs with --remote

### 2. Postgres Migrations
- [ ] Add `[[postgres_databases]]` with connection_string to aerostack.toml
- [ ] Create migrations_postgres/0001_initial.sql
- [ ] `aerostack db migrate apply` applies to Postgres (local only)
- [ ] _aerostack_migrations table tracks applied migrations

### 3. Generate Types
- [ ] `aerostack generate types` introspects D1 and Postgres
- [ ] `aerostack db pull` produces same output (alias)
- [ ] Tables from different DBs are namespaced (DBUsers vs PgDbUsers)
- [ ] Fails with clear error when Postgres env vars unset

### 4. Dev Server
- [ ] `aerostack dev` generates wrangler.toml with D1 + Hyperdrive
- [ ] Postgres bindings get CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_* env
- [ ] Server starts on port 8787

### 5. Neon
- [ ] `aerostack db neon create my-db` (requires NEON_API_KEY)
- [ ] Returns connection string
- [ ] Adds to aerostack.toml with --add-to-config

## Prerequisites
- Go 1.21+
- Node.js 18+ (for wrangler)
- NEON_API_KEY (for db neon create)
- Postgres connection string (for Postgres tests)
