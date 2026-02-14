# Deploy Flow: Your Cloudflare Account vs Aerostack

> **Product direction:** Aerostack-first. Users create projects on Aerostack, login/validate, and deploy to Aerostack's infrastructure. See `docs/planning/AEROSTACK_FIRST_DEPLOY_FLOW.md` for the target flow.

## Current Implementation: **Your Cloudflare Account** (Temporary)

`aerostack deploy` runs `wrangler deploy` under the hood. Wrangler uses **your** Cloudflare credentials:

- From `npx wrangler login` (browser OAuth)
- Or from `CLOUDFLARE_API_TOKEN` environment variable

**Result:** Workers, D1 databases, and bindings are created in **your Cloudflare account**. You own everything. Aerostack CLI does not use a separate "Aerostack account."

---

## Deploy Flow (Current)

```
┌─────────────────────────────────────────────────────────────────┐
│  aerostack deploy --env staging                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. Parse aerostack.toml                                         │
│  2. Generate wrangler.toml (with env.staging overrides)           │
│  3. Run: npx wrangler deploy --config wrangler.toml --env staging │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Wrangler uses YOUR credentials (from wrangler login)             │
│  → Deploys to YOUR Cloudflare account                             │
│  → Worker URL: https://<name>-staging.<your-subdomain>.workers.dev│
└─────────────────────────────────────────────────────────────────┘
```

---

## The Error You Saw

```
binding DB of type d1 must have a valid `id` specified [code: 10021]
```

**Cause:** `database_id` in `aerostack.toml` is still the placeholder `YOUR_STAGING_D1_ID`. Cloudflare expects a real D1 database UUID (e.g. `a1b2c3d4-e5f6-7890-abcd-ef1234567890`).

**Fix:**

1. Create a D1 database in your Cloudflare account:
   ```bash
   npx wrangler d1 create demo-api-db
   ```

2. Copy the `database_id` from the output (UUID format).

3. Update `aerostack.toml`:
   ```toml
   [[env.staging.d1_databases]]
   binding = "DB"
   database_name = "api-db"
   database_id = "<paste-the-uuid-here>"
   ```

4. Run deploy again:
   ```bash
   aerostack deploy --env staging
   ```

---

## Target: Aerostack-First (Preferred)

**Product vision:** Users deploy to **Aerostack**, not their own Cloudflare account.

| Step | Description |
|------|-------------|
| 1 | User creates project on Aerostack (dashboard) |
| 2 | User runs `aerostack link <project-id>` to link local code |
| 3 | User runs `aerostack login` to authenticate with Aerostack |
| 4 | User runs `aerostack deploy` → deploys to **Aerostack's infrastructure** |

**Requires:** Aerostack backend (auth, projects API, deploy API), Cloudflare API integration.

See `docs/planning/AEROSTACK_FIRST_DEPLOY_FLOW.md` for full plan.

## Fallback: Your Account (Current)

Until Aerostack backend exists, deploy uses `wrangler deploy` → **your Cloudflare account**. Optional future flag: `--use-own-account` for users who prefer their own Cloudflare.

---

## Summary

| Question | Answer |
|----------|--------|
| Where does deploy go? | **Your Cloudflare account** (via wrangler login) |
| Is there an Aerostack account? | No. Aerostack CLI does not deploy to a separate Aerostack account. |
| Why the D1 error? | `database_id` must be a real D1 UUID, not a placeholder. |
| How to fix? | Create D1 with `wrangler d1 create`, copy ID, update aerostack.toml. |
