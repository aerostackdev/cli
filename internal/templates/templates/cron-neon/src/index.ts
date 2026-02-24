import { sdk } from '@aerostack/sdk';

export default {
    async scheduled(event: any, env: any, ctx: any) {
        sdk.init(env);
        console.log('Running Aerostack cron task...');

        // 1. Database - Store execution log
        await sdk.db.query("INSERT INTO cron_logs (message) VALUES ('Cron job ran at ' || NOW())");

        // 2. Cache - Track status
        await sdk.cache.set('last_cron_run', new Date().toISOString());

        // 3. AI - Optionally use AI for periodic analysis/tasks
        // const res = await sdk.ai.generate('Summarize recent logs');
    },

    // Optional: Add fetch handler for manual trigger or health checks
    async fetch(request: Request, env: any) {
        sdk.init(env);
        return new Response("Cron worker is active.");
    }
};
