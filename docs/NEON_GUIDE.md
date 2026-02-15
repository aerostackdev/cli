# Using Neon PostgreSQL with Aerostack

Aerostack provides native support for Neon PostgreSQL, allowing you to build serverless applications with a fully managed, scalable Postgres database at the edge.

## Setup

1. **Create a Neon Project**:
   If you haven't already, create a new project on [Neon](https://console.neon.tech).

2. **Get your API Key**:
   Go to your Neon Settings and create a new API Key.

3. **Configure the CLI**:
   Export your API key as an environment variable:
   ```bash
   export NEON_API_KEY=your-api-key
   ```

4. **Initialize a Neon Project**:
   ```bash
   aerostack init my-app --db=neon
   ```

5. **Create the Database via CLI**:
   Inside your project directory, run:
   ```bash
   aerostack db:neon create my-db --add-to-config
   ```

## Usage in Code

Your database will be available on the `env` object:

```typescript
app.get('/data', async (c) => {
  const db = c.env.PG;
  const result = await db.query('SELECT * FROM my_table');
  return c.json(result.rows);
});
```

## Benefits of Neon

- **Serverless**: Database scales to zero when not in use.
- **Branching**: Create instant copies of your database for testing.
- **Edge-Ready**: Optimized for connectivity from serverless functions.
