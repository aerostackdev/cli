#!/usr/bin/env node
/**
 * Clean up cached binary when npm uninstall runs.
 */
const fs = require("fs");
const path = require("path");

const INSTALL_DIR = process.env.AEROSTACK_INSTALL_DIR || path.join(process.env.HOME || process.env.USERPROFILE, ".aerostack", "bin");
const binPath = path.join(INSTALL_DIR, process.platform === "win32" ? "aerostack.exe" : "aerostack");

try {
  if (fs.existsSync(binPath)) {
    fs.unlinkSync(binPath);
  }
} catch (_) {}
