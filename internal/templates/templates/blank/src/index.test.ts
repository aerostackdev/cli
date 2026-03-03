import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
  it("responds with ok status at root", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ status: "ok", template: "blank" });
  });

  it("handles /test/db", async () => {
    const res = await SELF.fetch("http://localhost/test/db");
    expect(res.status).toBe(200);
    const body: any = await res.json();
    expect(body.success).toBe(true);
    expect(body.notes.length).toBeGreaterThan(0);
  });

  it("handles /test/cache", async () => {
    const res = await SELF.fetch("http://localhost/test/cache");
    expect(res.status).toBe(200);
    const body: any = await res.json();
    expect(body.success).toBe(true);
    expect(body.lastVisit).toBeDefined();
  });
});
