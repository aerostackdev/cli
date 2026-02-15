// Express on Cloudflare Workers (nodejs_compat + httpServerHandler)
import express from "express";
import { httpServerHandler } from "cloudflare:node";

const app = express();
app.use(express.json());

app.get("/", (req, res) => {
  res.send("Hello from Express on Aerostack!");
});

app.get("/users/:id", (req, res) => {
  res.json({ id: req.params.id, name: `User ${req.params.id}` });
});

app.listen(8080);
export default httpServerHandler({ port: 8080 });
