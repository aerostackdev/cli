import { Hono } from 'hono';
import { sdk } from '@aerostack/sdk';

const app = new Hono<{ Bindings: any }>();

app.use('*', async (c, next) => {
    sdk.init(c.env);
    await next();
});

app.get('/', (c) => {
    return c.text('Welcome to your Aerostack API with Neon PostgreSQL!');
});

app.get('/users', async (c) => {
    try {
        const result = await sdk.db.query('SELECT * FROM users');
        return c.json(result.results);
    } catch (e: any) {
        return c.json({ error: e.message }, 500);
    }
});

export default app;
