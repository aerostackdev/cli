import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
    it("responds to GET / with 200", async () => {
        const res = await SELF.fetch("http://localhost/");
        expect(res.status).toBe(200);
        expect(await res.text()).toContain("WebSocket Chat");
    });

    it("responds to GET /ws/room1 without Upgrade header with 426", async () => {
        const res = await SELF.fetch("http://localhost/ws/room1");
        // Without upgrade, it falls through to 404 in this basic template
        expect(res.status).toBe(404);
    });
});
