## Aerostack WebSocket Chat Template

This project was generated from the **ws-chat** template. It provides a simple **group chat** over WebSockets: everyone in the same room sees each other's messages.

You can extend it later with one-to-one chat (send to a specific user) or full persistence via Neon/D1.

### What's Included

- Worker entry at `src/index.ts`.
- A `/` HTTP route (health check).
- A `/ws/<roomId>` endpoint — clients connecting to the same roomId share a room.
- Last 20 messages per room stored in KV (`CACHE`) and replayed to new joiners.
- Non-blocking KV writes via `ctx.waitUntil()` — history saves never stall the message loop.

### Message Protocol

**Send (client → server):**

| Message | Description |
|---------|-------------|
| `{ type: "join", username: "Alice" }` | Set your display name |
| `{ type: "chat", text: "Hello!" }` | Broadcast a message to the room |
| `{ type: "ping" }` | Keepalive — server responds with `pong` |

**Receive (server → client):**

| Message | Description |
|---------|-------------|
| `{ type: "joined", userId, username, roomId }` | Confirms your join |
| `{ type: "chat", from, username, text, timestamp }` | A chat message |
| `{ type: "system", message }` | Join/leave notifications |
| `{ type: "pong" }` | Response to ping |
| `{ type: "error", message }` | Error details |

### 1. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and add your `AEROSTACK_PROJECT_ID` and `AEROSTACK_API_KEY`.

> **Note**: This template uses the `CACHE` binding (KV) to store chat history. It is pre-configured in `aerostack.toml`.

### 2. Run Locally

```bash
aerostack dev
```

Open two browser tabs (or two terminal `wscat` sessions) pointed at the same room:

```js
const roomId = 'lobby';
const ws = new WebSocket(`ws://localhost:8787/ws/${roomId}`);

ws.onopen = () => {
  ws.send(JSON.stringify({ type: 'join', username: 'Alice' }));
};

ws.onmessage = (event) => console.log('Received:', event.data);

// Send a chat message
ws.send(JSON.stringify({ type: 'chat', text: 'Hello everyone!' }));
```

Both connections should receive each other's messages and system join/leave events. New joiners receive the last 20 messages as replay.

### 3. Deploy

**First time only — link to your Aerostack project:**

```bash
aerostack link --write-toml
```

This picks your project interactively and writes `project_id` into `aerostack.toml` so every future deploy goes to the right project automatically.

```bash
aerostack deploy
```

### Extending

- **One-to-one DMs**: Add a `to` field to chat messages and route only to the target user's WebSocket.
- **Database persistence**: Replace the KV history with `sdk.db` (D1) or Neon for durable, queryable message storage.
- **Auth**: Validate a token on the WebSocket upgrade request and replace the random UUID with a real user ID.

### More Documentation

- Online docs: `https://docs.aerostack.dev`
