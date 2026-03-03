## Aerostack Project Templates

This document describes the core templates exposed by `aerostack init` and how to run them in both local development and production.

- **blank** – Minimal Worker with direct `fetch` / `queue` handlers and sample DB, cache, AI, and queue usage.
- **api** – Hono-based REST API using D1 (SQLite at the edge) plus cache, AI, and queue helpers.
- **api-neon** – Hono-based REST API using Neon PostgreSQL via `DATABASE_URL`, plus cache, AI, and queue helpers.
- **ws-voice-agent** – WebSocket voice/chat agent that turns text into AI responses.
- **ws-multiplayer-game** – WebSocket multiplayer room sample that syncs simple player state.
- **ws-chat** – WebSocket group chat (extend later with one-to-one or persistence).

### Common Environment Variables

All templates expect the Aerostack SDK to be configured via `sdk.init(env)`. For local development, copy the template’s `.dev.vars.example` file to `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` – Project ID from the Aerostack dashboard.
- `AEROSTACK_API_KEY` – API key or JWT for this project (keep this secret and never commit it).
- `AEROSTACK_API_URL` – Optional override for the Aerostack API base URL (defaults to the hosted API).

These variables are not committed to git. Only `.dev.vars.example` lives in the template.

---

## `blank` – Minimal Worker

**Purpose**: Start from the smallest possible Worker while still having end‑to‑end examples for DB, cache, AI, and queues.

- **Entry**: `src/index.ts`
- **Routing**: Direct `fetch(request, env, ctx)` handler (no router).
- **Features demonstrated**:
  - `GET /test/db` – Creates a `notes` table, inserts a row, and returns all rows.
  - `GET /test/cache` – Writes and reads a value from KV.
  - `GET /test/ai` – Calls `sdk.ai.generate(...)` to return a short text.
  - `POST /test/queue` – Enqueues a background job, handled by the `queue` consumer.

### Local Development

```bash
aerostack init my-blank-app --template=blank --db=d1
cd my-blank-app
cp .dev.vars.example .dev.vars  # fill in values
aerostack dev
```

Hit the test endpoints in your browser or with `curl`:

- `GET /test/db`
- `GET /test/cache`
- `GET /test/ai`
- `POST /test/queue`

### Production

After configuring your project in the Aerostack dashboard (DB, KV, Queue, AI proxy), deploy:

```bash
aerostack deploy
```

The same routes should behave the same way in production.

---

## `api` – Hono + D1

**Purpose**: Build conventional REST APIs using Hono with D1 as the primary database.

- **Entry**: `src/index.ts`
- **Routing**:
  - `GET /` – Welcome message.
  - `/test` router with:
    - `GET /test/db` – Example D1 query.
    - `GET /test/cache` – KV example.
    - `GET /test/ai` – AI proxy example.
  - `GET /users/:id` – Simple path parameter example.
- **Queue**:
  - `queue(batch, env)` consumer that logs and acknowledges background jobs.

### Local Development

```bash
aerostack init my-api-app --template=api --db=d1
cd my-api-app
cp .dev.vars.example .dev.vars  # fill in values
aerostack dev
```

Smoke test:

- `GET /`
- `GET /test/db`
- `GET /test/cache`
- `GET /test/ai`
- `GET /users/123`

### Production

Ensure your Aerostack project has:

- A D1 database attached to the `DB` binding.
- KV namespace for `CACHE`.
- Queue for `QUEUE`.
- AI proxy configured.

Then deploy with:

```bash
aerostack deploy
```

---

## `api-neon` – Hono + Neon PostgreSQL

**Purpose**: Same as the `api` template, but backed by Neon PostgreSQL instead of D1.

- **Entry**: `src/index.ts`
- **Routing**:
  - `GET /` – Welcome message mentioning Neon.
  - `GET /users` – Reads from a `users` table via `sdk.db.query(...)`.
  - `/test` router with:
    - `GET /test/cache` – KV example.
    - `GET /test/ai` – AI proxy example.
- **Queue**:
  - `queue(batch, env)` logs and acknowledges jobs.

### Required Environment

In addition to the common Aerostack variables, this template expects:

