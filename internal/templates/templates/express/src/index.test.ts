import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Express Worker", () => {
  it("responds to GET /", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.text()).toContain("Express");
  });
});
