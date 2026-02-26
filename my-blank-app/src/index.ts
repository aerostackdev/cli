export default {
    async fetch(request, env, ctx) {
        // env.DB is available when D1 is configured in aerostack.toml
        return new Response("Hello from your new Aerostack project!");
    },
};
