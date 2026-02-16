import express from 'express';
import { sdk } from '@aerostack/sdk';

const app = express();
app.use(express.json());

// Workers environment middleware
app.use((req, res, next) => {
    // @ts-ignore
    const env = req.env || (req as any).env;
    if (env) sdk.init(env);
    next();
});

app.get('/', (req, res) => {
    res.send('Welcome to your Aerostack Express API with Neon PostgreSQL!');
});

app.get('/api/status', (req, res) => {
    res.json({
        status: 'online',
        database: 'Neon PostgreSQL',
        framework: 'Express.js'
    });
});

export default app;
