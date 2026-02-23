/**
 * `aerostack publish [module-path]`
 * 
 * Reads a local function module and pushes it to the community registry.
 * 
 * Usage:
 *   npx aerostack publish                       → uses current directory
 *   npx aerostack publish ./src/modules/my-fn   → specific module
 */

import chalk from 'chalk';
import prompts from 'prompts';
import ora from 'ora';
import { readFileSync, existsSync, readdirSync, statSync } from 'fs';
import { join, resolve, basename } from 'path';
import { createFunction, updateFunction, publishFunction, DEFAULT_REGISTRY } from '../lib/registry.js';
import { isAuthenticated } from '../lib/config.js';

interface LocalManifest {
    id?: string;
    name?: string;
    slug?: string;
    author?: string;
    version?: string;
    type?: string;
    routeExport?: string;
    routePath?: string;
    drizzleSchema?: boolean;
    npmDependencies?: string[];
    envVars?: string[];
}

function readLocalManifest(moduleDir: string): LocalManifest {
    const manifestPath = join(moduleDir, 'aerostack-manifest.json');
    if (existsSync(manifestPath)) {
        return JSON.parse(readFileSync(manifestPath, 'utf-8'));
    }
    return {};
}

function readAerostackJson(moduleDir: string): any {
    const configPath = join(moduleDir, 'aerostack.json');
    if (existsSync(configPath)) {
        try {
            return JSON.parse(readFileSync(configPath, 'utf-8'));
        } catch (e) {
            console.warn(chalk.yellow(`  Warning: Could not parse aerostack.json in ${moduleDir}`));
        }
    }
    return {};
}

function readCoreFile(moduleDir: string): string {
    // Prefer core.ts, then adapter.ts, then index.ts
    for (const fname of ['core.ts', 'adapter.ts', 'index.ts']) {
        const p = join(moduleDir, fname);
        if (existsSync(p)) return readFileSync(p, 'utf-8');
    }
    return '';
}

function inferCategory(moduleDir: string): string {
    const files = readdirSync(moduleDir);
    const code = files
        .filter(f => f.endsWith('.ts'))
        .map(f => readFileSync(join(moduleDir, f), 'utf-8'))
        .join('\n');
    if (code.includes('stripe')) return 'payments';
    if (code.includes('drizzle') && code.includes('auth')) return 'auth';
    if (code.includes('email') || code.includes('sendgrid')) return 'email';
    return 'utility';
}

function parseArgs(args: string[]) {
    const path = args.find(a => !a.startsWith('--'));
    const flags: Record<string, string> = {};
    for (const arg of args) {
        if (arg.startsWith('--')) {
            const [key, val] = arg.slice(2).split('=');
            flags[key] = val || 'true';
        }
    }
    return { path, flags };
}

