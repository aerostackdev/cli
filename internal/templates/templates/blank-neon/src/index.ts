import { sdk } from '@aerostack/sdk';

export default {
    async fetch(request: Request, env: any, ctx: any): Promise<Response> {
        sdk.init(env);
        return new Response('Hello from a blank Neon worker!');
    },
};
