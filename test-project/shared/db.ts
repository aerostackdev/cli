// Shared database utilities — import via: import { getDb } from "@shared/db"
// Use @shared/* for code shared across services (tree-shaken per deploy)

export function getDb(env: { DB?: unknown }) {
  return env.DB;
}
