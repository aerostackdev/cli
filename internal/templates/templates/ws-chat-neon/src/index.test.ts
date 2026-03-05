import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
  it("responds to GET / with 200", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.text()).toContain("WebSocket Chat (Neon)");
  });

  it("GET /ws/room1 without Upgrade header returns 404", async () => {
    const res = await SELF.fetch("http://localhost/ws/room1");
    expect(res.status).toBe(404);
  });

  // History endpoint requires Neon — expects 500 with no DATABASE_URL set
  it("GET /rooms/room1/history returns 500 when no DATABASE_URL", async () => {
    const res = await SELF.fetch("http://localhost/rooms/room1/history");
    expect(res.status).toBe(500);
  });
});
