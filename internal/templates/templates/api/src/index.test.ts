import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
  it("responds to GET /", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.text()).toContain("Welcome");
  });

  it("responds to GET /health", async () => {
    const res = await SELF.fetch("http://localhost/health");
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ status: "ok", template: "api" });
  });

  it("handles POST /notes and GET /notes", async () => {
    // Create note
    const postRes = await SELF.fetch("http://localhost/notes", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text: "test note" })
    });
    expect(postRes.status).toBe(201);
    const postJson = (await postRes.json()) as any;
    expect(postJson.success).toBe(true);

    // Get notes
    const getRes = await SELF.fetch("http://localhost/notes");
    expect(getRes.status).toBe(200);
    const getArray = (await getRes.json()) as any[];
    expect(Array.isArray(getArray)).toBe(true);
    expect(getArray.length).toBeGreaterThan(0);
    expect(getArray[0].text).toBe("test note");
  });

  it("responds to GET /users/:id", async () => {
    const res = await SELF.fetch("http://localhost/users/42");
    expect(res.status).toBe(200);
    const json = (await res.json()) as { id: string; name: string };
    expect(json.id).toBe("42");
    expect(json.name).toContain("42");
  });
});
