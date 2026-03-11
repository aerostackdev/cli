## Aerostack WebSocket Multiplayer Game Template

This project was generated from the **ws-multiplayer-game** template. It demonstrates real-time state sync across multiple players in a shared room using Cloudflare Workers WebSockets.

Room state (player positions) is snapshotted to KV after every join/move/leave via `ctx.waitUntil()`, so clients can pre-load the current state via HTTP before connecting.

### What's Included

- `GET /` — health check.
- `GET /rooms/<roomId>` — HTTP snapshot of current players in the room (reads from KV). Call this before connecting to pre-load state.
- `WS /ws/<roomId>` — WebSocket endpoint. All players in the same `roomId` receive state updates.

### Message Protocol

**Send (client → server):**

| Message | Description |
|---------|-------------|
| `{ type: "join", name: "Alice" }` | Set your player name |
| `{ type: "move", x: 10, y: 20 }` | Update your position |
| `{ type: "chat", text: "hi" }` | Broadcast a chat message |
| `{ type: "ping" }` | Keepalive — server responds with `pong` |

**Receive (server → client):**

| Message | Description |
|---------|-------------|
| `{ type: "joined", playerId, name, roomId }` | Confirms your join |
| `{ type: "state", players: [...] }` | Full room state (sent after every change) |
| `{ type: "chat", from, text }` | A player's chat message |
| `{ type: "system", message }` | Join/leave notifications |
| `{ type: "pong" }` | Response to ping |
| `{ type: "error", message }` | Error details |

**PlayerState shape:** `{ id, name?, roomId, x, y }`

### 1. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` — Your project ID from the Aerostack dashboard.
- `AEROSTACK_API_KEY` — API key for this project (keep this secret).

> **Note**: This template uses the `CACHE` binding (KV) to snapshot room state. It is pre-configured in `aerostack.toml`.

### 2. Run Locally

```bash
aerostack dev
```

Open two browser tabs (or two terminal sessions):

```js
// Tab 1
const roomId = 'demo-room';
const ws1 = new WebSocket(`ws://localhost:8787/ws/${roomId}`);
ws1.onmessage = (e) => console.log('Tab1:', e.data);
ws1.onopen = () => ws1.send(JSON.stringify({ type: 'join', name: 'Alice' }));

// Tab 2
const ws2 = new WebSocket(`ws://localhost:8787/ws/${roomId}`);
ws2.onmessage = (e) => console.log('Tab2:', e.data);
ws2.onopen = () => ws2.send(JSON.stringify({ type: 'join', name: 'Bob' }));
```

Move and chat:

```js
ws1.send(JSON.stringify({ type: 'move', x: 5, y: 10 }));
ws2.send(JSON.stringify({ type: 'chat', text: 'hello from Bob' }));
```

Both tabs receive `state` messages after every change. Check the HTTP snapshot too:

```bash
curl http://localhost:8787/rooms/demo-room
```

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

- **Durable Objects**: For true multi-instance consistency, migrate room state to a Durable Object so all worker instances share the same in-memory map.
- **Auth**: Validate a token on the WebSocket upgrade request and replace the random UUID with a real user ID.
- **Game loop**: Add server-side ticks via `ctx.waitUntil` + recursive `setTimeout`.

### More Documentation

- Online docs: `https://docs.aerostack.dev`
