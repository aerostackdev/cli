export default {
    async scheduled(event: any, env: any, ctx: any) {
        const pg = env.PG;
        console.log('Running cron task...');
        await pg.query("INSERT INTO logs (message) VALUES ('Cron job ran at ' || NOW())");
    },
};
