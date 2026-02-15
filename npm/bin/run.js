#!/usr/bin/env node
/**
 * Aerostack CLI - npm/pnpm/yarn wrapper
 * Downloads the Go binary from GitHub releases on first run.
 */

const { spawn } = require("child_process");
const fs = require("fs");
const path = require("path");

const REPO = "aerostackdev/cli";
const BINARY = "aerostack";
const INSTALL_DIR = process.env.AEROSTACK_INSTALL_DIR || path.join(process.env.HOME || process.env.USERPROFILE, ".aerostack", "bin");

function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;
  const map = {
    darwin: { arm64: "darwin_arm64", x64: "darwin_amd64" },
    linux: { arm64: "linux_arm64", x64: "linux_amd64" },
    win32: { arm64: "windows_arm64", x64: "windows_amd64" },
  };
  const p = map[platform]?.[arch];
  if (!p) throw new Error(`Unsupported platform: ${platform}-${arch}`);
  return { platform, arch, asset: p };
}

async function getLatestVersion() {
  const res = await fetch(`https://api.github.com/repos/${REPO}/releases/latest`, {
    headers: { "User-Agent": "aerostack-cli-npm/1.0" },
  });
  if (!res.ok) throw new Error("Failed to fetch latest version");
  const data = await res.text();
  const m = data.match(/"tag_name":\s*"v([^"]+)"/);
  if (!m) throw new Error("Could not parse version");
  return m[1];
}

async function downloadBinary(version, asset) {
  const ext = asset.startsWith("windows") ? "zip" : "tar.gz";
  const archive = `${BINARY}_${version}_${asset}.${ext}`;
  const url = `https://github.com/${REPO}/releases/download/v${version}/${archive}`;

  const tmpDir = path.join(require("os").tmpdir(), `aerostack-${Date.now()}`);
  fs.mkdirSync(tmpDir, { recursive: true });
  const archivePath = path.join(tmpDir, archive);

  const res = await fetch(url, {
    headers: { "User-Agent": "aerostack-cli-npm/1.0" },
    redirect: "follow",
  });
  if (!res.ok) throw new Error(`Download failed: ${url}`);
  const buffer = Buffer.from(await res.arrayBuffer());
  fs.writeFileSync(archivePath, buffer);

  const binDir = INSTALL_DIR;
  fs.mkdirSync(binDir, { recursive: true });
  const binPath = path.join(binDir, process.platform === "win32" ? `${BINARY}.exe` : BINARY);

  if (ext === "zip") {
    const AdmZip = require("adm-zip");
    const zip = new AdmZip(archivePath);
    zip.extractAllTo(tmpDir, true);
    const extracted = path.join(tmpDir, `${BINARY}.exe`);
    fs.renameSync(extracted, binPath);
  } else {
    const tar = require("tar");
    await tar.x({ file: archivePath, cwd: tmpDir });
    fs.renameSync(path.join(tmpDir, BINARY), binPath);
  }

  fs.chmodSync(binPath, 0o755);
  fs.rmSync(tmpDir, { recursive: true, force: true });
  return binPath;
}

async function getInstalledVersion(binPath) {
  try {
    const { execFileSync } = require("child_process");
    const out = execFileSync(binPath, ["--version"], { encoding: "utf8" });
    const m = out.match(/v?(\d+\.\d+\.\d+)/);
    return m ? m[1] : null;
  } catch {
    return null;
  }
}

async function ensureBinary() {
  const binPath = path.join(INSTALL_DIR, process.platform === "win32" ? `${BINARY}.exe` : BINARY);
  const { asset } = getPlatform();
  const latestVersion = await getLatestVersion();

  if (fs.existsSync(binPath)) {
    const installed = await getInstalledVersion(binPath);
    if (installed && installed === latestVersion) return binPath;
    if (installed) console.error(`Updating Aerostack CLI v${installed} → v${latestVersion}...`);
  } else {
    console.error(`Downloading Aerostack CLI v${latestVersion}...`);
  }

  return downloadBinary(latestVersion, asset);
}

async function main() {
  try {
    const binPath = await ensureBinary();
    const child = spawn(binPath, process.argv.slice(2), {
      stdio: "inherit",
      shell: false,
    });
    child.on("exit", (code) => process.exit(code ?? 0));
  } catch (err) {
    console.error("Error:", err.message);
    process.exit(1);
  }
}

main();
