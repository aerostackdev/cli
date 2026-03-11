## Aerostack API Template (Hono + D1)

This project was generated from the **api** template. It is a Hono-based REST API running on Cloudflare Workers, using D1 as the primary database, with KV, AI, and queue support included.

### What’s Included

- Hono app defined in `src/index.ts`.
- Routes:
  - `GET /` – Basic welcome message.
  - `GET /health` – JSON health status.
  - `GET /notes`, `POST /notes` – Fully functional D1 CRUD example.
  - `GET /test/cache`, `GET /test/ai`, `POST /test/queue` – KV, AI, and Queue examples.
  - `GET /users/:id` – Simple parameterized endpoint.
- `queue` handler for background jobs.
- `aerostack.toml` with bindings for:
  - `DB` (D1)
  - `CACHE` (KV)
  - `QUEUE` (queue producer)

### 1. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` – Your project ID from the Aerostack dashboard.
- `AEROSTACK_API_KEY` – API key or JWT for this project (keep this secret).

> **Note**: `.dev.vars` is for local development only and should never be committed.

### 2. Run Locally

Start the local development server:

```bash
npm run dev
# or: aerostack dev
```

Test the endpoints (in another terminal or in your browser):

```bash
# Health check
curl http://localhost:8787/health

# Create a note in D1
curl -X POST http://localhost:8787/notes \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello from Aerostack!"}'

# Read notes from D1
curl http://localhost:8787/notes

# Cache test
curl http://localhost:8787/test/cache

# AI test (requires valid credentials)
curl http://localhost:8787/test/ai

# Queue test
curl -X POST http://localhost:8787/test/queue
```

### 3. Deploy

When you’re ready to deploy:

**First time only — link to your Aerostack project:**

```bash
aerostack link --write-toml
```

This picks your project interactively and writes `project_id` into `aerostack.toml` so every future deploy goes to the right project automatically.

```bash
aerostack deploy
```

Make sure in the Aerostack dashboard that:

- A D1 database is attached to the `DB` binding.
- A KV namespace is attached to `CACHE`.
- A queue is configured for `QUEUE`.

### More Documentation

- Neon usage guide (if you upgrade to Neon later): `sdks/packages/cli/docs/NEON_GUIDE.md`.
- View the full documentation at: `https://docs.aerostack.dev`.

