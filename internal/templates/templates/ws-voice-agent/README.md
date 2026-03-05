## Aerostack WebSocket Voice Agent Template

This project was generated from the **ws-voice-agent** template. It provides a Cloudflare Worker WebSocket server that acts as a conversational AI agent using `sdk.ai.generate`.

Conversation history is persisted to KV (1 hour TTL), so reconnecting with the same `sessionId` restores the full context.

### What's Included

- `GET /` — health check and usage info.
- `WS /ws?sessionId=<id>` — WebSocket endpoint. Omit `sessionId` for a fresh session; reuse it to continue a conversation.
- History persisted via `ctx.waitUntil()` — KV writes never block the AI response.

### Message Protocol

**Send (client → server):**

| Message | Description |
|---------|-------------|
| `{ type: "user_text", text: "Hello" }` | Send a message to the AI |
| `{ type: "clear" }` | Clear history for this session |
| `{ type: "ping" }` | Keepalive — server responds with `pong` |

**Receive (server → client):**

| Message | Description |
|---------|-------------|
| `{ type: "status", historyLength, sessionId }` | Sent on connect — includes restored history count |
| `{ type: "assistant_text", text }` | AI response |
| `{ type: "pong" }` | Response to ping |
| `{ type: "error", message }` | Error details |

### 1. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and add your `AEROSTACK_PROJECT_ID` and `AEROSTACK_API_KEY`.

> **Note**: This template uses the `CACHE` binding (KV) for session persistence. It is pre-configured in `aerostack.toml`.

### 2. Run Locally

```bash
aerostack dev
```

Connect from a browser console or a Node.js script:

```js
// Fresh session — server generates a new sessionId
const ws = new WebSocket('ws://localhost:8787/ws');

ws.onopen = () => {
  ws.send(JSON.stringify({ type: 'user_text', text: 'What is the capital of France?' }));
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'status') {
    console.log('Connected. Session:', msg.sessionId, '— history:', msg.historyLength, 'turns');
  }
  if (msg.type === 'assistant_text') {
    console.log('AI:', msg.text);
  }
};
```

**Resume a previous session** by passing `sessionId`:

```js
const sessionId = 'your-session-id-from-a-previous-status-message';
const ws = new WebSocket(`ws://localhost:8787/ws?sessionId=${sessionId}`);
```

You will receive a `status` message with the number of history turns restored.

### 3. Deploy

```bash
aerostack deploy
```

### Extending

- **Streaming responses**: Replace `sdk.ai.generate` with a streaming variant and use SSE or chunked WebSocket messages for token-by-token output.
- **System prompt**: Customise the AI persona by prepending a system prompt to the `contextLines` string.
- **Auth**: Validate a token on the WebSocket upgrade and scope sessions per user so their history is private.
- **Longer retention**: Increase `SESSION_TTL_SECONDS` or store history in D1/Neon for permanent conversation logs.

### More Documentation

- Online docs: `https://docs.aerostack.dev`