- `DATABASE_URL` – Neon PostgreSQL connection string used by the `PG` binding.

### Local Development

```bash
aerostack init my-neon-api --template=api-neon --db=neon
cd my-neon-api
cp .dev.vars.example .dev.vars  # fill in values, including DATABASE_URL
aerostack dev
```

Make sure your Neon database has a `users` table before hitting:

- `GET /`
- `GET /users`
- `GET /test/cache`
- `GET /test/ai`

### Production

Configure your Aerostack project so that:

- Neon is wired via `DATABASE_URL` or the `PG` binding.
- KV and Queue bindings are set up as in the `api` template.

Then deploy with:

```bash
aerostack deploy
```

---

## `ws-voice-agent` – WebSocket Voice/Chat Agent

**Purpose**: Provide a minimal WebSocket endpoint that converts user text messages into AI-generated responses using the Aerostack SDK.

- **Entry**: `src/index.ts`
- **Routing**:
  - `GET /` – Health check and short instructions.
  - `GET /ws` (with `Upgrade: websocket`) – WebSocket endpoint.
- **Protocol**:
  - Incoming:
    - `{ "type": "ping" }`
    - `{ "type": "user_text", "text": "Hello" }`
  - Outgoing:
    - `{ "type": "pong" }`
    - `{ "type": "assistant_text", "text": "..." }`
    - `{ "type": "error", "message": "..." }`

### Local Development

```bash
aerostack init my-voice-agent --template=ws-voice-agent --db=d1
cd my-voice-agent
cp .dev.vars.example .dev.vars  # fill in values
aerostack dev
```

Connect from a browser or Node client to `ws://localhost:8787/ws` and send JSON messages as described above.

---

## `ws-multiplayer-game` – WebSocket Multiplayer Sample

**Purpose**: Demonstrate a simple room-based multiplayer pattern over WebSockets, where multiple clients share player state.

- **Entry**: `src/index.ts`
- **Routing**:
  - `GET /` – Health check and instructions.
  - `GET /ws/:roomId` – WebSocket endpoint for the given room.
- **Protocol**:
  - Incoming:
    - `{ "type": "ping" }`
    - `{ "type": "move", "x": 5, "y": 10 }`
    - `{ "type": "chat", "text": "hello" }`
  - Outgoing:
    - `{ "type": "state", "players": [...] }`
    - `{ "type": "chat", "from": "<playerId>", "text": "..." }`
    - `{ "type": "system", "message": "..." }`

### Local Development

```bash
aerostack init my-game --template=ws-multiplayer-game --db=d1
cd my-game
cp .dev.vars.example .dev.vars  # fill in values
aerostack dev
```

Open two browser tabs pointing at the same room ID and verify that movement and chat messages are synchronized between them.

---

## `ws-chat` – WebSocket Group Chat

**Purpose**: Simple group chat over WebSockets. Everyone in the same room sees each other’s messages. You can extend it later with one-to-one chat or persistence (e.g. Neon/D1).

- **Entry**: `src/index.ts`
- **Routing**:
  - `GET /` – Health check and instructions.
  - `GET /ws/:roomId` – WebSocket endpoint for the given room.
- **Protocol**:
  - Incoming:
    - `{ "type": "ping" }`
    - `{ "type": "chat", "text": "hello" }`
  - Outgoing:
    - `{ "type": "pong" }`
    - `{ "type": "chat", "from": "<userId>", "text": "..." }`
    - `{ "type": "system", "message": "..." }` (join/leave)
    - `{ "type": "error", "message": "..." }`

### Local Development

```bash
aerostack init my-chat --template=ws-chat --db=d1
cd my-chat
cp .dev.vars.example .dev.vars  # fill in values
aerostack dev
```

Connect two clients to the same room (e.g. `ws://localhost:8787/ws/lobby`) and send `{ "type": "chat", "text": "Hello" }`; both should receive the message.

---

## Further Reading

For more details, see:

- Neon usage: `sdks/packages/cli/docs/NEON_GUIDE.md`
- Multi-function examples: `sdks/packages/cli/docs/MULTI_FUNCTION_GUIDE.md`
- External docs site: `https://docs.aerostack.dev` (update this link if your docs URL differs).


