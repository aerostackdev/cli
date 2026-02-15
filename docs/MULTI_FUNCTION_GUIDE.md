# Multi-Function Architecture in Aerostack

Aerostack supports complex architectures where a single project can contain multiple functions (Workers) that share code and interact with each other.

## The `shared` Folder

By convention, you should put common logic, types, and utilities in a `shared/` folder at the root of your project. This folder is accessible to all functions in your project.

### Example Structure

```
my-project/
├── aerostack.toml
├── shared/
│   └── utils.ts
├── api/
│   └── index.ts
└── functions/
    └── auth-hook.ts
```

## Configuring `aerostack.toml`

Define your functions in the `aerostack.toml` file:

```toml
[main]
path = "api/index.ts"

[[functions]]
name = "auth-hook"
path = "functions/auth-hook.ts"
route = "/auth/*"
```

## Cross-Function Calls (RPC)

Aerostack automatically sets up Service Bindings for your functions. You can call other functions directly from your code:

```typescript
// In api/index.ts
app.get('/profile', async (c) => {
  const auth = c.env.AUTH_HOOK;
  const user = await auth.checkSession(c.req.header('Authorization'));
  return c.json(user);
});
```

## Benefits

1. **Reduced Cold Starts**: Small, focused functions start faster.
2. **Independent Scaling**: Scale critical paths independently.
3. **Better Organization**: Separate concerns (auth, api, background jobs) clearly.
4. **Shared Types**: Maintain type safety across your entire stack.
