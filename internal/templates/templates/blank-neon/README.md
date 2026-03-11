## Aerostack Blank Neon Template

This project was generated from the **blank-neon** template. It is a minimal Cloudflare Worker pre-wired to Neon PostgreSQL, KV cache, and the AI proxy — a clean starting point for Postgres-backed Workers.

### What's Included

- `GET /test/db` — Runs `SELECT NOW()` against Neon to verify the connection.
- `GET /test/cache` — Stores and retrieves a value from KV.
- `GET /test/ai` — Calls the Aerostack AI proxy.
- Pre-configured Neon (`PG`), KV (`CACHE`), and Queue bindings in `aerostack.toml`.

### 1. Set Up Neon

```bash
# Using the Aerostack CLI:
aerostack db neon create my-neon-db --add-to-config

# Or create manually at https://console.neon.tech, then set DATABASE_URL in .dev.vars
```

### 2. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` — Your project ID from the Aerostack dashboard.
- `AEROSTACK_API_KEY` — API key for this project (keep this secret).
- `DATABASE_URL` — Your Neon connection string (`postgres://user:pass@host/dbname`).

> **Important**: `DATABASE_URL` must also be set as a production secret before deploying:
> ```bash
> aerostack secrets set DATABASE_URL "your-neon-url" --env production
> ```

### 3. Run Locally

```bash
aerostack dev
```

Test the endpoints:

```bash
# Verify Neon connection
curl http://localhost:8787/test/db

# Cache test
curl http://localhost:8787/test/cache

# AI test (requires valid credentials)
curl http://localhost:8787/test/ai
```

### 4. Deploy

**First time only — link to your Aerostack project:**

```bash
aerostack link --write-toml
```

This picks your project interactively and writes `project_id` into `aerostack.toml` so every future deploy goes to the right project automatically.

```bash
aerostack deploy
```

> **Reminder**: Set your Neon connection string as a production secret **before** deploying, or use `--sync-secrets` to push it automatically from `.dev.vars`:
> ```bash
> aerostack deploy --sync-secrets --env production
> ```

### Using a Non-Neon Postgres Database

`sdk.db` works with **any Postgres-compatible database**, not just Neon. Set `DATABASE_URL` to any Postgres connection string in `.dev.vars`:

```
DATABASE_URL=postgresql://user:pass@your-host:5432/dbname
```

**Supported**: Neon, Supabase, Railway, Render, Fly.io Postgres, AWS RDS, GCP Cloud SQL, CockroachDB
**Not supported**: MySQL, MongoDB, Turso, or any non-Postgres wire protocol

> **Production note**: Non-Neon databases require [Cloudflare Hyperdrive](https://developers.cloudflare.com/hyperdrive/) since Workers can't open raw TCP connections. Neon uses HTTP natively — no Hyperdrive needed.

### More Documentation

- Online docs: `https://docs.aerostack.dev`
