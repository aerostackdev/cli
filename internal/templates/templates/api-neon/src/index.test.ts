import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
    it("responds to GET /", async () => {
        const res = await SELF.fetch("http://localhost/");
        expect(res.status).toBe(200);
        expect(await res.text()).toContain("Welcome to your Aerostack API");
    });

    it("responds to GET /health", async () => {
        const res = await SELF.fetch("http://localhost/health");
        expect(res.status).toBe(200);
        expect(await res.json()).toEqual({ status: "ok", template: "api-neon" });
    });

    // the /users and DB logic can't be fully tested locally without real neon credentials,
    // but we test that the endpoint responds.
    it("responds to GET /users with failure when no PG url", async () => {
        const res = await SELF.fetch("http://localhost/users");
        expect(res.status).toBe(500);
        const json = (await res.json()) as any;
        expect(json.error).toBeDefined();
    });
});
