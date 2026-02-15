import { Hono } from 'hono';

const app = new Hono();

app.post('/webhook', async (c) => {
    const payload = await c.req.json();
    const pg = c.env.PG;

    await pg.query('INSERT INTO webhooks (payload, received_at) VALUES ($1, NOW())', [JSON.stringify(payload)]);

    return c.json({ status: 'received' });
});

export default app;
