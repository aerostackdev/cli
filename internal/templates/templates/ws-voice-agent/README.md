## Aerostack WebSocket Voice Agent Template

This project was generated from the **ws-voice-agent** template. It provides a Cloudflare Worker WebSocket server that acts as a conversational AI agent using `sdk.ai.generate`. It automatically tracks conversation history during the session.

### What’s Included

- A `/` HTTP route for health checks.
- A `/ws` route that accepts standard WebSocket Upgrades.
- In-memory conversation history that sends context to the AI on every prompt.

### 1. Configure Local Environment

1. Copy the example vars:
```bash
cp .dev.vars.example .dev.vars
```
2. Edit `.dev.vars` and add your `AEROSTACK_PROJECT_ID` and `AEROSTACK_API_KEY`.

### 2. Run Locally

```bash
aerostack dev
```

You should see the health check at:

- `GET /` – returns a short text message.

To test the WebSocket endpoint, open your browser console or a small Node script and run:

```js
const ws = new WebSocket('ws://localhost:8787/ws');

ws.onopen = () => {
  ws.send(JSON.stringify({ type: 'user_text', text: 'Hello from the browser' }));
};

ws.onmessage = (event) => {
  console.log('Received:', event.data);
};
```

You should receive a JSON message with `type: "assistant_text"`.

### 3. Deploy

When you’re ready:

```bash
aerostack deploy
```

### More Documentation

- Template overview: `sdks/packages/cli/docs/TEMPLATES_OVERVIEW.md` (in the Aerostack repo).
- Online docs: `https://docs.aerostack.dev`.

