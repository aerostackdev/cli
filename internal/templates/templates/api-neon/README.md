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

```bash
aerostack deploy
```

Make sure in the Aerostack dashboard that:

- Neon is correctly wired via `DATABASE_URL` or the `PG` binding.
- KV and Queue bindings are configured (`CACHE`, `QUEUE`).

### More Documentation

- Template overview: `sdks/packages/cli/docs/TEMPLATES_OVERVIEW.md`.
- Neon specifics: `sdks/packages/cli/docs/NEON_GUIDE.md`.
- Online docs: `https://docs.aerostack.dev`.

