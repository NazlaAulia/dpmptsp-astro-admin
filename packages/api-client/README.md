# @dpmptsp/api-client

Server-side client for the Go API.

**Never import this from browser-facing code.** It carries the internal service
key. The API is reachable only on the internal network: the browser talks to
Astro, and Astro talks to this.

## Types are generated, the transport is not

`src/generated/` is produced from `apps/api/openapi.yaml` and must not be edited
by hand. Regenerate after changing the spec:

    pnpm --filter @dpmptsp/api-client generate

`src/index.ts` is hand-written and stays that way. Generated clients tend to
bury the things that matter operationally — timeouts, how failures are
surfaced, which header carries the service key — and those are exactly the parts
worth reading here.

## Failure is a value, not an exception

Calls return `{ ok: true, data }` or `{ ok: false, status, error }` and never
throw. An uncaught throw in Astro frontmatter turns one slow dependency into a
500 for the entire page; returning a result lets the caller decide whether that
section is essential. Every request is also bounded by a timeout, because
without one an unhealthy API becomes hung SSR workers and the site looks down
rather than degraded.
