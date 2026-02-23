/**
 * Registry API client
 * Communicates with the Aerostack community API to fetch and publish functions.
 */

import { getConfig } from './config.js';

export const DEFAULT_REGISTRY = 'https://api.aerostack.dev/api';

export interface FunctionManifest {
    id: string;
    name: string;
    slug: string;
    description: string;
    category: string;
    type: 'feature' | 'utility';
    author: string;
    version: string;
    language: string;
    starCount: number;
    cloneCount: number;
    license: string;
    tags: string[];
    npmDependencies?: string[];
    envVars?: string[];
    routeExport?: string;
    routePath?: string;
    drizzleSchema?: boolean;
    aiConfig?: any;
    monetization?: any;
    files: Array<{
        path: string;
        content: string;
    }>;
}

export interface FunctionListItem {
    id: string;
    name: string;
    slug: string;
    description: string;
    category: string;
    author_username: string;
    star_count: number;
    clone_count: number;
    version: string;
    tags: string[];
}

function getRegistryUrl(): string {
    try {
        const config = getConfig();
        return config.registry || DEFAULT_REGISTRY;
    } catch {
        return DEFAULT_REGISTRY;
    }
}

function getAuthToken(): string | null {
    try {
        const config = getConfig();
        return config.token || null;
    } catch {
        return null;
    }
}

function buildHeaders(requireAuth = false): Record<string, string> {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        'User-Agent': 'aerostack-cli/1.0.0',
    };
    const token = getAuthToken();
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    } else if (requireAuth) {
        throw new Error('Not authenticated. Run: npx aerostack login');
    }
    return headers;
}

export async function fetchInstallManifest(registry: string, slug: string): Promise<FunctionManifest> {
    // Support "author/slug" or just "slug"
    let url: string;
    if (slug.includes('/')) {
        const [author, fnSlug] = slug.split('/');
        url = `${registry}/community/functions/${author}/${fnSlug}/install`;
    } else {
        // Search by slug directly
        url = `${registry}/community/functions/install/${slug}`;
    }

    const res = await fetch(url, { headers: buildHeaders() });

    if (res.status === 404) {
        throw new Error(`Function "${slug}" not found in registry. Run: npx aerostack list`);
    }
    if (!res.ok) {
        const body = await res.json().catch(() => ({})) as any;
        throw new Error(body?.error || `Registry error: ${res.status}`);
    }

    return res.json() as Promise<FunctionManifest>;
}

export async function listFunctions(registry: string, opts: {
    category?: string;
    search?: string;
    sort?: string;
    limit?: number;
} = {}): Promise<FunctionListItem[]> {
    const params = new URLSearchParams();
    if (opts.category) params.set('category', opts.category);
    if (opts.search) params.set('search', opts.search);
    if (opts.sort) params.set('sort', opts.sort);
    if (opts.limit) params.set('limit', String(opts.limit));

    const url = `${registry}/community/functions?${params}`;
    const res = await fetch(url, { headers: buildHeaders() });

    if (!res.ok) throw new Error(`Registry error: ${res.status}`);
    const data = await res.json() as any;
    return data.functions || [];
}

export async function createFunction(registry: string, payload: {
    name: string;
    description: string;
    category: string;
    code: string;
    language?: string;
    tags?: string[];
    readme?: string;
    license?: string;
    aiConfig?: any;
    monetization?: any;
}): Promise<{ id: string; slug: string; author: string }> {
    const res = await fetch(`${registry}/community/functions`, {
        method: 'POST',
        headers: buildHeaders(true),
        body: JSON.stringify(payload),
    });

    if (!res.ok) {
        const body = await res.json().catch(() => ({})) as any;
        throw new Error(body?.error || `Failed to create function: ${res.status}`);
    }
    return res.json() as any;
}

export async function updateFunction(registry: string, id: string, payload: Partial<{
    name: string;
    description: string;
    code: string;
    tags: string[];
    readme: string;
    version: string;
    aiConfig?: any;
    monetization?: any;
}>): Promise<void> {
    const res = await fetch(`${registry}/community/functions/${id}`, {
        method: 'PATCH',
        headers: buildHeaders(true),
        body: JSON.stringify(payload),
    });
    if (!res.ok) {
        const body = await res.json().catch(() => ({})) as any;
        throw new Error(body?.error || `Failed to update function: ${res.status}`);
    }
}

export async function publishFunction(registry: string, id: string): Promise<{ url: string }> {
    const res = await fetch(`${registry}/community/functions/${id}/publish`, {
        method: 'POST',
        headers: buildHeaders(true),
    });
    if (!res.ok) {
        const body = await res.json().catch(() => ({})) as any;
        throw new Error(body?.error || `Failed to publish function: ${res.status}`);
    }
    return res.json() as any;
}

export async function loginToRegistry(registry: string, email: string, password: string): Promise<string> {
    const res = await fetch(`${registry}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
    });
    if (!res.ok) {
        const body = await res.json().catch(() => ({})) as any;
        throw new Error(body?.error || 'Login failed');
    }
    const data = await res.json() as any;
    return data.token;
}
