import { sdk } from '@aerostack/sdk';

export default {
    async fetch(request: Request, env: any, ctx: any) {
        sdk.init(env);
        return new Response("Hello from your new Aerostack project!");
    },
};
