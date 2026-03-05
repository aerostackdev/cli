import { sdk } from '@aerostack/sdk';

export interface Env {
  CACHE: any;
  AEROSTACK_PROJECT_ID: string;
  AEROSTACK_API_KEY: string;
}

// Max conversation turns kept in context (older turns are dropped)
const MAX_HISTORY = 10;

// Session TTL in KV — 1 hour. Reconnect with the same sessionId to restore context.
const SESSION_TTL_SECONDS = 3600;

export type IncomingMessage =
  | { type: 'ping' }
  | { type: 'user_text'; text: string }
  | { type: 'clear' };

export type OutgoingMessage =
  | { type: 'pong' }
  | { type: 'assistant_text'; text: string }
  | { type: 'status'; historyLength: number; sessionId: string }
  | { type: 'error'; message: string };

type HistoryEntry = { role: 'user' | 'assistant'; content: string };

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    sdk.init(env);
    const url = new URL(request.url);

    if (url.pathname === '/') {
      return new Response(
        'WebSocket Voice Agent is running.\n\n' +
        'Connect: /ws?sessionId=<id>  (omit sessionId for a fresh session)\n' +
        'Send:    { "type": "user_text", "text": "Hello" }\n\n' +
        'Conversation history is persisted to KV for 1 hour.\n' +
        'Reconnect with the same sessionId to continue where you left off.',
        { status: 200 },
      );
    }

    if (url.pathname === '/ws' && request.headers.get('Upgrade') === 'websocket') {
      // Use provided sessionId or generate a new one for a fresh session
      const sessionId = url.searchParams.get('sessionId') || crypto.randomUUID();

      const pair = new WebSocketPair();
      const [client, server] = Object.values(pair) as [WebSocket, WebSocket];

      handleSession(server, sessionId, ctx);

      return new Response(null, {
        status: 101,
        webSocket: client,
      });
    }

    return new Response('Not found', { status: 404 });
  },
};

function handleSession(ws: WebSocket, sessionId: string, ctx: ExecutionContext) {
  ws.accept();

  let history: HistoryEntry[] = [];

  // Restore persisted history for this sessionId on connect
  sdk.cache.get<HistoryEntry[]>(`session_${sessionId}`)
    .then(saved => {
      if (saved && saved.length > 0) {
        history = saved;
      }
      ws.send(JSON.stringify({ type: 'status', historyLength: history.length, sessionId } as OutgoingMessage));
    })
    .catch(() => {
      // No saved history — send initial status
      ws.send(JSON.stringify({ type: 'status', historyLength: 0, sessionId } as OutgoingMessage));
    });

  ws.addEventListener('message', async (event) => {
    try {
      const raw = typeof event.data === 'string' ? event.data : String(event.data);
      const msg = JSON.parse(raw) as IncomingMessage;

      if (msg.type === 'ping') {
        ws.send(JSON.stringify({ type: 'pong' } as OutgoingMessage));
        return;
      }

      if (msg.type === 'clear') {
        history = [];
        ctx.waitUntil(
          sdk.cache.set(`session_${sessionId}`, [], { expirationTtl: SESSION_TTL_SECONDS }).catch(() => {})
        );
        ws.send(JSON.stringify({ type: 'status', historyLength: 0, sessionId } as OutgoingMessage));
        return;
      }

      if (msg.type === 'user_text') {
        history.push({ role: 'user', content: msg.text });

        // Keep only the last MAX_HISTORY turns to avoid unbounded context growth
        if (history.length > MAX_HISTORY) history = history.slice(history.length - MAX_HISTORY);

        try {
          // Build a conversation prompt from history
          const contextLines = history
            .map(h => `${h.role === 'user' ? 'User' : 'Assistant'}: ${h.content}`)
            .join('\n');
          const prompt = `You are a helpful assistant. Continue the conversation below.\n\n${contextLines}\nAssistant:`;

          const { text } = await sdk.ai.generate(prompt);
          history.push({ role: 'assistant', content: text });

          // Persist updated history in background
          ctx.waitUntil(
            sdk.cache.set(`session_${sessionId}`, history, { expirationTtl: SESSION_TTL_SECONDS }).catch(() => {})
          );

          ws.send(JSON.stringify({ type: 'assistant_text', text } as OutgoingMessage));
        } catch (e: any) {
          ws.send(JSON.stringify({ type: 'error', message: e.message } as OutgoingMessage));
        }
        return;
      }

      ws.send(JSON.stringify({ type: 'error', message: 'Unknown message type' } as OutgoingMessage));
    } catch (err: any) {
      try {
        ws.send(JSON.stringify({ type: 'error', message: err?.message ?? 'Internal error' } as OutgoingMessage));
      } catch { /* ignore send failures */ }
    }
  });

  ws.addEventListener('close', () => {
    // History is already persisted in KV — nothing else to clean up
  });

  ws.addEventListener('error', () => {
    try { ws.close(1011, 'Unexpected error'); } catch { /* ignore */ }
  });
}
