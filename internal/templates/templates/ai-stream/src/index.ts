import { sdk } from '@aerostack/sdk';

export interface Env {
  AEROSTACK_PROJECT_ID: string;
  AEROSTACK_API_KEY: string;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    sdk.init(env);
    const url = new URL(request.url);

    if (url.pathname === '/') {
      return new Response(
        'AI Streaming Worker is running.\n\n' +
        'POST /stream   — stream an AI response via Server-Sent Events\n' +
        '               Body: { "prompt": "your prompt here" }\n\n' +
        'GET  /generate — non-streaming AI response for comparison\n' +
        '               Query: ?prompt=your+prompt+here',
        { status: 200 },
      );
    }

    // ─── Non-streaming reference endpoint ───────────────────────────────────
    if (url.pathname === '/generate' && request.method === 'GET') {
      const prompt = url.searchParams.get('prompt') || 'Tell me a joke';
      const { text } = await sdk.ai.generate(prompt);
      return Response.json({ text });
    }

    // ─── SSE Streaming endpoint ──────────────────────────────────────────────
    if (url.pathname === '/stream' && request.method === 'POST') {
      let prompt = 'Tell me a short story in two sentences';
      try {
        const body: any = await request.json();
        if (body?.prompt) prompt = String(body.prompt);
      } catch { /* use default */ }

      // Set up a TransformStream so we can write SSE events asynchronously
      const { readable, writable } = new TransformStream();
      const writer = writable.getWriter();
      const encoder = new TextEncoder();

      const sendEvent = async (event: string, data: unknown) => {
        // SSE format: "event: <name>\ndata: <json>\n\n"
        const line = `event: ${event}\ndata: ${JSON.stringify(data)}\n\n`;
        await writer.write(encoder.encode(line));
      };

      // Run AI and stream the response asynchronously.
      // The response is streamed word-by-word to demonstrate real SSE output.
      // For native token streaming, swap sdk.ai.generate() with a Workers AI
      // binding call using { stream: true } (requires `ai = true` in aerostack.toml).
      const runAndStream = async () => {
        try {
          await sendEvent('start', { message: 'Generating response...' });

          const { text } = await sdk.ai.generate(prompt);

          // Stream word by word so the client sees progressive output
          const words = text.split(/(\s+)/); // split preserving whitespace tokens
          let accumulated = '';
          for (const word of words) {
            accumulated += word;
            await sendEvent('token', { token: word, text: accumulated });
          }

          await sendEvent('done', { text });
        } catch (e: any) {
          await sendEvent('error', { message: e?.message ?? 'Generation failed' });
        } finally {
          await writer.close();
        }
      };

      // Fire and forget — the readable stream stays open until close() is called
      runAndStream().catch(() => writer.abort());

      return new Response(readable, {
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache, no-transform',
          'Connection': 'keep-alive',
          'X-Accel-Buffering': 'no', // disable nginx/proxy buffering
        },
      });
    }

    return new Response('Not found', { status: 404 });
  },
};
