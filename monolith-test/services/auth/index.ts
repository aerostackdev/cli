// Service: auth
// Import shared code: import { getDb } from "@shared/db"
import { formatGreeting } from "@shared/utils";

export default {
  async fetch(request: Request, env: Record<string, unknown>, ctx: ExecutionContext): Promise<Response> {
    return new Response(formatGreeting("Auth Service"));
  },
};
