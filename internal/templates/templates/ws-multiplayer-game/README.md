## Aerostack WebSocket Multiplayer Game Template

This project was generated from the **ws-multiplayer-game** template. It demonstrates a simple state-synchronized multiplayer room using Cloudflare Workers WebSockets. 

### What’s Included

- A `/` HTTP route for simple info.
- A `/ws/<roomId>` route that handles room-based state sync, player joining, movement, and chat.

### 1. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` – Your project ID from the Aerostack dashboard.
- `AEROSTACK_API_KEY` – API key or JWT for this project (keep this secret).
- `AEROSTACK_API_URL` – Optional override of the Aerostack API base URL.

> **Note**: `.dev.vars` is for local development only and should never be committed.

### 2. Run Locally

```bash
aerostack dev
```

Open two browser tabs with the same room ID, for example:

```js
// Tab 1
const roomId = 'demo-room';
const ws1 = new WebSocket(`ws://localhost:8787/ws/${roomId}`);
ws1.onmessage = (event) => console.log('Tab1:', event.data);

// Tab 2
const ws2 = new WebSocket(`ws://localhost:8787/ws/${roomId}`);
ws2.onmessage = (event) => console.log('Tab2:', event.data);
```

Then send messages from either tab:

```js
ws1.send(JSON.stringify({ type: 'move', x: 5, y: 10 }));
ws2.send(JSON.stringify({ type: 'chat', text: 'hello from tab 2' }));
```

Both tabs should receive updated `state` and `chat` messages.

### 3. Deploy

When you’re ready:

```bash
aerostack deploy
```

### More Documentation

- Template overview: `sdks/packages/cli/docs/TEMPLATES_OVERVIEW.md` (in the Aerostack repo).
- Online docs: `https://docs.aerostack.dev`.

