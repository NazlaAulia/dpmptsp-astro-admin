// The single authentication guard for the admin app.

import { defineMiddleware } from "astro:middleware";
import { readSession, csrfToken } from "./lib/session";
import { runWithContext } from "./lib/request-context";

/** Exact paths reachable without a session. Everything else is guarded. */
const PUBLIC_PATHS = new Set(["/admin/login", "/admin/logout"]);

/** Prefixes for build output and static assets, which carry no data. */
const PUBLIC_PREFIXES = ["/_astro/", "/_image", "/favicon"];

const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);

/**
 * Is this a cross-site write?
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

export const onRequest = defineMiddleware(async (context, next) =>
  runWithContext({ sessionId: readSession(context.cookies)?.sid }, async () => {
  const { request, url, cookies } = context;
  const method = request.method.toUpperCase();

  const session = readSession(cookies);
  context.locals.session = session;
  context.locals.csrfToken = session ? csrfToken(session) : null;

  // Cross-origin write protection, applied before anything reads a body.
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
  })
);
