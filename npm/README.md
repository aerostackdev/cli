# aerostack

Aerostack CLI - Zero-config serverless development for Cloudflare.

## Install (recommended: npx — no PATH config)

```bash
npx aerostack init my-app
```

Works immediately. No global install or PATH setup.

## Global install (optional)

```bash
npm install -g aerostack
```

If `aerostack` not found after install, the postinstall will print a one-liner to add to PATH. Or use `npx aerostack` instead.

## Usage

```bash
npx aerostack init my-app
cd my-app
npx aerostack dev
npx aerostack deploy
```

The first run downloads the binary from GitHub releases. Subsequent runs use the cached binary at `~/.aerostack/bin` (shared with the curl install).

## Registry: add community functions

Install open-source functions from the Aerostack registry into your project (Cloudflare or Node.js):

```bash
npx aerostack add stripe-checkout
npx aerostack add alice/stripe-checkout
npx aerostack add stripe-checkout --runtime=node
```

- **Cloudflare** (default): Adds Hono + D1 adapter and wires the route. Uses Drizzle in `src/db/` when the function has a schema.
- **Node.js** (`--runtime=node`): Adds a Node/Express adapter. Pass your own Drizzle client (pg or sqlite) and mount the router in your app.

Functions follow the **Open Function Standard**: one portable core, multiple adapters. You can use them in any Node.js project, not only on the Aerostack platform. See `planning/OPEN_FUNCTION_STANDARD.md` for the spec.

## Uninstall

```bash
npm remove -g aerostack
```

This removes the package and the cached binary at `~/.aerostack/bin`.
