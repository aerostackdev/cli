/**
 * `aerostack list`
 * 
 * Browse the community function registry in your terminal.
 * 
 * Usage:
 *   npx aerostack list
 *   npx aerostack list --category=payments
 *   npx aerostack list --search=stripe
 *   npx aerostack list --sort=stars
 */

import chalk from 'chalk';
import ora from 'ora';
import { listFunctions, DEFAULT_REGISTRY, type FunctionListItem } from '../lib/registry.js';

function parseArgs(args: string[]) {
    const flags: Record<string, string> = {};
    for (const arg of args) {
        if (arg.startsWith('--')) {
            const [key, val] = arg.slice(2).split('=');
            flags[key] = val || 'true';
        }
    }
    return flags;
}

function truncate(str: string, max: number): string {
    return str.length > max ? str.slice(0, max - 1) + '…' : str;
}

function renderFunction(fn: FunctionListItem, i: number): string {
    const stars = chalk.yellow(`★ ${fn.star_count}`);
    const clones = chalk.gray(`⬇ ${fn.clone_count}`);
    const category = chalk.blue(`[${fn.category}]`);
    const name = chalk.bold.white(fn.name);
    const author = chalk.gray(`@${fn.author_username}`);
    const desc = chalk.gray(truncate(fn.description || '', 70));
    const install = chalk.cyan(`npx aerostack add ${fn.author_username}/${fn.slug}`);

    return [
        `  ${chalk.gray(String(i + 1).padStart(2, ' '))}. ${name} ${author}`,
        `      ${category} ${stars}  ${clones}`,
        `      ${desc}`,
        `      ${install}`,
    ].join('\n');
}

export async function listCommand(args: string[]) {
    const flags = parseArgs(args);
    const registry = flags['registry'] || DEFAULT_REGISTRY;

    const opts = {
        category: flags['category'],
        search: flags['search'],
        sort: flags['sort'] || 'stars',
        limit: flags['limit'] ? parseInt(flags['limit']) : 20,
    };

    console.log(`\n${chalk.bold.blue('  ⚡ Aerostack Registry')}`);
    if (opts.search) console.log(`  ${chalk.gray(`Search: "${opts.search}"`)}`);
    if (opts.category) console.log(`  ${chalk.gray(`Category: ${opts.category}`)}`);
    console.log();

    const spinner = ora('Fetching functions...').start();

    let functions: FunctionListItem[];
    try {
        functions = await listFunctions(registry, opts);
        spinner.stop();
    } catch (err: any) {
        spinner.fail(err.message);
        process.exit(1);
    }

    if (functions.length === 0) {
        console.log(chalk.yellow('  No functions found.\n'));
        return;
    }

    console.log(`  ${chalk.bold(`${functions.length} functions`)}\n`);

    for (let i = 0; i < functions.length; i++) {
        console.log(renderFunction(functions[i], i));
        console.log();
    }

    console.log(chalk.gray(`  Browse more: ${chalk.underline('https://hub.aerostack.dev/functions')}\n`));
}
