# Aerostack CLI Install Script

One-command install for the Aerostack CLI. The script shows branded output, download progress (when run in a terminal), and clear next steps after install or upgrade.

## Usage

```bash
# Install latest
curl -fsSL https://get.aerostack.dev | sh

# Install specific version
curl -fsSL https://get.aerostack.dev | VERSION=1.0.0 sh
```

## Deploying get.aerostack.dev

Host `install.sh` at `https://get.aerostack.dev` so the curl command works.

### Option 1: Cloudflare Pages

1. Create a Pages project
2. Set root to `install/` or configure to serve `install.sh` at `/`
3. Add custom domain `get.aerostack.dev`
4. Ensure `install.sh` is served with `Content-Type: text/plain` or `application/x-sh`

### Option 2: Cloudflare Worker (recommended)

A Worker is provided at `workers/get-aerostack/`. Deploy with:

```bash
cd workers/get-aerostack && npm install && npm run deploy
```

This worker is configured to serve `install.sh` (via redirect to GitHub Raw) at `https://get.aerostack.dev`.


### Option 3: Redirect to GitHub raw

Configure `get.aerostack.dev` to redirect (301) to:
`https://raw.githubusercontent.com/aerostackdev/cli/main/install/install.sh`
