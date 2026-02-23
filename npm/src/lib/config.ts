import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'fs';
import { homedir } from 'os';
import { join } from 'path';

const CONFIG_DIR = join(homedir(), '.aerostack');
const CONFIG_FILE = join(CONFIG_DIR, 'config.json');

export interface AerostackConfig {
    token?: string;
    email?: string;
    registry?: string;
}

export function getConfig(): AerostackConfig {
    if (!existsSync(CONFIG_FILE)) return {};
    try {
        return JSON.parse(readFileSync(CONFIG_FILE, 'utf-8'));
    } catch {
        return {};
    }
}

export function saveConfig(config: AerostackConfig): void {
    if (!existsSync(CONFIG_DIR)) mkdirSync(CONFIG_DIR, { recursive: true });
    writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2));
}

export function isAuthenticated(): boolean {
    return !!getConfig().token;
}
