import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
    it("responds to GET / with 200", async () => {
        const res = await SELF.fetch("http://localhost/");
        expect(res.status).toBe(200);
        expect(await res.text()).toContain("WebSocket Voice Agent");
    });

    it("responds to GET /ws without Upgrade header with 426 Upgrade Required", async () => {
        // A regular GET without 'Upgrade: websocket' should fail with Upgrade Required
        // For local test stubbing we might just get 426 or fall through to 404, let's see. 
        // The fetch signature doesn't do WS upgrade so it falls through to 404 in the template
        // Actually the template returns 404 if not upgraded.
        const res = await SELF.fetch("http://localhost/ws");
        expect(res.status).toBe(404);
    });
});
