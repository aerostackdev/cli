import { Hono } from 'hono';
import { sdk } from '@aerostack/sdk';

const app = new Hono<{ Bindings: any }>();

app.use('*', async (c, next) => {
    sdk.init(c.env);
    await next();
});

app.get('/', (c) => c.text('Welcome to the Aerostack API!'));

app.get('/users/:id', async (c) => {
    const id = c.req.param('id');
    // In a real app, you'd fetch from env.DB
    return c.json({ id, name: `User ${id}`, role: 'developer' });
});

export default app;
