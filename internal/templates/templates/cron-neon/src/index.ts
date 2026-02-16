import { sdk } from '@aerostack/sdk';

export default {
    async scheduled(event: any, env: any, ctx: any) {
        sdk.init(env);
        console.log('Running cron task...');
        await sdk.db.query("INSERT INTO logs (message) VALUES ('Cron job ran at ' || NOW())");
    },
};
