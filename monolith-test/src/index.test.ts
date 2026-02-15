import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
  it("responds to GET /", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.text()).toContain("Welcome");
  });

  it("responds to GET /users/:id", async () => {
    const res = await SELF.fetch("http://localhost/users/42");
    expect(res.status).toBe(200);
    const json = (await res.json()) as { id: string; name: string };
    expect(json.id).toBe("42");
    expect(json.name).toContain("42");
  });
});
