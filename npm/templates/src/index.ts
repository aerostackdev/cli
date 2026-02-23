/**
 * src/index.ts — Aerostack Base Project
 * 
 * This is the main entry point for your Cloudflare Worker.
 * Routes from installed modules are mounted here automatically
 * when you run `npx aerostack add <function-name>`.
 */
import { Hono } from 'hono';

// ─── Module Routes ─────────────────────────────────────────────────────────────
// aerostack:imports — imports are auto-injected above this line

const app = new Hono<{
    Bindings: {
        DB: D1Database;
        CACHE: KVNamespace;
        AI: Ai;
        [key: string]: unknown;
    }
}>();

// ─── Installed Module Routes ──────────────────────────────────────────────────
// aerostack:routes — routes are auto-injected above this line

// ─── Health Check ─────────────────────────────────────────────────────────────
app.get('/health', (c) => c.json({ status: 'ok', timestamp: Date.now() }));

export default app;
