# aerostack

Aerostack CLI - Zero-config serverless development for Cloudflare.

## Install

**npm**
```bash
npm install -g aerostack
```

**pnpm**
```bash
pnpm add -g aerostack
```

**yarn**
```bash
yarn global add aerostack
```

**bun**
```bash
bun install -g aerostack
```

**npx (no install)**
```bash
npx aerostack init my-app
```

## Usage

```bash
aerostack init my-app
cd my-app
aerostack dev
aerostack deploy
```

The first run downloads the binary from GitHub releases. Subsequent runs use the cached binary at `~/.aerostack/bin` (shared with the curl install).
