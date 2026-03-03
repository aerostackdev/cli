import { Hono } from 'hono';
import { sdk } from '@aerostack/sdk';

export interface Env {
    PG: any; // Neon Postgres connection
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

app.get('/', (c) => {
    return c.text('Welcome to your Aerostack API with Neon PostgreSQL! Try GET /health or /users');
});

app.get('/health', (c) => c.json({ status: "ok", template: "api-neon" }));

// ┌─────────────────────────────────────────────────────────┐
// │  CRUD Postgres Example                                   │
// └─────────────────────────────────────────────────────────┘
app.get('/users', async (c) => {
    try {
        await sdk.db.query('CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT)');
        const result = await sdk.db.query('SELECT * FROM users ORDER BY id DESC');
        return c.json(result.results);
    } catch (e: any) {
        return c.json({ error: e.message, hint: "Check DATABASE_URL in .dev.vars" }, 500);
    }
});

app.post('/users', async (c) => {
    try {
        const body: any = await c.req.json();
        if (!body.name) return c.json({ error: "Missing name" }, 400);

        await sdk.db.query('CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT)');
        await sdk.db.query('INSERT INTO users (name) VALUES ($1)', [body.name]);
        return c.json({ success: true, name: body.name }, 201);
    } catch (e: any) {
        return c.json({ error: e.message, hint: "Check DATABASE_URL in .dev.vars" }, 500);
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

// 3. Queue publish
test.post('/queue', async (c) => {
    await sdk.queue.enqueue({ type: 'welcome_email', data: { userId: 123 } });
    return c.json({ success: true, message: 'Job enqueued via Hono!' });
});

app.route('/test', test);

export default {
    fetch: app.fetch,
    async queue(batch: any, env: Env) {
        sdk.init(env);
        for (const msg of batch.messages) {
            const body = msg.body as any;
            console.log(`Neon template processed background job of type: ${body?.type || 'unknown'}`);
            if (body?.type === 'welcome_email') {
                console.log(`Sending welcome email to Neon user ${body.data?.userId}`);
            }
            msg.ack();
        }
    }
};
