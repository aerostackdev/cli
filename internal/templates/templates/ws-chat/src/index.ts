import { sdk } from '@aerostack/sdk';

export interface Env {
  CACHE: any;
  AEROSTACK_PROJECT_ID: string;
  AEROSTACK_API_KEY: string;
}

export type IncomingMessage =
  | { type: 'ping' }
  | { type: 'join'; username: string }
  | { type: 'chat'; text: string };

export type OutgoingMessage =
  | { type: 'pong' }
  | { type: 'joined'; userId: string; username: string; roomId: string }
  | { type: 'chat'; from: string; username?: string; text: string; timestamp: number }
  | { type: 'system'; message: string }
  | { type: 'error'; message: string };

const rooms = new Map<string, Set<WebSocket>>();
const userIds = new WeakMap<WebSocket, string>();
const userNames = new WeakMap<WebSocket, string>();

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === '/') {
      return new Response(
        'WebSocket Chat is running. Connect via WebSocket to /ws/<roomId> for group chat.',
        { status: 200 },
      );
    }

    if (url.pathname.startsWith('/ws/') && request.headers.get('Upgrade') === 'websocket') {
      sdk.init(env);

      const roomId = url.pathname.split('/')[2] || 'default';

      const pair = new WebSocketPair();
      const [client, server] = Object.values(pair) as [WebSocket, WebSocket];

      handleSession(server, roomId);

      return new Response(null, {
        status: 101,
        webSocket: client,
      });
    }

    return new Response('Not found', { status: 404 });
  },
};

function handleSession(ws: WebSocket, roomId: string) {
  ws.accept();

  const room = getOrCreateRoom(roomId);
  const userId = crypto.randomUUID();

  // We assign a default username until they send 'join'
  const defaultName = `User ${userId.substring(0, 4)}`;

  room.add(ws);
  userIds.set(ws, userId);
  userNames.set(ws, defaultName);

  broadcast(roomId, {
    type: 'system',
    message: `${defaultName} joined room ${roomId}`,
  });

  // Fetch and send recent chat history from KV
  sdk.cache.get<{ from: string; username?: string; text: string; timestamp: number }[]>(`history_${roomId}`)
    .then(history => {
      if (history && history.length > 0) {
        for (const msg of history) {
          // Send history messages as chat messages individually
          ws.send(JSON.stringify({ ...msg, type: 'chat' }));
        }
      }
    }).catch(() => { /* ignore */ });

  ws.addEventListener('message', async (event) => {
    const id = userIds.get(ws);
    const username = userNames.get(ws) || defaultName;
    if (!id) return;

    try {
      const raw = typeof event.data === 'string' ? event.data : String(event.data);
      const msg = JSON.parse(raw) as IncomingMessage;

      if (msg.type === 'ping') {
        const response: OutgoingMessage = { type: 'pong' };
        ws.send(JSON.stringify(response));
        return;
      }

      if (msg.type === 'join') {
        userNames.set(ws, msg.username);
        ws.send(JSON.stringify({ type: 'joined', userId: id, username: msg.username, roomId }));
        broadcast(roomId, {
          type: 'system',
          message: `${msg.username} entered the chat`,
        });
        return;
      }

      if (msg.type === 'chat') {
        const chat: OutgoingMessage = {
          type: 'chat',
          from: id,
          username,
          text: msg.text,
          timestamp: Date.now()
        };
        broadcast(roomId, chat);

        // Save to KV history (in background to avoid blocking socket loop)
        try {
          const key = `history_${roomId}`;
          const history = (await sdk.cache.get<any[]>(key)) || [];
          history.push({ from: id, username, text: msg.text, timestamp: chat.timestamp });
          if (history.length > 20) history.shift(); // Keep last 20 messages
          await sdk.cache.set(key, history);
        } catch (e) {
          console.error("Failed to save chat history", e);
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
    const room = rooms.get(roomId);
    if (room) {
      room.delete(ws);
      if (room.size === 0) {
        rooms.delete(roomId);
      }
    }
    const username = userNames.get(ws) || `User ${userId.substring(0, 4)}`;
    broadcast(roomId, {
      type: 'system',
      message: `${username} left room ${roomId}`,
    });
  });

  ws.addEventListener('error', () => {
    try {
      ws.close(1011, 'Unexpected error');
    } catch {
      // ignore
    }
  });
}

function getOrCreateRoom(roomId: string): Set<WebSocket> {
  let room = rooms.get(roomId);
  if (!room) {
    room = new Set<WebSocket>();
    rooms.set(roomId, room);
  }
  return room;
}

function broadcast(roomId: string, message: OutgoingMessage) {
  const room = rooms.get(roomId);
  if (!room) return;

  const payload = JSON.stringify(message);
  for (const ws of room) {
    try {
      ws.send(payload);
    } catch {
      // ignore failed sends
    }
  }
}
