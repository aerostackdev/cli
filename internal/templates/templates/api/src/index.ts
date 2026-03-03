import { Hono } from 'hono';
import { sdk } from '@aerostack/sdk';

export interface Env {
    DB: any; // D1Database
    CACHE: any; // KVNamespace
    QUEUE: any; // Queue
    AEROSTACK_PROJECT_ID: string;
    AEROSTACK_API_KEY: string;
}

const app = new Hono<{ Bindings: Env }>();

app.use('*', async (c, next) => {
    sdk.init(c.env);
    await next();
});

app.get('/', (c) => c.text('Welcome to the Aerostack API! Try GET /health or /notes'));

app.get('/health', (c) => c.json({ status: "ok", template: "api" }));

// ┌─────────────────────────────────────────────────────────┐
// │  CRUD Example (D1)                                       │
// └─────────────────────────────────────────────────────────┘
app.get('/notes', async (c) => {
    await sdk.db.query('CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, text TEXT)');
    const { results } = await sdk.db.query('SELECT * FROM notes ORDER BY id DESC');
    return c.json(results);
});

app.post('/notes', async (c) => {
    const body: any = await c.req.json();
    if (!body.text) return c.json({ error: "Missing text" }, 400);

    await sdk.db.query('CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, text TEXT)');
    await sdk.db.query('INSERT INTO notes (text) VALUES (?)', [body.text]);
    return c.json({ success: true, text: body.text }, 201);
});

// ┌─────────────────────────────────────────────────────────┐
// │  Aerostack Feature Examples                              │
// │  Hit these endpoints to test your local configuration   │
// └─────────────────────────────────────────────────────────┘
const test = new Hono();

// 1. Database - Persist data locally
test.get('/db', async (c) => {
    await sdk.db.query('CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, text TEXT)');
    await sdk.db.query('INSERT INTO notes (text) VALUES (?)', ['Hono + Aerostack!']);
    const { results } = await sdk.db.query('SELECT * FROM notes');
    return c.json({ success: true, notes: results });
});

// 2. Cache - Key-value storage
test.get('/cache', async (c) => {
    await sdk.cache.set('api_hit', new Date().toISOString());
    return c.json({ success: true, lastHit: await sdk.cache.get('api_hit') });
});

// 3. AI - Llama 3 Proxy
test.get('/ai', async (c) => {
    const { text } = await sdk.ai.generate('Say "Aerostack is awesome" in French');
    return c.json({ success: true, translation: text });
});

// 4. Queue publish
test.post('/queue', async (c) => {
    await sdk.queue.enqueue({ type: 'welcome_email', data: { userId: 123 } });
    return c.json({ success: true, message: 'Job enqueued via Hono!' });
});

app.route('/test', test);

app.get('/users/:id', async (c) => {
    const id = c.req.param('id');
    return c.json({ id, name: `User ${id}`, role: 'developer' });
});

export default {
    fetch: app.fetch,
    // Handle background queue tasks
    async queue(batch: any, env: Env) {
        sdk.init(env);
        for (const msg of batch.messages) {
            const body = msg.body as any;
            console.log(`Processing background job of type: ${body?.type || 'unknown'}`);
            if (body?.type === 'welcome_email') {
                console.log(`Sending welcome email to user ${body.data?.userId}`);
            }
            msg.ack();
        }
    }
};
