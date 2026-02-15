export default {
    async fetch(request: Request, env: any, ctx: any): Promise<Response> {
        return new Response('Hello from a blank Neon worker!');
    },
};
