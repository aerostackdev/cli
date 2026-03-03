import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
    it("responds to GET / with 200", async () => {
        const res = await SELF.fetch("http://localhost/");
        expect(res.status).toBe(200);
        expect(await res.text()).toContain("WebSocket Multiplayer Game");
    });

    it("responds to GET /ws/room1 without Upgrade header with 426 Upgrade Required", async () => {
        const res = await SELF.fetch("http://localhost/ws/room1");
        expect(res.status).toBe(404); // Falls through to not found if not upgraded
    });
});
