import express from "express";
import { httpServerHandler } from "cloudflare:node";
import { sdk } from "@aerostack/sdk";

const app = express();
app.use(express.json());

// Workers environment middleware
app.use((req, res, next) => {
  // @ts-ignore - Workers environment is attached to req in cloudflare:node
  const env = req.env || (req as any).env;
  if (env) sdk.init(env);
  next();
});

app.get("/", (req, res) => {
  res.send("Hello from Express on Aerostack!");
});

app.get("/users/:id", (req, res) => {
  res.json({ id: req.params.id, name: `User ${req.params.id}` });
});

app.listen(8080);
export default httpServerHandler({ port: 8080 });
