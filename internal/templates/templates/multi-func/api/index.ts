import { Hono } from 'hono';
import { getGreeting } from '../shared/utils';

const app = new Hono();

app.get('/', (c) => {
    return c.text(getGreeting('Aerostack Multi-Function User'));
});

app.get('/api/data', (c) => {
    return c.json({
        message: 'This data comes from the main API',
        timestamp: new Date().toISOString()
    });
});

export default app;
