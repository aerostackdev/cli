import { sdk } from '@aerostack/sdk';

export interface Env {
  CACHE: any;
  AEROSTACK_PROJECT_ID: string;
  AEROSTACK_API_KEY: string;
}

export type IncomingMessage =
  | { type: 'ping' }
  | { type: 'join'; name: string }
  | { type: 'move'; x: number; y: number }
  | { type: 'chat'; text: string };

export type OutgoingMessage =
  | { type: 'pong' }
  | { type: 'state'; players: PlayerState[] }
  | { type: 'joined'; playerId: string; name: string; roomId: string }
  | { type: 'chat'; from: string; text: string }
  | { type: 'system'; message: string }
  | { type: 'error'; message: string };

export type PlayerState = {
  id: string;
  name?: string;
  roomId: string;
  x: number;
  y: number;
};

const rooms = new Map<string, Set<WebSocket>>();
const players = new WeakMap<WebSocket, PlayerState>();

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    sdk.init(env);
    const url = new URL(request.url);

    if (url.pathname === '/') {
      return new Response(
        'WebSocket Multiplayer Game is running.\n' +
        'Connect via WebSocket to /ws/<roomId>.\n' +
        'GET /rooms/<roomId> for current player list (useful before connecting).',
        { status: 200 },
      );
    }

    // HTTP snapshot endpoint — returns current player list from KV
    // Useful for clients to pre-load state before opening a WebSocket
    if (url.pathname.startsWith('/rooms/') && request.method === 'GET') {
      const roomId = url.pathname.split('/')[2] || 'default';
      const snapshot = await sdk.cache.get<PlayerState[]>(`room_${roomId}`);
      return Response.json({ roomId, players: snapshot ?? [] });
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

function handleSession(ws: WebSocket, roomId: string, ctx: ExecutionContext) {
  ws.accept();

  const room = getOrCreateRoom(roomId);
  const player: PlayerState = {
    id: crypto.randomUUID(),
    roomId,
    x: 0,
    y: 0,
  };

  room.add(ws);
  players.set(ws, player);

  broadcast(roomId, {
    type: 'system',
    message: `Player ${player.id.substring(0, 6)} joined room ${roomId} (${room.size} players)`,
  });
  sendState(roomId);
  persistRoomState(roomId, ctx);

  ws.addEventListener('message', (event) => {
    const current = players.get(ws);
    if (!current) return;

    try {
      const raw = typeof event.data === 'string' ? event.data : String(event.data);
      const msg = JSON.parse(raw) as IncomingMessage;

      if (msg.type === 'ping') {
        ws.send(JSON.stringify({ type: 'pong' } as OutgoingMessage));
        return;
      }

      if (msg.type === 'join') {
        current.name = msg.name;
        ws.send(JSON.stringify({ type: 'joined', playerId: current.id, name: current.name, roomId } as OutgoingMessage));
        broadcast(roomId, { type: 'system', message: `${current.name} entered the room` });
        sendState(roomId);
        persistRoomState(roomId, ctx);
        return;
      }

      if (msg.type === 'move') {
        current.x = msg.x;
        current.y = msg.y;
        sendState(roomId);
        persistRoomState(roomId, ctx);
        return;
      }

      if (msg.type === 'chat') {
        broadcast(roomId, { type: 'chat', from: current.id, text: msg.text } as OutgoingMessage);
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
    const displayName = player.name || `Player ${player.id.substring(0, 6)}`;
    broadcast(roomId, {
      type: 'system',
      message: `${displayName} left room ${roomId} (${room?.size || 0} players remaining)`,
    });
    sendState(roomId);
    persistRoomState(roomId, ctx);
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

function getRoomPlayers(roomId: string): PlayerState[] {
  const room = rooms.get(roomId);
  if (!room) return [];

  const result: PlayerState[] = [];
  for (const ws of room) {
    const state = players.get(ws);
    if (state) result.push({ ...state });
  }
  return result;
}

function sendState(roomId: string) {
  const room = rooms.get(roomId);
  if (!room) return;

  const state: OutgoingMessage = { type: 'state', players: getRoomPlayers(roomId) };
  const payload = JSON.stringify(state);
  for (const ws of room) {
    try { ws.send(payload); } catch { /* ignore */ }
  }
}

function broadcast(roomId: string, message: OutgoingMessage) {
  const room = rooms.get(roomId);
  if (!room) return;

  const payload = JSON.stringify(message);
  for (const ws of room) {
    try { ws.send(payload); } catch { /* ignore */ }
  }
}

// Persist room snapshot to KV in the background (non-blocking)
function persistRoomState(roomId: string, ctx: ExecutionContext) {
  ctx.waitUntil(
    sdk.cache.set(`room_${roomId}`, getRoomPlayers(roomId)).catch(() => { /* ignore */ })
  );
}
