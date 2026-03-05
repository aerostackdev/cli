import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
  it("responds to GET / with 200 and usage info", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.text()).toContain("Hello from a blank Neon worker");
  });

  // /test/db, /test/cache, /test/ai require real credentials
  // They are integration-tested after deploying with DATABASE_URL and AEROSTACK_API_KEY set
});
