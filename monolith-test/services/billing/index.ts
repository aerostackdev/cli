// Service: billing
// Import shared code: import { getDb } from "@shared/db"

export default {
  async fetch(request: Request, env: Record<string, unknown>, ctx: ExecutionContext): Promise<Response> {
    return new Response("Hello from billing!");
  },
};
