/**
 * `aerostack init [directory]`
 * 
 * Scaffolds a new Aerostack project with:
 * - Hono framework (Cloudflare Workers)
 * - Drizzle ORM (D1 Database)
 * - Standard module structure (src/modules/, src/lib/, src/db/)
 * - Injection markers so `aerostack add` can auto-wire modules
 */

import chalk from 'chalk';
import prompts from 'prompts';
import ora from 'ora';
import { existsSync, mkdirSync, readFileSync, writeFileSync, readdirSync, statSync, cpSync } from 'fs';
import { join, resolve, dirname } from 'path';
import { fileURLToPath } from 'url';
import { execSync } from 'child_process';
import { applyTokens } from '../lib/injector.js';

const __dirname = dirname(fileURLToPath(import.meta.url));
const TEMPLATES_DIR = resolve(__dirname, '../../templates');

function copyDir(src: string, dest: string, tokens: Record<string, string>) {
    mkdirSync(dest, { recursive: true });
    for (const entry of readdirSync(src)) {
        const srcPath = join(src, entry);
        const destPath = join(dest, entry);
        if (statSync(srcPath).isDirectory()) {
            copyDir(srcPath, destPath, tokens);
        } else {
            const raw = readFileSync(srcPath, 'utf-8');
            writeFileSync(destPath, applyTokens(raw, tokens), 'utf-8');
        }
    }
}

export async function initCommand(args: string[]) {
    console.log(`\n${chalk.bold.blue('  ⚡ Aerostack Init')}\n`);

    let targetDir = args[0] || '.';

    // Prompt for project name
    const { projectName } = await prompts({
        type: 'text',
        name: 'projectName',
        message: 'Project name:',
        initial: targetDir === '.' ? 'my-aerostack-project' : targetDir,
        validate: (v: string) => v.trim().length > 0 || 'Name is required',
    });

    if (!projectName) {
        console.log(chalk.yellow('\n  Aborted.\n'));
        process.exit(0);
    }

    const { confirmInstall } = await prompts({
        type: 'confirm',
        name: 'confirmInstall',
        message: 'Install npm dependencies now?',
        initial: true,
    });

    const resolvedDir = resolve(process.cwd(), targetDir === '.' ? '.' : projectName);

    if (existsSync(resolvedDir) && targetDir !== '.') {
        const { overwrite } = await prompts({
            type: 'confirm',
            name: 'overwrite',
            message: `Directory "${projectName}" already exists. Continue?`,
            initial: false,
        });
        if (!overwrite) {
            console.log(chalk.yellow('\n  Aborted.\n'));
            process.exit(0);
        }
    }

    const spinner = ora('Creating project files...').start();

    try {
        const tokens = {
            PROJECT_NAME: projectName,
        };

        // Copy all template files with token substitution
        copyDir(TEMPLATES_DIR, resolvedDir, tokens);

        // Create empty module placeholder
        mkdirSync(join(resolvedDir, 'src', 'modules'), { recursive: true });
        mkdirSync(join(resolvedDir, 'src', 'lib'), { recursive: true });
        mkdirSync(join(resolvedDir, 'drizzle'), { recursive: true });

        // Create .env.example
        writeFileSync(
            join(resolvedDir, '.env.example'),
            `CLOUDFLARE_ACCOUNT_ID=\nCLOUDFLARE_DATABASE_ID=\nCLOUDFLARE_D1_TOKEN=\n`
        );

        // Create .gitignore
        writeFileSync(
            join(resolvedDir, '.gitignore'),
            `node_modules/\ndist/\n.wrangler/\n.aerostack/\n.env\n`
        );

        spinner.succeed('Project files created!');

        if (confirmInstall) {
            const installSpinner = ora('Installing dependencies...').start();
            try {
                execSync('npm install', { cwd: resolvedDir, stdio: 'pipe' });
                installSpinner.succeed('Dependencies installed!');
            } catch {
                installSpinner.warn('npm install failed — run it manually');
            }
        }

        console.log(`
${chalk.green('  ✓ Project created successfully!')}

${chalk.bold('  Next steps:')}

  ${chalk.gray('1.')} ${chalk.cyan(`cd ${targetDir === '.' ? '.' : projectName}`)}
  ${chalk.gray('2.')} Set up Cloudflare D1: ${chalk.cyan('npx wrangler d1 create my-db')}
  ${chalk.gray('3.')} Update ${chalk.bold('wrangler.toml')} with your database ID
  ${chalk.gray('4.')} Push schema: ${chalk.cyan('npx drizzle-kit push')}
  ${chalk.gray('5.')} Start dev: ${chalk.cyan('npx wrangler dev')}
  ${chalk.gray('6.')} Add functions: ${chalk.cyan('npx aerostack add <function-name>')}

${chalk.bold('  Registry:')}  ${chalk.underline('https://hub.aerostack.dev')}
`);
    } catch (err: any) {
        spinner.fail('Failed to create project');
        throw err;
    }
}
