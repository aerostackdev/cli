// Shared database utilities — import via: import { getDb } from "@shared/db"
export function getDb(env: { DB?: unknown }) {
  return env.DB;
}
