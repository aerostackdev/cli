## Aerostack AI Stream Template

This project was generated from the **ai-stream** template. It demonstrates **Server-Sent Events (SSE)** streaming AI responses from a Cloudflare Worker — the standard pattern for chat UIs, copilots, and any feature where users should see output progressively.

### What's Included

- `GET /generate?prompt=...` — non-streaming AI response (for comparison / testing).
- `POST /stream` — SSE endpoint that generates a response and streams it word by word.
- Clean `TransformStream` pattern that works reliably in Cloudflare Workers.

### SSE Event Format

Every event follows the standard SSE wire format:

```
event: token
data: {"token": " the", "text": "Paris is the"}

event: done
data: {"text": "Paris is the capital of France."}
```

| Event | When | Data |
|-------|------|------|
| `start` | Before AI call | `{ message }` |
| `token` | Each word chunk | `{ token, text }` — `text` is the accumulated response so far |
| `done` | After last token | `{ text }` — full response |
| `error` | On failure | `{ message }` |

### 1. Configure Local Environment

```bash
cp .dev.vars.example .dev.vars
```

Fill in `.dev.vars`:
- `AEROSTACK_PROJECT_ID` — project ID from the dashboard
- `AEROSTACK_API_KEY` — project API key

### 2. Run Locally

```bash
aerostack dev
```

**Stream a response:**

```bash
curl -N -X POST http://localhost:8787/stream \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "Explain serverless computing in 2 sentences"}'
```

The `-N` flag disables buffering so you see events as they arrive.

**In JavaScript (EventSource):**

```js
// Note: EventSource only supports GET. For POST, use fetch with a ReadableStream.
const response = await fetch('http://localhost:8787/stream', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ prompt: 'What is Cloudflare Workers?' }),
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;

  const chunk = decoder.decode(value);
  // Parse SSE lines
  for (const line of chunk.split('\n')) {
    if (line.startsWith('data: ')) {
      const data = JSON.parse(line.slice(6));
      if (data.token) process.stdout.write(data.token); // or update the UI
    }
  }
}
```

**Non-streaming for comparison:**

```bash
curl "http://localhost:8787/generate?prompt=What+is+a+Worker"
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

- **Native token streaming**: Replace `sdk.ai.generate()` with a Workers AI binding call using `stream: true`. Add `ai = true` to `aerostack.toml` to enable the `AI` binding, then call `env.AI.run('@cf/meta/llama-3-8b-instruct', { prompt, stream: true })` which returns a native `EventSourceStream`.
- **Chat history**: Accept a `messages` array and pass it to `sdk.ai.chat()` for multi-turn conversations.
- **Rate limiting**: Add a request counter in KV (`sdk.cache`) per IP/API key to prevent abuse.

### More Documentation

- Online docs: `https://docs.aerostack.dev`
