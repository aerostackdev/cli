#!/usr/bin/env node
/**
 * Aerostack CLI - Hybrid Dispatcher
 * 
 * Routes specific commands (init, add, publish, list, login) to the
 * pure Node.js implementation (drizzle/hono support).
 * 
 * All other commands map to the Go binary (legacy/platform).
 */

import { spawn } from 'child_process';
import { existsSync, mkdirSync, chmodSync, writeFileSync, rmSync, renameSync } from 'fs';
import { join, resolve } from 'path';
import { platform as getOsPlatform, arch as getOsArch, tmpdir } from 'os';
import { fileURLToPath } from 'url';

const __dirname = fileURLToPath(new URL('.', import.meta.url));

// ─── 1. Hybrid Dispatcher ──────────────────────────────────────────────────────

const args = process.argv.slice(2);
const command = args[0];
const NODE_COMMANDS = ['init', 'add', 'publish', 'list', 'login'];

if (command && NODE_COMMANDS.includes(command)) {
  // Dispatch to Node.js implementation
  import('../dist/index.js').then(m => m.run(args)).catch(err => {
    console.error('Failed to run Node.js CLI:', err);
    process.exit(1);
  });
} else {
  // Dispatch to Go implementation
  ensureBinary().then(binPath => {
    const child = spawn(binPath, args, { stdio: 'inherit' });
    child.on('exit', code => process.exit(code ?? 0));
  }).catch(err => {
    console.error('Error:', err.message);
    process.exit(1);
  });
}

// ─── 2. Go Binary Downloader (Legacy) ──────────────────────────────────────────

const REPO = "aerostackdev/cli";
const BINARY = "aerostack";
// Use HOME or USERPROFILE for global install location
const HOME = process.env.HOME || process.env.USERPROFILE || tmpdir();
const INSTALL_DIR = process.env.AEROSTACK_INSTALL_DIR || join(HOME, ".aerostack", "bin");

function getPlatform() {
  const platform = getOsPlatform();
  const arch = getOsArch();
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
  // If we can't check, fallback or fail. Here we try checking.
  try {
    const res = await fetch(`https://api.github.com/repos/${REPO}/releases/latest`, {
      headers: { "User-Agent": "aerostack-cli-npm/1.5.0" },
    });
    if (!res.ok) return "1.3.0"; // Fallback if rate limited
    const data = await res.json();
    return data.tag_name.replace(/^v/, '');
  } catch {
    return "1.3.0";
  }
}

async function downloadBinary(version, asset) {
  const ext = asset.startsWith("windows") ? "zip" : "tar.gz";
  const archive = `${BINARY}_v${version}_${asset}.${ext}`; // Check github release naming convention
  // Usually: aerostack_1.3.0_darwin_arm64.tar.gz
  // Let's assume standard GoReleaser naming: name_version_os_arch
  const releaseName = `${BINARY}_${version}_${asset}.${ext}`;

  const url = `https://github.com/${REPO}/releases/download/v${version}/${releaseName}`;
  const tmpDir = join(tmpdir(), `aerostack-${Date.now()}`);
  mkdirSync(tmpDir, { recursive: true });

  const archivePath = join(tmpDir, releaseName);

  console.error(`Downloading Aerostack Go Core v${version}...`);
  const res = await fetch(url, { redirect: "follow" });
  if (!res.ok) throw new Error(`Download failed: ${url} (${res.status})`);

  const buffer = Buffer.from(await res.arrayBuffer());
  writeFileSync(archivePath, buffer);

  const binDir = INSTALL_DIR;
  mkdirSync(binDir, { recursive: true });
  const binPath = join(binDir, process.platform === "win32" ? `${BINARY}.exe` : BINARY);

  if (ext === "zip") {
    const AdmZip = (await import("adm-zip")).default;
    const zip = new AdmZip(archivePath);
    zip.extractAllTo(tmpDir, true);
    const extracted = join(tmpDir, `${BINARY}.exe`);
    if (existsSync(extracted)) renameSync(extracted, binPath);
  } else {
    const tar = (await import("tar")).default;
    await tar.x({ file: archivePath, cwd: tmpDir });
    const extracted = join(tmpDir, BINARY);
    if (existsSync(extracted)) renameSync(extracted, binPath);
  }

  chmodSync(binPath, 0o755);
  try { rmSync(tmpDir, { recursive: true, force: true }); } catch { }
  return binPath;
}

async function ensureBinary() {
  const binPath = join(INSTALL_DIR, process.platform === "win32" ? `${BINARY}.exe` : BINARY);
  const { arch, asset } = getPlatform();

  // For now, always try to grab latest or fallback
  // In production we should cache version check
  if (existsSync(binPath)) return binPath;

  const version = await getLatestVersion();
  return await downloadBinary(version, asset);
}
