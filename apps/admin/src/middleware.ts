// The single authentication guard for the admin app.
//
// Astro runs onRequest for every SSR request, including endpoints and dynamic
// routes, which is exactly why the check belongs here. Previously it was
// copy-pasted into 2 of 22 admin pages and into none of the 8 API routes, so
// anonymous callers could render every management screen and POST to every
// write endpoint.
//
// THE RULE: Astro answers one question — is this request carrying a valid,
// unexpired session? It never answers "may this subject do this?". That is the
// Go API's job (SPEC.md §7 and §11.6). There must be no role checks here.

import { defineMiddleware } from "astro:middleware";
import { readSession, csrfToken } from "./lib/session";

/**
 * Exact paths reachable without a session. Everything else is guarded, so a
 * page added tomorrow is protected by default rather than by remembering to
 * add it to a list.
 */
const PUBLIC_PATHS = new Set(["/admin/login", "/admin/logout"]);

/** Prefixes for build output and static assets, which carry no data. */
const PUBLIC_PREFIXES = ["/_astro/", "/_image", "/favicon"];

const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);

/**
 * Is this a cross-site write?
 *
 * Sec-Fetch-Site is set by the browser and cannot be forged by an attacking
 * page, so it is the authoritative signal when present. The Origin/Host
 * comparison is only a fallback for clients that do not send it.
 *
 * Both are useless against a non-browser client such as curl, which can send
 * anything — but a non-browser client has no victim's cookies to ride on, and
 * the session check below is what stops it.
 */
function isCrossSite(request: Request): boolean {
  const site = request.headers.get("sec-fetch-site");
  if (site !== null) {
    return site !== "same-origin" && site !== "none";
  }

  const origin = request.headers.get("origin");
  if (origin === null) return false;

  // x-forwarded-host is set by our own gateway, which strips any client-supplied
  // copy; host is the direct case.
  const host = request.headers.get("x-forwarded-host") ?? request.headers.get("host");
  if (host === null) return true;

  try {
    return new URL(origin).host !== host;
  } catch {
    return true;
  }
}


function isPublic(pathname: string): boolean {
  if (PUBLIC_PATHS.has(pathname)) return true;
  return PUBLIC_PREFIXES.some((p) => pathname.startsWith(p));
}

function wantsJson(request: Request): boolean {
  return (request.headers.get("accept") ?? "").includes("application/json");
}

export const onRequest = defineMiddleware(async (context, next) => {
  const { request, url, cookies } = context;
  const method = request.method.toUpperCase();

  const session = readSession(cookies);
  context.locals.session = session;
  context.locals.csrfToken = session ? csrfToken(session) : null;

  // Cross-origin write protection, applied before anything reads a body.
  //
  // SameSite=Lax already stops a cross-site form POST from carrying the session
  // cookie, so this is defence in depth.
  //
  // Note this deliberately does NOT compare against url.origin. Under the node
  // adapter url.origin is always "http://localhost" regardless of the request,
  // so any such comparison rejects every real browser request. That is exactly
  // why Astro's own security.checkOrigin is disabled in astro.config.mjs — it
  // makes that comparison and 403s every login.
  if (!SAFE_METHODS.has(method)) {
    if (isCrossSite(request)) {
      return new Response("Cross-origin request refused.", { status: 403 });
    }
  }

  if (isPublic(url.pathname)) {
    return next();
  }

  if (!session) {
    if (method !== "GET" || wantsJson(request)) {
      return new Response(JSON.stringify({ error: "Unauthenticated." }), {
        status: 401,
        headers: { "content-type": "application/json" },
      });
    }
    const next_ = encodeURIComponent(url.pathname + url.search);
    return context.redirect(`/admin/login?next=${next_}`);
  }

  const response = await next();

  // A logged-in admin page must never be cached by the browser or by anything
  // between it and the browser.
  response.headers.set("cache-control", "private, no-store");
  response.headers.set("x-frame-options", "DENY");
  response.headers.set("referrer-policy", "same-origin");
  return response;
});
