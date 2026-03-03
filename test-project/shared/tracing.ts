// Distributed tracing — import via: import { getTraceId, propagateTraceId } from "@shared/tracing"
// Use trace IDs to correlate logs across services

const TRACE_HEADER = "X-Trace-ID";

/** Get trace ID from request or generate one. Use in handlers and log it. */
export function getTraceId(request: Request): string {
  const existing = request.headers.get(TRACE_HEADER);
  if (existing) return existing;
  return crypto.randomUUID();
}

/** Add trace ID to fetch init for outbound requests (propagate to downstream services). */
export function propagateTraceId(request: Request, init?: RequestInit): RequestInit {
  const traceId = getTraceId(request);
  const headers = new Headers(init?.headers);
  headers.set(TRACE_HEADER, traceId);
  return { ...init, headers };
}
