import { Hono } from 'hono';
import { sdk } from '@aerostack/sdk';

const app = new Hono<{ Bindings: any }>();

app.use('*', async (c, next) => {
    sdk.init(c.env);
    await next();
});

app.get('/', (c) => {
    return c.text('Welcome to your Aerostack API with Neon PostgreSQL! Try /users or /test/ai');
});

// Postgres Example
app.get('/users', async (c) => {
    try {
        const result = await sdk.db.query('SELECT * FROM users');
        return c.json(result.results);
    } catch (e: any) {
        return c.json({ error: e.message }, 500);
    }
});

// ┌─────────────────────────────────────────────────────────┐
// │  Aerostack Feature Examples                              │
// └─────────────────────────────────────────────────────────┘
const test = new Hono();

// 1. Cache (KV)
test.get('/cache', async (c) => {
    await sdk.cache.set('neon_last_query', new Date().toISOString());
    return c.json({ success: true, lastQuery: await sdk.cache.get('neon_last_query') });
});

// 2. AI (Proxy)
test.get('/ai', async (c) => {
    const { text } = await sdk.ai.generate('Explain database indexing in 20 words');
    return c.json({ success: true, aiResponse: text });
});

app.route('/test', test);

export default {
    fetch: app.fetch,
    async queue(batch: any, env: any) {
        sdk.init(env);
        console.log("Background job processed in Neon template");
        for (const msg of batch.messages) msg.ack();
    }
};
