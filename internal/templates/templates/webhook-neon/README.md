## Aerostack Webhook + Neon Template

This project was generated from the **webhook-neon** template. It provides a Cloudflare Worker that receives webhook `POST` requests and stores each payload to a Neon PostgreSQL database.

### What's Included

- `POST /webhook` — Accepts any JSON payload and inserts it into a `webhooks` table in Neon.
- Hono routing via `@aerostack/sdk` + Hono.
- Pre-configured Neon (`PG`), KV (`CACHE`), and Queue bindings in `aerostack.toml`.

### 1. Set Up Neon

```bash
# Using the Aerostack CLI:
aerostack db neon create my-webhook-db --add-to-config

# Or create manually at https://console.neon.tech and set DATABASE_URL below
```

Create the table in your Neon database:

```sql
CREATE TABLE IF NOT EXISTS webhooks (
  id SERIAL PRIMARY KEY,
  payload JSONB,
  received_at TIMESTAMPTZ DEFAULT NOW()
);
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

> **Important**: `DATABASE_URL` must also be set as a production secret:
> ```bash
> aerostack secrets set DATABASE_URL "your-neon-url" --env production
> ```

### 3. Run Locally

```bash
aerostack dev
```

Send a test webhook:

```bash
curl -X POST http://localhost:8787/webhook \
  -H 'Content-Type: application/json' \
  -d '{"event": "user.created", "userId": "abc123"}'
```

You should receive `{ "status": "received" }` and see the row in your Neon database.

### 4. Deploy

```bash
aerostack deploy --public
```

> `--public` makes the endpoint accessible without an API key, which is typically what webhook sources need.

> **Reminder**: Set `DATABASE_URL` as a production secret before your first deploy:
> ```bash
> aerostack secrets set DATABASE_URL "your-neon-url" --env production
> ```
> Or use `--sync-secrets` to push all `.dev.vars` secrets automatically:
> ```bash
> aerostack deploy --public --sync-secrets --env production
> ```

### Extending

- **Signature verification**: Validate HMAC signatures (e.g. Stripe, GitHub) before inserting to prevent spoofed payloads.
- **Queue fanout**: After saving, enqueue the payload for async processing via `sdk.queue.enqueue`.
- **Retry handling**: Parse `X-Retry-Count` headers and deduplicate re-deliveries using a unique event ID.

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
