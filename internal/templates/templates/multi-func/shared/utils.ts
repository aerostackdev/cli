export const getGreeting = (name: string) => {
    return `Hello, ${name}! Welcome to the world of multi-function serverless apps.`;
};

export const formatUser = (user: any) => {
    return {
        ...user,
        fullName: `${user.firstName} ${user.lastName}`,
        processedAt: new Date().toISOString()
    };
};
