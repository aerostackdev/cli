## Aerostack Multi-Function Template

This project was generated from the **multi-func** template. It demonstrates how to structure a project with multiple entrypoints that share common utilities — a main API and a separate auth function that can be deployed independently.

### What's Included

```
.
├── api/index.ts          — Main Hono API (GET /, GET /api/data)
├── functions/auth-hook.ts — Auth function (POST /login)
└── shared/utils.ts        — Shared utilities used by both
```

- `aerostack.toml` routes `/auth/*` to the auth function and everything else to the main API.
- Both functions share code via `shared/utils.ts` — no duplication.
- Both functions use `sdk.init(c.env)` independently so bindings are available in each.

### 1. Configure Local Environment

1. Copy the example vars:

```bash
cp .dev.vars.example .dev.vars
```

2. Edit `.dev.vars` and fill in:

- `AEROSTACK_PROJECT_ID` — Your project ID from the Aerostack dashboard.
- `AEROSTACK_API_KEY` — API key for this project (keep this secret).

### 2. Run Locally

```bash
aerostack dev
```

Test the different entrypoints:

```bash
# Main API
curl http://localhost:8787/
curl http://localhost:8787/api/data

# Auth function (routes to functions/auth-hook.ts)
curl -X POST http://localhost:8787/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com"}'
```

### 3. Add Shared Code

Put utilities that multiple functions need in `shared/utils.ts` and import them:

```ts
// In api/index.ts or functions/auth-hook.ts
import { getGreeting } from '../shared/utils';
```

### 4. Deploy

```bash
aerostack deploy
```

### Extending

- **More functions**: Add new entrypoints under `functions/` and add `[[functions]]` entries to `aerostack.toml`.
- **Database**: Enable D1 or Neon in `aerostack.toml` and use `sdk.db` in any function.
- **Shared middleware**: Move common auth/logging middleware into `shared/` and reuse it across all functions.

### More Documentation

- Online docs: `https://docs.aerostack.dev`
