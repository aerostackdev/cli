/**
 * AST-safe injector for src/index.ts and src/db/schema.ts
 * 
 * Uses marker comments so injections are idempotent and predictable.
 * Markers set during `aerostack init` — never modifies files without them.
 */

import { readFileSync, writeFileSync, existsSync } from 'fs';

// Marker constants — must match templates/src/index.ts
const IMPORTS_MARKER = '// aerostack:imports';
const ROUTES_MARKER = '// aerostack:routes';
const SCHEMA_IMPORTS_MARKER = '// aerostack:schema-imports';
const SCHEMA_EXPORTS_MARKER = '// aerostack:schema-exports';

export interface InjectionResult {
    modified: boolean;
    reason?: string;
}

/**
 * Injects an import + route into src/index.ts.
 * Returns false if already injected (idempotent).
 */
export function injectRoute(
    indexPath: string,
    opts: {
        importStatement: string;   // e.g. import { myRoute } from './modules/my-fn/adapter';
        routeStatement: string;    // e.g. app.route('/api/my-fn', myRoute);
    }
): InjectionResult {
    if (!existsSync(indexPath)) {
        return { modified: false, reason: `File not found: ${indexPath}` };
    }

    let content = readFileSync(indexPath, 'utf-8');

    // Check if already injected (idempotent)
    if (content.includes(opts.importStatement)) {
        return { modified: false, reason: 'Already injected' };
    }

    // Inject import before the marker
    if (content.includes(IMPORTS_MARKER)) {
        content = content.replace(
            IMPORTS_MARKER,
            `${opts.importStatement}\n${IMPORTS_MARKER}`
        );
    } else {
        // Fallback: add near top after last existing import
        const lines = content.split('\n');
        let lastImportIdx = -1;
        for (let i = 0; i < lines.length; i++) {
            if (lines[i].startsWith('import ')) lastImportIdx = i;
        }
        if (lastImportIdx >= 0) {
            lines.splice(lastImportIdx + 1, 0, opts.importStatement);
            content = lines.join('\n');
        }
    }

    // Inject route before the marker
    if (content.includes(ROUTES_MARKER)) {
        content = content.replace(
            ROUTES_MARKER,
            `${opts.routeStatement}\n${ROUTES_MARKER}`
        );
    } else {
        // Fallback: add before export default
        content = content.replace(
            'export default app',
            `${opts.routeStatement}\n\nexport default app`
        );
    }

    writeFileSync(indexPath, content, 'utf-8');
    return { modified: true };
}

/**
 * Injects schema import + export re-export into src/db/schema.ts
 */
export function injectSchema(
    schemaPath: string,
    opts: {
        importStatement: string;   // e.g. import * as myFnSchema from '../modules/my-fn/schema';
        exportStatement: string;   // e.g. export * from '../modules/my-fn/schema';
    }
): InjectionResult {
    if (!existsSync(schemaPath)) {
        // Non-fatal: db/schema.ts might not exist (users can run drizzle-kit directly)
        return { modified: false, reason: 'db/schema.ts not found — skipping schema injection' };
    }

    let content = readFileSync(schemaPath, 'utf-8');

    if (content.includes(opts.exportStatement)) {
        return { modified: false, reason: 'Already injected' };
    }

    if (content.includes(SCHEMA_IMPORTS_MARKER)) {
        content = content.replace(
            SCHEMA_IMPORTS_MARKER,
            `${opts.importStatement}\n${SCHEMA_IMPORTS_MARKER}`
        );
    }

    if (content.includes(SCHEMA_EXPORTS_MARKER)) {
        content = content.replace(
            SCHEMA_EXPORTS_MARKER,
            `${opts.exportStatement}\n${SCHEMA_EXPORTS_MARKER}`
        );
    } else {
        content += `\n${opts.exportStatement}\n`;
    }

    writeFileSync(schemaPath, content, 'utf-8');
    return { modified: true };
}

/**
 * Replace template tokens like {{PROJECT_NAME}} in a file string.
 */
export function applyTokens(content: string, tokens: Record<string, string>): string {
    return content.replace(/\{\{(\w+)\}\}/g, (_, key) => tokens[key] ?? `{{${key}}}`);
}
