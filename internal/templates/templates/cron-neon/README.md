## Aerostack Cron + Neon Template

This project was generated from the **cron-neon** template. It runs a scheduled Cloudflare Worker (cron trigger) that writes to a Neon PostgreSQL database on every tick.

### What's Included

- `scheduled` handler that runs on a cron schedule (configured in `aerostack.toml`).
- Writes an execution log to Neon (`sdk.db`) and updates a KV status key (`sdk.cache`).
- Optional `fetch` handler for a health check or manual trigger during development.

### 1. Set Up Neon

Create a Neon database and copy the connection string:

```bash
# Using the Aerostack CLI:
aerostack db neon create my-cron-db --add-to-config

# Or create it manually at https://console.neon.tech and paste DATABASE_URL below
```

### 2. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` — Your project ID.
- `AEROSTACK_API_KEY` — API key for this project.
- `DATABASE_URL` — Your Neon connection string (`postgres://user:pass@host/dbname`).

> **Important**: `DATABASE_URL` must also be set as a production secret before deploying:
> ```bash
> aerostack secrets set DATABASE_URL "your-neon-url" --env production
> ```

### 3. Run Locally

```bash
aerostack dev
```

The cron schedule in `aerostack.toml` defaults to `* * * * *` (every minute). You can adjust it:

```toml
[[triggers]]
crons = ["0 * * * *"]   # hourly
```

Trigger the scheduled handler manually during dev:

```bash
curl "http://localhost:8787/__scheduled?cron=*+*+*+*+*"
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

> **Reminder**: Set `DATABASE_URL` as a production secret **before** or **immediately after** deploying:
> ```bash
> aerostack secrets set DATABASE_URL "your-neon-url" --env production
> ```
> Or add `--sync-secrets` to auto-push secrets from `.dev.vars`:
> ```bash
> aerostack deploy --sync-secrets --env production
> ```

### Extending

- **Multiple schedules**: Add more cron expressions to `crons = [...]` in `aerostack.toml`.
- **AI analysis**: Uncomment the `sdk.ai.generate` call to run AI-powered analysis on your data during each tick.
- **Alerting**: Use `sdk.queue.enqueue` to fire off alerts or notifications from the scheduled handler.

### Using a Non-Neon Postgres Database

`sdk.db` works with **any Postgres-compatible database**, not just Neon. Set `DATABASE_URL` to any Postgres connection string:

```
DATABASE_URL=postgresql://user:pass@your-host:5432/dbname
```

**Supported**: Neon, Supabase, Railway, Render, Fly.io Postgres, AWS RDS, GCP Cloud SQL, CockroachDB
**Not supported**: MySQL, MongoDB, Turso, or any non-Postgres wire protocol

> **Production note**: Non-Neon databases require [Cloudflare Hyperdrive](https://developers.cloudflare.com/hyperdrive/) since Workers can't open raw TCP connections. Neon uses HTTP natively and needs no extra setup.

### More Documentation

- Online docs: `https://docs.aerostack.dev`
