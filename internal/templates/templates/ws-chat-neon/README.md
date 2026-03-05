## Aerostack WebSocket Chat (Neon) Template

This project was generated from the **ws-chat-neon** template. It is a production-ready group chat over WebSockets backed by **Neon PostgreSQL** for durable message persistence — messages survive worker restarts and are queryable via SQL.

Upgrade from `ws-chat` (KV-only) when you need full message history, search, or moderation features.

### What's Included

- `WS /ws/<roomId>` — WebSocket chat endpoint. All clients in the same room share messages.
- `GET /rooms/<roomId>/history` — HTTP endpoint to fetch the last 50 messages from Neon.
- Auto-creates the `messages` table on first use (`ensureSchema()`).
- Message writes go to Neon via `ctx.waitUntil()` — never blocks the broadcast.
- Join replay: new joiners receive the last 50 messages from the database.

### Message Protocol

**Send (client → server):**

| Message | Description |
|---------|-------------|
| `{ type: "join", username: "Alice" }` | Set display name |
| `{ type: "chat", text: "Hello!" }` | Broadcast to room |
| `{ type: "ping" }` | Keepalive |

**Receive (server → client):**

| Message | Description |
|---------|-------------|
| `{ type: "joined", userId, username, roomId }` | Confirm join |
| `{ type: "chat", from, username, text, timestamp }` | A message |
| `{ type: "system", message }` | Join/leave events |
| `{ type: "pong" }` | Keepalive response |
| `{ type: "error", message }` | Error |

### 1. Set Up Neon

```bash
aerostack db neon create my-chat-db --add-to-config
# or create manually at https://console.neon.tech
```

### 2. Configure Local Environment

```bash
cp .dev.vars.example .dev.vars
```

Fill in `.dev.vars`:
- `AEROSTACK_PROJECT_ID` — project ID from the dashboard
- `AEROSTACK_API_KEY` — project API key
- `DATABASE_URL` — Neon connection string

### 3. Run Locally

```bash
aerostack dev
```

```js
const ws = new WebSocket('ws://localhost:8787/ws/lobby');
ws.onopen = () => ws.send(JSON.stringify({ type: 'join', username: 'Alice' }));
ws.onmessage = (e) => console.log(e.data);
ws.send(JSON.stringify({ type: 'chat', text: 'Hello!' }));
```

Fetch history via HTTP:

```bash
curl http://localhost:8787/rooms/lobby/history
```

### 4. Deploy

```bash
# Set DATABASE_URL secret first, then deploy
aerostack secrets set DATABASE_URL "your-neon-url" --env production
aerostack deploy

# Or use --sync-secrets to push .dev.vars secrets automatically
aerostack deploy --sync-secrets --env production
```

### Extending

- **Message search**: Add a `GET /rooms/:roomId/search?q=...` endpoint using `sdk.db.query` with `ILIKE`.
- **Moderation**: Add a `DELETE /messages/:id` endpoint with role checks.
- **Auth**: Validate a token on the WebSocket upgrade and record the real user ID in Neon.
- **Pagination**: Add cursor-based pagination to the `/history` endpoint using `WHERE id < $cursor`.

### Using a Non-Neon Postgres Database

`sdk.db` works with **any Postgres-compatible database**. Swap Neon for Supabase, Railway, Render, etc. by setting `DATABASE_URL` in `.dev.vars`:

```
DATABASE_URL=postgresql://user:pass@your-host:5432/dbname
```

**Supported**: Neon, Supabase, Railway, Render, Fly.io Postgres, AWS RDS, GCP Cloud SQL, CockroachDB
**Not supported**: MySQL, MongoDB, Turso, or any non-Postgres wire protocol

> **Production note**: Non-Neon databases require [Cloudflare Hyperdrive](https://developers.cloudflare.com/hyperdrive/) since Workers can't open raw TCP connections. Neon uses HTTP natively — no Hyperdrive needed.

### More Documentation

- Online docs: `https://docs.aerostack.dev`
