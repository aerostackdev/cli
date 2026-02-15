import { Hono } from 'hono';

const app = new Hono();

app.get('/', (c) => {
    return c.text('Welcome to your Aerostack API with Neon PostgreSQL!');
});

app.get('/users', async (c) => {
    // @ts-ignore
    const pg = c.env.PG;
    try {
        const result = await pg.query('SELECT * FROM users');
        return c.json(result.rows);
    } catch (e: any) {
        return c.json({ error: e.message }, 500);
    }
});

export default app;
