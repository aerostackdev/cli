## Aerostack Blank Template

This project was generated from the **blank** template. It is a minimal Cloudflare Worker that demonstrates how to use the Aerostack SDK for database, cache, AI, and background jobs.

### What’s Included

- Single entry file at `src/index.ts`.
- `fetch` handler with sample routes:
  - `GET /` – Health check returning JSON status.
  - `GET /test/db` – Creates a `notes` table, inserts a row, and returns all rows via D1.
  - `GET /test/cache` – Stores and retrieves a value from KV.
  - `GET /test/ai` – Calls the Aerostack AI proxy for a 1-sentence joke.
  - `POST /test/queue` – Enqueues a background job.
- `queue` handler that receives and acknowledges jobs.

### 1. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` – Your project ID from the Aerostack dashboard.
- `AEROSTACK_API_KEY` – API key or JWT for this project (keep this secret).

> **Important**: Do **not** commit `.dev.vars` to git. Only `.dev.vars.example` should be tracked.

### 2. Run Locally

Start the local development server:

```bash
npm run dev
# or: aerostack dev
```

Test the endpoints (in another terminal or in your browser):

```bash
# Health check
curl http://localhost:8787/

# Database test
curl http://localhost:8787/test/db

# Cache test
curl http://localhost:8787/test/cache

# AI test (requires valid credentials)
curl http://localhost:8787/test/ai

# Queue test
curl -X POST http://localhost:8787/test/queue
```

### 3. Deploy

Once your project is configured in the Aerostack dashboard, deploy with:

```bash
aerostack deploy
```

### More Documentation

- View the full documentation at: `https://docs.aerostack.dev`

