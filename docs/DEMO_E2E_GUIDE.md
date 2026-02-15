# Aerostack CLI + SDK — End-to-End Demo Guide

Run a full demo: create project → dev → test → deploy.

---

## Quick Start (Automated)

```bash
# From repo root
./cli/scripts/demo-e2e.sh
```

This will:
1. Build the CLI
2. Create `demo-aerostack/` with api template
3. Run `aerostack test`
4. Start `aerostack dev` briefly (5 sec)
5. Skip deploy (see below for manual deploy)

---

## Manual Demo (Step by Step)

### 1. Build CLI

```bash
cd cli
go build -o bin/aerostack ./cmd/aerostack
export PATH="$PWD/bin:$PATH"
```

### 2. Create Demo Project

```bash
aerostack init demo-api --template=api
cd demo-api
```

### 3. Run Tests

```bash
aerostack test
```

### 4. Start Dev Server

```bash
aerostack dev
# Visit http://localhost:8787
# GET /         → "Welcome to the Aerostack API!"
# GET /users/42 → {"id":"42","name":"User 42","role":"developer"}
```

### 5. Deploy to Cloudflare

**Prerequisites:**
- Cloudflare account
- Wrangler logged in

**Steps:**

```bash
# 1. Login to Cloudflare
npx wrangler login

# 2. Create a D1 database (one-time)
npx wrangler d1 create demo-api-db
# Copy the database_id from output

# 3. Update aerostack.toml — replace YOUR_STAGING_D1_ID with your database_id
# [[env.staging.d1_databases]]
# binding = "DB"
# database_name = "api-db"
# database_id = "<paste-your-d1-id-here>"

# 4. Deploy
aerostack deploy --env staging
```

After deploy, Wrangler will output your Worker URL (e.g. `https://demo-api.<subdomain>.workers.dev`).

---

## Demo with Deploy (Automated)

To run the full demo including deploy attempt:

```bash
# 1. Login first
npx wrangler login

# 2. Create D1 and get ID
npx wrangler d1 create demo-aerostack-db

# 3. Update aerostack.toml in demo project (after script runs) with the database_id

# 4. Run demo with deploy
DEPLOY=1 ./cli/scripts/demo-e2e.sh
```

Or run deploy manually after the script:

```bash
cd demo-aerostack
# Edit aerostack.toml with your D1 database_id
aerostack deploy --env staging
```

---

## Templates to Try

| Template | Command | Use Case |
|----------|---------|----------|
| **blank** | `aerostack init x --template=blank` | Minimal Worker |
| **api** | `aerostack init x --template=api` | REST API with Hono |
| **express** | `aerostack init x --template=express` | Express.js on Workers |

---

## Full Feature Checklist

- [ ] `aerostack init` — project created
- [ ] `aerostack dev` — dev server runs, D1 bound
- [ ] `aerostack test` — Vitest passes
- [ ] `aerostack add lib auth` — shared module created
- [ ] `aerostack add function api-gateway` — service added (multi-worker)
- [ ] `aerostack secrets list` — (requires login)
- [ ] `aerostack deploy --env staging` — deploys to Cloudflare
- [ ] `aerostack deploy --env production` — deploys to production

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `aerostack.toml not found` | Run from project root (where aerostack.toml is) |
| `wrangler secret list failed` | Run `npx wrangler login` |
| Deploy fails with D1 error | Ensure `database_id` in aerostack.toml matches your D1 database |
| Tests fail (hono not found) | Run `npm install` in project |
| Dev fails (esbuild) | Ensure Node 18+ and `npm install` was run |
