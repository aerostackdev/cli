import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
  // POST /webhook attempts a DB write — expect 500 when no DATABASE_URL is set
  it("POST /webhook returns 500 when no DATABASE_URL (no DB configured)", async () => {
    const res = await SELF.fetch("http://localhost/webhook", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ event: "test", data: { id: 1 } }),
    });
    // 500 = DB error (no DATABASE_URL) — connection failed, caught by Hono's error handler
    expect(res.status).toBe(500);
  });
});
