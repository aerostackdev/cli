import { Hono } from 'hono';
import { sdk } from '@aerostack/sdk';

const app = new Hono<{ Bindings: any }>();

app.use('*', async (c, next) => {
    sdk.init(c.env);
    await next();
});

app.post('/login', async (c) => {
    const body = await c.req.json();
    return c.json({
        status: 'success',
        user: { id: 1, email: body.email }
    });
});

export default app;