export async function publishCommand(args: string[]) {
    const { path: modulePath, flags } = parseArgs(args);
    const registry = flags['registry'] || DEFAULT_REGISTRY;

    console.log(`\n${chalk.bold.blue('  ⚡ Aerostack Publish')}\n`);

    if (!isAuthenticated()) {
        console.error(chalk.red('  Not logged in. Run: npx aerostack login\n'));
        process.exit(1);
    }

    // Find module directory
    let moduleDir: string;
    if (modulePath) {
        moduleDir = resolve(process.cwd(), modulePath);
    } else {
        // Try current directory or look for a single module
        const srcModules = join(process.cwd(), 'src', 'modules');
        if (existsSync(srcModules)) {
            const dirs = readdirSync(srcModules).filter(d =>
                statSync(join(srcModules, d)).isDirectory()
            );
            if (dirs.length === 1) {
                moduleDir = join(srcModules, dirs[0]);
            } else if (dirs.length > 1) {
                const { selected } = await prompts({
                    type: 'select',
                    name: 'selected',
                    message: 'Which module to publish?',
                    choices: dirs.map(d => ({ title: d, value: join(srcModules, d) })),
                });
                moduleDir = selected;
            } else {
                console.error(chalk.red('  No modules found in src/modules/\n'));
                process.exit(1);
            }
        } else {
            moduleDir = process.cwd();
        }
    }

    if (!existsSync(moduleDir)) {
        console.error(chalk.red(`  Directory not found: ${moduleDir}\n`));
        process.exit(1);
    }

    const localManifest = readLocalManifest(moduleDir);
    const moduleName = localManifest.name || basename(moduleDir);

    // Read the Gateway configurations (Phase 1.7)
    const aerostackConfig = readAerostackJson(moduleDir);
    const aiConfig = aerostackConfig.ai_config || aerostackConfig.aiConfig;
    const monetization = aerostackConfig.monetization;

    // Collect all TS file content
    const tsFiles = readdirSync(moduleDir)
        .filter(f => f.endsWith('.ts') && f !== 'node-adapter.ts')
        .map(f => ({ path: f, content: readFileSync(join(moduleDir, f), 'utf-8') }));

    const primaryCode = readCoreFile(moduleDir);

    if (!primaryCode) {
        console.error(chalk.red('  No TypeScript source files found.\n'));
        process.exit(1);
    }

    // Prompt for function metadata
    const answers = await prompts([
        {
            type: 'text',
            name: 'name',
            message: 'Function name:',
            initial: moduleName,
        },
        {
            type: 'text',
            name: 'description',
            message: 'Description:',
            initial: '',
            validate: (v: string) => v.trim().length >= 10 || 'At least 10 characters required',
        },
        {
            type: 'select',
            name: 'category',
            message: 'Category:',
            choices: [
                { title: 'Payments', value: 'payments' },
                { title: 'Auth', value: 'auth' },
                { title: 'Email', value: 'email' },
                { title: 'Data Processing', value: 'data' },
                { title: 'API Integration', value: 'api' },
                { title: 'AI / ML', value: 'ai' },
                { title: 'Web3', value: 'web3' },
                { title: 'Utility', value: 'utility' },
                { title: 'Other', value: 'other' },
            ],
            initial: 0,
        },
        {
            type: 'text',
            name: 'tags',
            message: 'Tags (comma separated):',
            initial: '',
        },
        {
            type: 'confirm',
            name: 'publishNow',
            message: 'Publish immediately? (No = save as draft)',
            initial: false,
        },
    ]);

    if (!answers.name) {
        console.log(chalk.yellow('\n  Aborted.\n'));
        process.exit(0);
    }

    const tags = answers.tags ? answers.tags.split(',').map((t: string) => t.trim()).filter(Boolean) : [];

    // Upload
    const spinner = ora('Uploading to registry...').start();
    try {
        let fnId: string;
        let fnSlug: string;

        if (localManifest.id) {
            // Update existing
            await updateFunction(registry, localManifest.id, {
                name: answers.name,
                description: answers.description,
                code: primaryCode,
                tags,
                aiConfig,
                monetization,
            });
            fnId = localManifest.id;
            fnSlug = localManifest.slug || answers.name;
            spinner.succeed('Function updated');
        } else {
            // Create new
            const created = await createFunction(registry, {
                name: answers.name,
                description: answers.description,
                category: answers.category,
                code: primaryCode,
                tags,
                language: 'typescript',
                license: 'MIT',
                aiConfig,
                monetization,
            });
            fnId = created.id;
            fnSlug = created.slug;
            spinner.succeed(`Function created: ${chalk.cyan(`${created.author}/${fnSlug}`)}`);
        }

        if (answers.publishNow) {
            const pubSpinner = ora('Publishing...').start();
            const result = await publishFunction(registry, fnId);
            pubSpinner.succeed(`Published! 🎉`);
            console.log(`\n  ${chalk.bold('Registry URL:')} ${chalk.underline(`https://hub.aerostack.dev${result.url}`)}\n`);
        } else {
            console.log(`\n  ${chalk.green('✓ Saved as draft.')} To publish later:`);
            console.log(`  ${chalk.cyan('npx aerostack publish --publish-only')}\n`);
        }
    } catch (err: any) {
        spinner.fail(err.message);
        process.exit(1);
    }
}
