// shared/utils.ts
export function formatGreeting(name: string): string {
    return `Hello, ${name}! This is a shared greeting.`;
}

export function getProjectInfo() {
    return { name: "Monolith Test", version: "1.0.0" };
}
