#!/usr/bin/env node
import { existsSync } from 'fs';
import { join } from 'path';

const binDir = join(process.env.HOME || process.env.USERPROFILE || '.', '.aerostack', 'bin');
if (existsSync(binDir)) {
  console.log('\n  Aerostack CLI is ready.');
  console.log('  Run `npx aerostack --help` or add the CLI to your PATH:');
  console.log(`    ${binDir}`);
  console.log('  Docs: https://aerostack.dev/docs/cli\n');
}
