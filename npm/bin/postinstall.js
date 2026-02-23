#!/usr/bin/env node
import { existsSync } from 'fs';
import { join } from 'path';

// Just a polite notice
const binDir = join(process.env.HOME || process.env.USERPROFILE || '.', '.aerostack', 'bin');
if (existsSync(binDir)) {
  console.log(`\nAerostack is installed.`);
  console.log(`To use the binary directly, add ${binDir} to your PATH.\n`);
}
