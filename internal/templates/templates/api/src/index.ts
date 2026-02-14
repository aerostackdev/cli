import { Hono } from 'hono';

const app = new Hono();

app.get('/', (c) => c.text('Welcome to the Aerostack API!'));

app.get('/users/:id', async (c) => {
    const id = c.req.param('id');
    // In a real app, you'd fetch from env.DB
    return c.json({ id, name: `User ${id}`, role: 'developer' });
});

export default app;
