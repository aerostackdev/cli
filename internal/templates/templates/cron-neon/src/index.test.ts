import { SELF } from "cloudflare:test";
import { describe, expect, it } from "vitest";

describe("Worker", () => {
  it("responds to GET / with 200 — fetch handler active", async () => {
    const res = await SELF.fetch("http://localhost/");
    expect(res.status).toBe(200);
    expect(await res.text()).toContain("Cron worker is active");
  });

  // Scheduled handler is tested via the scheduled() export — not via HTTP
  // Use `wrangler dev --test-scheduled` to trigger it manually during development
});
