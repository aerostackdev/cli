import express from "express";
import { httpServerHandler } from "cloudflare:node";
import { sdk } from "@aerostack/sdk";

const app = express();
app.use(express.json());

// Workers environment middleware
app.use((req, res, next) => {
  // @ts-ignore - Workers environment is attached to req in cloudflare:node
  const env = (req as any).env;
  if (env) sdk.init(env);
  next();
});

app.get("/", (req, res) => {
  res.send("Hello from Express on Aerostack! Try /test/db or /test/ai");
});

// ┌─────────────────────────────────────────────────────────┐
// │  Aerostack Feature Examples                              │
// │  Hit these endpoints to test your local configuration   │
// └─────────────────────────────────────────────────────────┘

// 1. Database (D1) - Persist data locally
app.get("/test/db", async (req, res) => {
  try {
    await sdk.db.query('CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, text TEXT)');
    await sdk.db.query('INSERT INTO notes (text) VALUES (?)', ['Express + Aerostack!']);
    const { results } = await sdk.db.query('SELECT * FROM notes');
    res.json({ success: true, notes: results });
  } catch (err: any) {
    res.status(500).json({ success: false, error: err.message });
  }
});

// 2. Cache (KV) - Key-value storage
app.get("/test/cache", async (req, res) => {
  try {
    const now = new Date().toISOString();
    await sdk.cache.set('express_hit', now);
    res.json({ success: true, lastHit: await sdk.cache.get('express_hit') });
  } catch (err: any) {
    res.status(500).json({ success: false, error: err.message });
  }
});

// 3. AI - Llama 3 Proxy
app.get("/test/ai", async (req, res) => {
  try {
    const result = await sdk.ai.generate('Explain Express.js in one sentence');
    res.json({ success: true, explanation: result.text });
  } catch (err: any) {
    res.status(500).json({ success: false, error: err.message });
  }
});

// We recommend using app.listen(8080) for cloudflare:node compatibility
// but if you experience "createServer not implemented" issues, you can
// use the httpServerHandler export directly.
app.listen(8080);

export default {
  ...httpServerHandler({ port: 8080 }),
  // Add Queue support to the Express template
  async queue(batch: any, env: any) {
    sdk.init(env);
    console.log("Processing background jobs via Express template...");
    for (const msg of batch.messages) {
      msg.ack();
    }
  }
};
