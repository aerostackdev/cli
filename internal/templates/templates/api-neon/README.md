## Aerostack API Template (Neon PostgreSQL)

This project was generated from the **api-neon** template. It is a Hono-based REST API running on Cloudflare Workers, using [Neon](https://neon.tech) (PostgreSQL) as the primary database, alongside Aerostack KV, AI, and Queue features.

### What’s Included

- Hono app defined in `src/index.ts`.
- Routes:
  - `GET /` – Basic welcome message.
  - `GET /health` – JSON health status.
  - `GET /users`, `POST /users` – Fully functional Postgres CRUD example.
  - `GET /test/cache`, `GET /test/ai`, `POST /test/queue` – KV, AI, and Queue examples.
- `queue` handler for background jobs.
- `aerostack.toml` with bindings for:
  - `PG` (PostgreSQL via Neon pooler)
  - `CACHE` (KV)
  - `QUEUE` (queue producer)

### 1. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` – Your project ID from the Aerostack dashboard.
- `GET /test/cache`
- `GET /test/ai`

You should see responses backed by Neon and KV.

### 4. Deploy

When ready:

**First time only — link to your Aerostack project:**

```bash
aerostack link --write-toml
```

This picks your project interactively and writes `project_id` into `aerostack.toml` so every future deploy goes to the right project automatically.

```bash
aerostack deploy
```

Make sure in the Aerostack dashboard that:

- Neon is correctly wired via `DATABASE_URL` or the `PG` binding.
- KV and Queue bindings are configured (`CACHE`, `QUEUE`).

### Using a Non-Neon Postgres Database

`sdk.db` works with **any Postgres-compatible database**, not just Neon. Set your connection string in `.dev.vars`:

```
DATABASE_URL=postgresql://user:pass@your-host:5432/dbname
```

**Supported**: Neon, Supabase, Railway, Render, Fly.io Postgres, AWS RDS, GCP Cloud SQL, CockroachDB
**Not supported**: MySQL, MongoDB, Turso, or any non-Postgres wire protocol

> **Production note**: Cloudflare Workers can't open raw TCP connections. For non-Neon databases you need [Cloudflare Hyperdrive](https://developers.cloudflare.com/hyperdrive/) to proxy TCP → HTTP:
> ```bash
> wrangler hyperdrive create my-db --connection-string="postgresql://..."
> # Then update the `id` field under [[hyperdrive]] in .aerostack/wrangler.toml
> ```
> For Neon, Hyperdrive is **not required** — Neon's serverless driver uses HTTP natively.

### More Documentation

- Template overview: `sdks/packages/cli/docs/TEMPLATES_OVERVIEW.md`.
- Neon specifics: `sdks/packages/cli/docs/NEON_GUIDE.md`.
- Online docs: `https://docs.aerostack.dev`.

