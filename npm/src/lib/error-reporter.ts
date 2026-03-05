/**
 * CLI Error Reporter + Zero-Fix™ AI Suggestions
 *
 * Reports unhandled CLI errors to the Aerostack telemetry backend,
 * then calls the AI analyze endpoint and renders suggestions inline.
 */

import chalk from 'chalk';
import { getConfig } from './config.js';

const VERSION = '1.0.0';

export interface AnalysisSuggestion {
    rootCause: string;
    suggestedFix: {
        description: string;
        code?: string;
        file?: string;
    };
    autoFixable: boolean;
    confidence: number;
}

/**
 * Reports a CLI error to the backend telemetry endpoint.
 * Never throws — silently no-ops if offline or unauthenticated.
 * Returns the error log ID (for downstream AI analysis) or null.
 */
export async function reportError(
    err: unknown,
    command: string,
    logs: string[] = []
): Promise<string | null> {
    const config = getConfig();
    if (!config.token || !config.registry) return null;

    const message = err instanceof Error ? err.message : String(err);
    const stack = err instanceof Error ? err.stack : undefined;

    try {
        const res = await fetch(`${config.registry}/v1/cli/telemetry/errors`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${config.token}`,
            },
            body: JSON.stringify({
                cli_version: VERSION,
                os: process.platform,
                command,
                error_message: message,
                error_stack: stack,
                logs,
            }),
            signal: AbortSignal.timeout(5000),
        });

        if (res.ok) {
            const data = await res.json() as { id?: string };
            return data.id ?? null;
        }
    } catch {
        // Network error or timeout — silently ignore
    }

    return null;
}

/**
 * Calls the AI analysis endpoint for a CLI error log.
 * Returns the analysis or null if unavailable.
 */
async function analyzeError(errorId: string): Promise<AnalysisSuggestion | null> {
    const config = getConfig();
    if (!config.token || !config.registry) return null;

    try {
        const res = await fetch(`${config.registry}/v1/cli/telemetry/errors/${errorId}/analyze`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${config.token}`,
            },
            signal: AbortSignal.timeout(30000), // AI can be slow
        });

        if (res.ok) {
            return await res.json() as AnalysisSuggestion;
        }
    } catch {
        // Offline or error — silently ignore
    }

    return null;
}

/**
 * Prints an AI-generated fix suggestion to the terminal.
 */
function renderSuggestion(suggestion: AnalysisSuggestion): void {
    const confidence = Math.round(suggestion.confidence * 100);
    const confidenceColor = confidence >= 75 ? chalk.green : confidence >= 50 ? chalk.yellow : chalk.gray;

    console.log('');
    console.log(
        chalk.bold('  🤖 Zero-Fix™ Analysis') +
        chalk.gray(' · ') +
        confidenceColor(`${confidence}% confidence`)
    );
    console.log(chalk.gray('  ' + '─'.repeat(50)));

    console.log('');
    console.log(chalk.bold('  Root cause:'));
    console.log(chalk.white('  ' + suggestion.rootCause));

    if (suggestion.suggestedFix?.description) {
        console.log('');
        console.log(chalk.bold('  Suggested fix:'));
        if (suggestion.suggestedFix.file) {
            console.log(chalk.gray(`  in ${chalk.cyan(suggestion.suggestedFix.file)}:`));
        }
        console.log(chalk.white('  ' + suggestion.suggestedFix.description));

        if (suggestion.suggestedFix.code) {
            console.log('');
            const lines = suggestion.suggestedFix.code.trim().split('\n');
            for (const line of lines.slice(0, 15)) {
                console.log(chalk.gray('  │ ') + chalk.green(line));
            }
            if (lines.length > 15) {
                console.log(chalk.gray(`  │ ... (${lines.length - 15} more lines)`));
            }
        }
    }

    if (suggestion.autoFixable) {
        console.log('');
        console.log(chalk.bgBlue.white.bold('  ✨ Auto-fixable  ') + chalk.gray(' This fix can be applied from the dashboard'));
    }

    console.log('');
}

/**
 * Main handler: reports error, fetches AI analysis, renders suggestion.
 * Call this in the catch block of any command.
 */
export async function handleCommandError(
    err: unknown,
    command: string,
    logs: string[] = []
): Promise<void> {
    const message = err instanceof Error ? err.message : String(err);

    // Print the raw error first
    console.error(chalk.red(`\n  ✗ Error in \`aerostack ${command}\``));
    console.error(chalk.red('  ' + message));
    if (err instanceof Error && err.stack) {
        const stackLines = err.stack.split('\n').slice(1, 4);
        for (const line of stackLines) {
            console.error(chalk.gray('  ' + line.trim()));
        }
    }

    // Report to backend (fire and forget for the ID)
    const config = getConfig();
    const isLoggedIn = !!config.token;

    if (!isLoggedIn) {
        console.error(chalk.gray('\n  Tip: Run `aerostack login` to get AI-powered error analysis.\n'));
        return;
    }

    // Show a brief "analyzing" indicator
    process.stdout.write(chalk.gray('\n  Analyzing with Zero-Fix™...'));

    const errorId = await reportError(err, command, logs);

    if (!errorId) {
        process.stdout.write('\r' + ' '.repeat(40) + '\r'); // Clear the "analyzing" line
        console.error(chalk.gray('\n  Could not reach analysis service — check your connection.\n'));
        return;
    }

    const suggestion = await analyzeError(errorId);
    process.stdout.write('\r' + ' '.repeat(40) + '\r'); // Clear the "analyzing" line

    if (suggestion) {
        renderSuggestion(suggestion);
    } else {
        console.error(chalk.gray('\n  AI analysis unavailable — error has been logged.\n'));
    }
}
