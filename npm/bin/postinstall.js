#!/usr/bin/env node
/**
 * Postinstall: print PATH fix if aerostack bin not in PATH.
 */
const path = require("path");
const fs = require("fs");

// Assuming there's a download function that might be added or already exists
// and the user wants to wrap its execution.
// For the purpose of this edit, we'll assume `downloadBinary()` is a placeholder
// for the actual download logic that would be placed here.
try {
  // Original download logic (placeholder, as it's not in the provided original code)
  // If downloadBinary() is not defined, this will cause a ReferenceError.
  // The user's instruction implies this function exists or will be added.
  // downloadBinary(); // Uncomment and define if actual download logic is intended here
} catch (e) {
  console.warn('⚠️  Aerostack CLI download failed. You may need to install it manually or check your connection.');
  console.warn('   Details: ' + e.message);
  // Do not exit with error, allow install to continue
  process.exit(0);
}

// Global install: package at prefix/node_modules/aerostack, bin at prefix/bin
const pkgRoot = path.resolve(__dirname, "..");
const prefix = path.dirname(path.dirname(pkgRoot));
const binDir = path.join(prefix, "bin");
const binPath = path.join(binDir, process.platform === "win32" ? "aerostack.cmd" : "aerostack");

if (!fs.existsSync(binDir)) return;

const pathEnv = (process.env.PATH || "").split(path.delimiter);
const inPath = pathEnv.some((p) => path.resolve(p) === path.resolve(binDir));

if (!inPath) {
  console.log("\n\u001b[33mAerostack installed. Add to PATH (run once, or add to ~/.zshrc):\u001b[0m");
  console.log(`  \u001b[36mexport PATH="$PATH:${binDir}"\u001b[0m\n`);
  console.log("Or use \u001b[32mnpx aerostack\u001b[0m (no PATH config needed).\n");
} else {
  console.log("\n\u001b[32mAerostack installed. Run: aerostack --version\u001b[0m\n");
}
