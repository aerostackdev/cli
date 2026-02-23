#!/usr/bin/env node
/**
 * Aerostack CLI — Main Entry Point
 * 
 * Usage:
 *   npx aerostack init [directory]   — scaffold a new project
 *   npx aerostack add <slug>          — install a function from registry
 *   npx aerostack list               — browse available functions
 *   npx aerostack publish            — push local function to registry
 *   npx aerostack login              — authenticate with Aerostack
 */

import chalk from 'chalk';
import { initCommand } from './commands/init.js';
import { addCommand } from './commands/add.js';
import { listCommand } from './commands/list.js';
import { publishCommand } from './commands/publish.js';
import { loginCommand } from './commands/login.js';

// args parsed inside run()

const VERSION = '1.0.0';

function printHelp() {
    console.log(`
${chalk.bold.blue('  ⚡ Aerostack CLI')} ${chalk.gray(`v${VERSION}`)}

${chalk.bold('Usage:')}
  ${chalk.cyan('npx aerostack')} <command> [options]

${chalk.bold('Commands:')}
  ${chalk.cyan('init')} [directory]        Scaffold a new Aerostack project (Drizzle + Hono)
  ${chalk.cyan('add')} <slug>              Install a community function from the registry
  ${chalk.cyan('list')}                   Browse available functions in the registry
  ${chalk.cyan('publish')}                Publish a local function to the community registry
  ${chalk.cyan('login')}                  Authenticate with your Aerostack account

${chalk.bold('Examples:')}
  ${chalk.gray('# Create a new project')}
  npx aerostack init my-backend

  ${chalk.gray('# Install a community function')}
  npx aerostack add stripe-checkout
  npx aerostack add alice/stripe-checkout       ${chalk.gray('# specific author')}
  npx aerostack add stripe-checkout --runtime=node  ${chalk.gray('# with Node.js adapter')}

  ${chalk.gray('# Browse the registry')}
  npx aerostack list --category=payments

  ${chalk.gray('# Publish your function')}
  npx aerostack publish ./src/modules/my-fn

${chalk.bold('Documentation:')}  ${chalk.underline('https://aerostack.dev/docs/cli')}
`);
}

export async function run(args: string[]) {
    const command = args[0];
    const rest = args.slice(1);

    if (!command || command === '--help' || command === '-h') {
        printHelp();
        // If it's the Node.js handler, we exit. If falling through to Go, we wouldn't be here.
        process.exit(0);
    }

    if (command === '--version' || command === '-v') {
        console.log(VERSION);
        process.exit(0);
    }

    try {
        switch (command) {
            case 'init':
                await initCommand(rest);
                break;
            case 'add':
                await addCommand(rest);
                break;
            case 'list':
                await listCommand(rest);
                break;
            case 'publish':
                await publishCommand(rest);
                break;
            case 'login':
                await loginCommand(rest);
                break;
            default:
                // Should not happen if run.js filters correctly
                console.error(chalk.red(`\n  Unknown command: ${command}\n`));
                printHelp();
                process.exit(1);
        }
    } catch (err: any) {
        console.error(chalk.red(`\n  Error: ${err.message}\n`));
        process.exit(1);
    }
}
