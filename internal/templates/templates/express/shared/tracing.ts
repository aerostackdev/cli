// Distributed tracing — import via: import { getTraceId, propagateTraceId } from "@shared/tracing"
const TRACE_HEADER = "X-Trace-ID";
export function getTraceId(request: Request): string {
  const existing = request.headers.get(TRACE_HEADER);
  if (existing) return existing;
  return crypto.randomUUID();
}
export function propagateTraceId(request: Request, init?: RequestInit): RequestInit {
  const traceId = getTraceId(request);
  const headers = new Headers(init?.headers);
  headers.set(TRACE_HEADER, traceId);
  return { ...init, headers };
}
