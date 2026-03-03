import { sdk } from '@aerostack/sdk';

export interface Env {
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
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === '/') {
      return new Response(
        'WebSocket Multiplayer Game is running. Connect via WebSocket to /ws/<roomId>.',
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

  ws.addEventListener('message', (event) => {
    const current = players.get(ws);
    if (!current) return;

    try {
      const raw = typeof event.data === 'string' ? event.data : String(event.data);
      const msg = JSON.parse(raw) as IncomingMessage;

      if (msg.type === 'ping') {
        const response: OutgoingMessage = { type: 'pong' };
        ws.send(JSON.stringify(response));
        return;
      }

      if (msg.type === 'join') {
        current.name = msg.name;
        ws.send(JSON.stringify({ type: 'joined', playerId: current.id, name: current.name, roomId }));
        broadcast(roomId, {
          type: 'system',
          message: `${current.name} entered the room`,
        });
        sendState(roomId);
        return;
      }

      if (msg.type === 'move') {
        current.x = msg.x;
        current.y = msg.y;
        sendState(roomId);
        return;
      }

      if (msg.type === 'chat') {
        const chat: OutgoingMessage = {
          type: 'chat',
          from: current.id,
          text: msg.text,
        };
        broadcast(roomId, chat);
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
        // ignore
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
    const displayName = player.name || `Player ${player.id.substring(0, 6)}`;
    broadcast(roomId, {
      type: 'system',
      message: `${displayName} left room ${roomId} (${room?.size || 0} players remaining)`,
    });
    sendState(roomId);
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

function getRoomPlayers(roomId: string): PlayerState[] {
  const room = rooms.get(roomId);
  if (!room) return [];

  const result: PlayerState[] = [];
  for (const ws of room) {
    const state = players.get(ws);
    if (state) {
      result.push({ ...state });
    }
  }
  return result;
}

function sendState(roomId: string) {
  const room = rooms.get(roomId);
  if (!room) return;

  const state: OutgoingMessage = {
    type: 'state',
    players: getRoomPlayers(roomId),
  };

  const payload = JSON.stringify(state);
  for (const ws of room) {
    try {
      ws.send(payload);
    } catch {
      // ignore failed sends
    }
  }
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

