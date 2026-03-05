import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
  it("responds to GET / with 200 and usage instructions", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.text()).toContain("AI Streaming Worker");
  });

  it("POST /stream returns SSE content-type", async () => {
    const res = await SELF.fetch("http://localhost/stream", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prompt: "Say hi" }),
    });
    // SSE stream starts immediately with 200 and event-stream content type
    expect(res.status).toBe(200);
    expect(res.headers.get("Content-Type")).toContain("text/event-stream");
  });

  it("GET /generate returns JSON with text or error", async () => {
    const res = await SELF.fetch("http://localhost/generate?prompt=Hello");
    // May return 200 (with AI configured) or 500 (no API key) — both are valid
    expect([200, 500]).toContain(res.status);
    if (res.status === 200) {
      const json = (await res.json()) as any;
      expect(json.text).toBeDefined();
    }
  });

  it("returns 404 for unknown routes", async () => {
    const res = await SELF.fetch("http://localhost/unknown");
    expect(res.status).toBe(404);
  });
});
