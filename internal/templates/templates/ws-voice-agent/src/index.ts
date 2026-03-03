import { sdk } from '@aerostack/sdk';

export interface Env {
  AEROSTACK_PROJECT_ID: string;
  AEROSTACK_API_KEY: string;
}

export type IncomingMessage =
  | { type: 'ping' }
  | { type: 'user_text'; text: string }
  | { type: 'clear' };

export type OutgoingMessage =
  | { type: 'pong' }
  | { type: 'assistant_text'; text: string }
  | { type: 'status'; historyLength: number }
  | { type: 'error'; message: string };

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === '/') {
      return new Response(
        'WebSocket Voice Agent is running. Connect via WebSocket to /ws and send { "type": "user_text", "text": "Hello" }.',
        { status: 200 },
      );
    }

    if (url.pathname === '/ws' && request.headers.get('Upgrade') === 'websocket') {
      sdk.init(env);

      const pair = new WebSocketPair();
      const [client, server] = Object.values(pair) as [WebSocket, WebSocket];

      handleSession(server);

      return new Response(null, {
        status: 101,
        webSocket: client,
      });
    }

    return new Response('Not found', { status: 404 });
  },
};

function handleSession(ws: WebSocket) {
  ws.accept();
  let history: { role: string; content: string }[] = [];

  // Send initial status on connection
  ws.send(JSON.stringify({ type: 'status', historyLength: history.length }));

  ws.addEventListener('message', async (event) => {
    try {
      const raw = typeof event.data === 'string' ? event.data : String(event.data);
      const msg = JSON.parse(raw) as IncomingMessage;

      if (msg.type === 'ping') {
        const response: OutgoingMessage = { type: 'pong' };
        ws.send(JSON.stringify(response));
        return;
      }

      if (msg.type === 'clear') {
        history = [];
        ws.send(JSON.stringify({ type: 'status', historyLength: history.length }));
        return;
      }

      if (msg.type === 'user_text') {
        history.push({ role: 'User', content: msg.text });

        // Keep last 10 messages in memory context
        if (history.length > 10) history = history.slice(history.length - 10);

        try {
          const context = history.map(h => `${h.role}: ${h.content}`).join('\n');
          const { text } = await sdk.ai.generate(`Use the following conversation history to respond to the user:\n${context}`);

          history.push({ role: 'Assistant', content: text });
          const response: OutgoingMessage = { type: 'assistant_text', text };
          ws.send(JSON.stringify(response));
        } catch (e: any) {
          ws.send(JSON.stringify({ type: 'error', message: e.message }));
        }
        return;
      }

      const unknown: OutgoingMessage = {
        type: 'error',
        message: 'Unknown message type',
      };
      ws.send(JSON.stringify(unknown));
    } catch (err: any) {
      const error: OutgoingMessage = {
        type: 'error',
        message: err?.message ?? 'Internal error',
      };
      try {
        ws.send(JSON.stringify(error));
      } catch {
        // ignore send failures
      }
    }
  });

  ws.addEventListener('close', () => {
    // Connection closed – nothing else to do for this simple example
  });

  ws.addEventListener('error', () => {
    try {
      ws.close(1011, 'Unexpected error');
    } catch {
      // ignore
    }
  });
}

