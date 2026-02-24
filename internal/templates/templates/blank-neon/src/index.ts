import { sdk } from '@aerostack/sdk';

export default {
    async fetch(request: Request, env: any, ctx: any): Promise<Response> {
        sdk.init(env);
        const url = new URL(request.url);

        // ┌─────────────────────────────────────────────────────────┐
        // │  Aerostack Feature Examples (Neon Postgres Edition)      │
        // └─────────────────────────────────────────────────────────┘

        // 1. Postgres Query
        if (url.pathname === '/test/db') {
            const { results } = await sdk.db.query('SELECT NOW() as time');
            return Response.json({ success: true, time: results[0].time });
        }

        // 2. Cache (KV)
        if (url.pathname === '/test/cache') {
            await sdk.cache.set('last_neon_hit', Date.now());
            return Response.json({ success: true, lastHit: await sdk.cache.get('last_neon_hit') });
        }

        // 3. AI
        if (url.pathname === '/test/ai') {
            const { text } = await sdk.ai.generate('Explain serverless in 10 words');
            return Response.json({ success: true, explanation: text });
        }

        return new Response('Hello from a blank Neon worker! Try /test/db, /test/cache, or /test/ai');
    },

    async queue(batch: any, env: any) {
        sdk.init(env);
        for (const msg of batch.messages) msg.ack();
    }
};
