import { Hono } from 'hono';
import { sdk } from '@aerostack/sdk';

const app = new Hono<{ Bindings: any }>();

app.use('*', async (c, next) => {
    sdk.init(c.env);
    await next();
});

app.post('/webhook', async (c) => {
    const payload = await c.req.json();

    await sdk.db.query('INSERT INTO webhooks (payload, received_at) VALUES ($1, NOW())', [JSON.stringify(payload)]);

    return c.json({ status: 'received' });
});

export default app;
