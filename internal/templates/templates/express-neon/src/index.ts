import express from 'express';

const app = express();

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
