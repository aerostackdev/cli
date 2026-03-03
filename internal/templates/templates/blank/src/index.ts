import { sdk } from '@aerostack/sdk';

export interface Env {
    DB: any; // D1Database
    CACHE: any; // KVNamespace
    QUEUE: any; // Queue
    AEROSTACK_PROJECT_ID: string;
    AEROSTACK_API_KEY: string;
}

export default {
    async fetch(request: Request, env: Env, ctx: any) {
        sdk.init(env);
        const url = new URL(request.url);

        if (url.pathname === '/') {
            return Response.json({ status: "ok", template: "blank" });
        }

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
        if (url.pathname === '/test/queue') {
            if (request.method === 'POST') {
                await sdk.queue.enqueue({ type: 'welcome_email', data: { userId: 123 } });
                return Response.json({ success: true, message: 'Job enqueued!' });
            }
            return Response.json({ message: "Use POST /test/queue to enqueue a job" });
        }

        return new Response("Not found", { status: 404 });
    },

    // Handle background queue tasks
    async queue(batch: any, env: Env) {
        sdk.init(env);
        for (const msg of batch.messages) {
            const body = msg.body as any;
            console.log(`Processing background job of type: ${body?.type || 'unknown'}`);
            if (body?.type === 'welcome_email') {
                console.log(`Sending welcome email to user ${body.data?.userId}`);
            }
            msg.ack();
        }
    }
};
