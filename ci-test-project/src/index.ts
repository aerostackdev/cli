import { sdk } from '@aerostack/sdk';

export default {
    async fetch(request: Request, env: any, ctx: any) {
        sdk.init(env);
        const url = new URL(request.url);

        // ┌─────────────────────────────────────────────────────────┐
        // │  Aerostack Feature Examples                              │
        // │  Hit these endpoints to test your local configuration   │
        // └─────────────────────────────────────────────────────────┘

        // 1. Database (D1) - Persist data locally
        if (url.pathname === '/test/db') {
            await sdk.db.query('CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, text TEXT)');
            await sdk.db.query('INSERT INTO notes (text) VALUES (?)', ['Hello from Aerostack!']);
            const { results } = await sdk.db.query('SELECT * FROM notes');
            return Response.json({ success: true, notes: results });
        }

        // 2. Cache (KV) - High-performance key-value storage
        if (url.pathname === '/test/cache') {
            await sdk.cache.set('last_visit', new Date().toISOString());
            const lastVisit = await sdk.cache.get('last_visit');
            return Response.json({ success: true, lastVisit });
        }

        // 3. AI - Use Llama 3 or other models via Proxy
        if (url.pathname === '/test/ai') {
            const { text } = await sdk.ai.generate('Tell me a 1-sentence joke about coding');
            return Response.json({ success: true, joke: text });
        }

        // 4. Queues - Background processing
        if (url.pathname === '/test/queue' && request.method === 'POST') {
            await sdk.queue.enqueue({ type: 'welcome_email', data: { userId: 123 } });
            return Response.json({ success: true, message: 'Job enqueued!' });
        }

        return new Response("Welcome to Aerostack! Try /test/db, /test/cache, /test/ai, or POST to /test/queue");
    },

    // Handle background queue tasks
    async queue(batch: any, env: any) {
        sdk.init(env);
        for (const msg of batch.messages) {
            console.log("Processing background job:", msg.body);
            msg.ack();
        }
    }
};
