import { Hono } from 'hono';
import { sdk } from '@aerostack/sdk';

const app = new Hono<{ Bindings: any }>();

app.use('*', async (c, next) => {
    sdk.init(c.env);
    await next();
});

app.get('/', (c) => c.text('Welcome to the Aerostack API! Try /test/db or /test/ai'));

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
    await sdk.cache.set('api_hit', Date.now());
    return c.json({ success: true, lastHit: await sdk.cache.get('api_hit') });
});

// 3. AI - Llama 3 Proxy
test.get('/ai', async (c) => {
    const { text } = await sdk.ai.generate('Say "Aerostack is awesome" in French');
    return c.json({ success: true, translation: text });
});

app.route('/test', test);

app.get('/users/:id', async (c) => {
    const id = c.req.param('id');
    return c.json({ id, name: `User ${id}`, role: 'developer' });
});

export default {
    fetch: app.fetch,
    // Add Queue support to the Hono template
    async queue(batch: any, env: any) {
        sdk.init(env);
        console.log("Processing background jobs...");
        for (const msg of batch.messages) {
            msg.ack();
        }
    }
};
