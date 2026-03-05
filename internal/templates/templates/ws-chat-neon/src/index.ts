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

// Max messages returned as history replay on join
const HISTORY_LIMIT = 50;

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    sdk.init(env);
    const url = new URL(request.url);

    if (url.pathname === '/') {
      return new Response(
        'WebSocket Chat (Neon) is running. Connect via WebSocket to /ws/<roomId>.\n' +
        'GET /rooms/<roomId>/history  — fetch message history',
        { status: 200 },
      );
    }

    // HTTP endpoint to retrieve message history for a room
    if (url.pathname.match(/^\/rooms\/[^/]+\/history$/) && request.method === 'GET') {
      const roomId = url.pathname.split('/')[2];
      await ensureSchema();
      const { results } = await sdk.db.query(
        'SELECT id, user_id, username, text, created_at FROM messages WHERE room_id = $1 ORDER BY id DESC LIMIT $2',
        [roomId, String(HISTORY_LIMIT)],
      );
      return Response.json({ roomId, messages: results.reverse() });
    }

    if (url.pathname.startsWith('/ws/') && request.headers.get('Upgrade') === 'websocket') {
      const roomId = url.pathname.split('/')[2] || 'default';

      const pair = new WebSocketPair();
      const [client, server] = Object.values(pair) as [WebSocket, WebSocket];

      handleSession(server, roomId, ctx);

      return new Response(null, {
        status: 101,
        webSocket: client,
      });
    }

    return new Response('Not found', { status: 404 });
  },
};

async function ensureSchema() {
  await sdk.db.query(`
    CREATE TABLE IF NOT EXISTS messages (
      id         BIGSERIAL PRIMARY KEY,
      room_id    TEXT        NOT NULL,
      user_id    TEXT        NOT NULL,
      username   TEXT        NOT NULL,
      text       TEXT        NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    )
  `);
  await sdk.db.query(
    'CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages (room_id, id DESC)',
  );
}

function handleSession(ws: WebSocket, roomId: string, ctx: ExecutionContext) {
  ws.accept();

  const room = getOrCreateRoom(roomId);
  const userId = crypto.randomUUID();
  const defaultName = `User ${userId.substring(0, 4)}`;

  room.add(ws);
  userIds.set(ws, userId);
  userNames.set(ws, defaultName);

  broadcast(roomId, { type: 'system', message: `${defaultName} joined room ${roomId}` });

  // Replay last HISTORY_LIMIT messages from Neon
  sdk.db
    .query(
      'SELECT user_id, username, text, extract(epoch from created_at)*1000 AS ts FROM messages WHERE room_id = $1 ORDER BY id DESC LIMIT $2',
      [roomId, String(HISTORY_LIMIT)],
    )
    .then(({ results }) => {
      for (const row of [...results].reverse()) {
        ws.send(
          JSON.stringify({
            type: 'chat',
            from: row.user_id as string,
            username: row.username as string,
            text: row.text as string,
            timestamp: Number(row.ts),
          } as OutgoingMessage),
        );
      }
    })
    .catch(() => { /* ignore */ });

  ws.addEventListener('message', async (event) => {
    const id = userIds.get(ws);
    const username = userNames.get(ws) || defaultName;
    if (!id) return;

    try {
      const raw = typeof event.data === 'string' ? event.data : String(event.data);
      const msg = JSON.parse(raw) as IncomingMessage;

      if (msg.type === 'ping') {
        ws.send(JSON.stringify({ type: 'pong' } as OutgoingMessage));
        return;
      }

      if (msg.type === 'join') {
        userNames.set(ws, msg.username);
        ws.send(JSON.stringify({ type: 'joined', userId: id, username: msg.username, roomId } as OutgoingMessage));
        broadcast(roomId, { type: 'system', message: `${msg.username} entered the chat` });
        return;
      }

      if (msg.type === 'chat') {
        const ts = Date.now();
        const chat: OutgoingMessage = { type: 'chat', from: id, username, text: msg.text, timestamp: ts };
        broadcast(roomId, chat);

        // Persist message to Neon in background
        ctx.waitUntil((async () => {
          try {
            await ensureSchema();
            await sdk.db.query(
              'INSERT INTO messages (room_id, user_id, username, text) VALUES ($1, $2, $3, $4)',
              [roomId, id, username, msg.text],
            );
          } catch (e) {
            console.error('Failed to persist message', e);
          }
        })());
        return;
      }

      ws.send(JSON.stringify({ type: 'error', message: 'Unknown message type' } as OutgoingMessage));
    } catch (err: any) {
      try {
        ws.send(JSON.stringify({ type: 'error', message: err?.message ?? 'Internal error' } as OutgoingMessage));
      } catch { /* ignore */ }
    }
  });

  ws.addEventListener('close', () => {
    const room = rooms.get(roomId);
    if (room) {
      room.delete(ws);
      if (room.size === 0) rooms.delete(roomId);
    }
    const username = userNames.get(ws) || `User ${userId.substring(0, 4)}`;
    broadcast(roomId, { type: 'system', message: `${username} left room ${roomId}` });
  });

  ws.addEventListener('error', () => {
    try { ws.close(1011, 'Unexpected error'); } catch { /* ignore */ }
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
    try { ws.send(payload); } catch { /* ignore */ }
  }
}
