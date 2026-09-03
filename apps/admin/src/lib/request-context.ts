// Per-request context for the admin app.
//
// The API authorizes writes on the session id, so every outgoing call needs it.
// Threading it through every page, service and repository would touch about
// twenty call sites and be easy to forget in one of them — and forgetting means
// a write that silently fails authorization.
//
// AsyncLocalStorage keeps it bound to the request instead. Astro's node adapter
// runs on Node, and each request is its own async context, so there is no risk
// of one request seeing another's session.

import { AsyncLocalStorage } from "node:async_hooks";

type RequestContext = { sessionId?: string };

const storage = new AsyncLocalStorage<RequestContext>();

export function runWithContext<T>(ctx: RequestContext, fn: () => T): T {
  return storage.run(ctx, fn);
}

/** The current session id, or undefined outside a request. */
export function currentSessionId(): string | undefined {
  return storage.getStore()?.sessionId;
}
