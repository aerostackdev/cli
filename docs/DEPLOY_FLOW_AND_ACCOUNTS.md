# Deploy Flow: Your Cloudflare Account vs Aerostack

> **Product direction:** Aerostack-first. Users create projects on Aerostack, login/validate, and deploy to Aerostack's infrastructure. See `docs/planning/AEROSTACK_FIRST_DEPLOY_FLOW.md` for the target flow.

## Current Implementation: **Aerostack Platform**

`aerostack deploy` now natively deploys your code directly to the **Aerostack Cloudflare infrastructure**.

- Uses your **Aerostack Project API Key** (or browser session)
- Automatically provisions a **D1 Database** (isolated by project ID)
- Automatically provisions a **KV Cache** namespace
- Does **not** require you to configure `database_id` values in your `aerostack.toml`.

---

## Deploy Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  aerostack deploy --env staging                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. Build your worker code locally using esbuild                 │
│  2. Send multipart form payload to Aerostack CLI API             │
│  3. Aerostack API attaches native D1, KV, and AI bindings        │
│  4. Aerostack API orchestrates Cloudflare API deployment         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Worker URL: https://<project>.<slug>.aerocall.ai/custom/...     │
└─────────────────────────────────────────────────────────────────┘
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
| Where does deploy go? | **Aerostack platform infrastructure** |
| Do I need Cloudflare? | No. Aerostack manages everything. |
| Do I need `database_id`? | No, the CLI API auto-provisions D1 per project. |
