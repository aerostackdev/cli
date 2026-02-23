/**
 * `aerostack login`
 * 
 * Authenticates with the Aerostack registry and stores the JWT
 * in ~/.aerostack/config.json.
 * 
 * Usage:
 *   npx aerostack login
 *   npx aerostack login --registry=https://api.mystack.dev/api
 */

import chalk from 'chalk';
import prompts from 'prompts';
import ora from 'ora';
import { loginToRegistry, DEFAULT_REGISTRY } from '../lib/registry.js';
import { saveConfig, getConfig } from '../lib/config.js';

export async function loginCommand(args: string[]) {
    const flags: Record<string, string> = {};
    for (const arg of args) {
        if (arg.startsWith('--')) {
            const [key, val] = arg.slice(2).split('=');
            flags[key] = val || 'true';
        }
    }

    const registry = flags['registry'] || DEFAULT_REGISTRY;

    console.log(`\n${chalk.bold.blue('  ⚡ Aerostack Login')}\n`);
    console.log(`  Connecting to ${chalk.cyan(registry)}\n`);

    const { email, password } = await prompts([
        {
            type: 'text',
            name: 'email',
            message: 'Email:',
            validate: (v: string) => v.includes('@') || 'Enter a valid email',
        },
        {
            type: 'password',
            name: 'password',
            message: 'Password:',
            validate: (v: string) => v.length > 0 || 'Password required',
        },
    ]);

    if (!email || !password) {
        console.log(chalk.yellow('\n  Aborted.\n'));
        process.exit(0);
    }

    const spinner = ora('Logging in...').start();
    try {
        const token = await loginToRegistry(registry, email, password);
        const existingConfig = getConfig();
        saveConfig({ ...existingConfig, token, email, registry });
        spinner.succeed(`Logged in as ${chalk.cyan(email)}`);
        console.log(chalk.gray(`\n  Token saved to ~/.aerostack/config.json\n`));
        console.log(`  You can now publish functions: ${chalk.cyan('npx aerostack publish')}\n`);
    } catch (err: any) {
        spinner.fail(err.message);
        process.exit(1);
    }
}
