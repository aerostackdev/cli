# Aerostack dev vs raw Wrangler

This doc explains what `aerostack dev` gives users beyond `npx wrangler dev`.

## Strategy: B + C

- **B (Wrapper):** Single config, auto setup, consistent UX
- **C (Value-add):** Features Wrangler doesn't provide

---

## What we give today (B)

| Feature | aerostack dev | npx wrangler dev |
|---------|---------------|------------------|
| **Config** | `aerostack.toml` only | `wrangler.toml` (manual) |
| **D1 by default** | Yes — blank projects get `env.DB` | No — must add `[[d1_databases]]` |
| **Config translation** | Auto-generates wrangler.toml | N/A |
| **One config for init/dev/deploy** | Yes | No — wrangler is dev-only |

---

## What we're adding (C)

| Feature | Status | Notes |
|---------|--------|-------|
| **--remote staging** | Done | Pass-through to wrangler |
| **Branded startup** | Done | Clear "Aerostack dev" messaging |
| **Unified multi-service logs** | Phase 3 | Color-coded, trace IDs |
| **AI error detection** | Phase 4 | Self-healing suggestions |
| **Logic Lab hooks** | Phase 4 | No-code customization |
| **Custom log formatting** | Phase 3+ | Aerostack prefix, structured output |

---

## Phase status

- **Phase 1:** Done — CLI scaffolding, dev with D1 (Wrangler/Miniflare), aerostack.toml
- **Phase 2:** To plan — Multi-DB Super Bridge, Neon, typegen, migrations
- **Phase 3+:** Backlog — Logs, AI, Logic Lab (see Master Vision)

---

## User flow

```
aerostack init my-app     → Creates aerostack.toml
cd my-app
aerostack dev             → Generates wrangler.toml, runs wrangler, D1 ready
aerostack deploy          → Same config, deploys to Aerostack
```

vs raw Wrangler:

```
mkdir my-app && cd my-app
# Manually create wrangler.toml with d1_databases
npx wrangler dev
# Different config for deploy
```

---

## Implementation notes

- `aerostack dev` spawns `npx wrangler dev` (Miniflare)
- We generate `wrangler.toml` from `aerostack.toml` on each run
- `wrangler.toml` is gitignored — source of truth is `aerostack.toml`
