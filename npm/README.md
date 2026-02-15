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

## Uninstall

```bash
npm remove -g aerostack
```

This removes the package and the cached binary at `~/.aerostack/bin`.
