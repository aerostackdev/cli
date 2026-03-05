import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("API Worker", () => {
  it("responds to GET / with 200", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.text()).toContain("Aerostack Multi-Function User");
  });

  it("GET /api/data returns JSON with message and timestamp", async () => {
    const res = await SELF.fetch("http://localhost/api/data");
    expect(res.status).toBe(200);
    const json = (await res.json()) as any;
    expect(json.message).toBeDefined();
    expect(json.timestamp).toBeDefined();
  });
});
