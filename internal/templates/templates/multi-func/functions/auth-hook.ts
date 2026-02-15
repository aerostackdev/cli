import { Hono } from 'hono';

const app = new Hono();

app.post('/login', async (c) => {
    const body = await c.req.json();
    return c.json({
        status: 'success',
        user: { id: 1, email: body.email }
    });
});

export default app;
