## Aerostack WebSocket Chat Template

This project was generated from the **ws-chat** template. It provides a simple **group chat** over WebSockets: everyone in the same room sees each other’s messages.

You can extend it later with one-to-one chat (send to a specific user) or persistence (e.g. Neon/D1).

### What’s Included

- Worker entry at `src/index.ts`.
- A `/` HTTP route.
- A `/ws/<roomId>` endpoint mapping clients into isolated chat rooms.
- Built-in message history mapping (the last 20 messages in a room are stored in KV and relayed to new joiners).

### 1. Configure Local Environment

1. Copy the example vars:
```bash
cp .dev.vars.example .dev.vars
```
2. Edit `.dev.vars` and add your `AEROSTACK_PROJECT_ID` and `AEROSTACK_API_KEY`.

> **Note**: This template uses the `CACHE` binding (Aerostack KV) to store chat history.

### 2. Run Locally

```bash

Open two browser tabs in the same room:

```js
const roomId = 'lobby';
const ws = new WebSocket(`ws://localhost:8787/ws/${roomId}`);

ws.onmessage = (event) => console.log('Received:', event.data);

// Send a message
ws.send(JSON.stringify({ type: 'chat', text: 'Hello everyone!' }));
```

Both tabs should receive each other’s messages and system join/leave events.

### 3. Deploy

When you’re ready:

```bash
aerostack deploy
```

### Extending (later)

- **One-to-one**: Add a `to` field to chat messages and route only to that user’s WebSocket.
- **Persistence**: Use `sdk.db` or Neon to store messages and replay on join.
- **Auth**: Validate tokens on connection and attach a real user id instead of a random UUID.

### More Documentation

- Template overview: `sdks/packages/cli/docs/TEMPLATES_OVERVIEW.md` (in the Aerostack repo).
- Online docs: `https://docs.aerostack.dev`.
