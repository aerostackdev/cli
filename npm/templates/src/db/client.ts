import { drizzle } from 'drizzle-orm/d1';
import * as schema from './schema';

export type DrizzleDB = ReturnType<typeof createDb>;

/**
 * Creates a Drizzle database client bound to the Cloudflare D1 instance.
 * Call this inside each request handler: const db = createDb(c.env.DB);
 */
export function createDb(d1: D1Database) {
    return drizzle(d1, { schema });
}
