// Admin session handling.
//
// Replaces the previous scheme, which set the cookie `admin-auth` to the
// literal string "true". That value was constant, so anyone could mint a valid
// session with `curl -b 'admin-auth=true'`; httpOnly did not help, because an
// attacker never needed to read the cookie, only to send one.
//
// This issues a signed, expiring cookie instead. The payload is HMAC-SHA256'd
// with SESSION_SECRET and compared in constant time, so it cannot be forged or
// extended without the secret.
//
// KNOWN LIMITATION, deliberate at this stage: the session is self-contained,
// not backed by a server-side store, because Redis arrives with apps/api. That
// means logout clears the browser's cookie but cannot revoke a token someone
// has already copied — it stays valid until it expires. When the Go API lands,
// the payload becomes an opaque id resolved against Redis and this limitation
// disappears. Keep the TTL short until then.

import crypto from "node:crypto";
import type { AstroCookies } from "astro";

const MAX_AGE_SECONDS = 60 * 60 * 8; // 8 hours

export type Session = {
  /** Who logged in. Carries no role: authorization is never Astro's decision. */
  sub: string;
  /** Random per-session id. CSRF tokens are derived from it. */
  sid: string;
  /** Issued at, epoch seconds. */
  iat: number;
  /** Expires at, epoch seconds. */
  exp: number;
};

/**
 * Read lazily rather than at module scope: throwing during import would fail
 * the build, not just the request that actually needs a session.
 */
function secret(): string {
  const s = process.env.SESSION_SECRET;
  if (!s || s.length < 32) {
    throw new Error(
      "SESSION_SECRET must be set to at least 32 characters. " +
        "Generate one with: openssl rand -base64 48"
    );
  }
  return s;
}

/** Secure cookies are required in production and impossible over plain-HTTP dev. */
function isSecure(): boolean {
  const explicit = process.env.SESSION_COOKIE_SECURE;
  if (explicit !== undefined) return explicit !== "false";
  return process.env.NODE_ENV !== "development";
}

/**
 * The __Host- prefix makes the browser itself enforce Secure, Path=/ and the
 * absence of Domain — free hardening. It is only legal on a secure cookie, so
 * over plain-HTTP dev the browser would silently drop it and login would look
 * broken for no visible reason. Hence the conditional name.
 */
export function cookieName(): string {
  return isSecure() ? "__Host-dpmptsp_admin" : "dpmptsp_admin";
}

function b64url(buf: Buffer | string): string {
  return Buffer.from(buf).toString("base64url");
}

function sign(payload: string): string {
  return crypto.createHmac("sha256", secret()).update(payload).digest("base64url");
}

/** Timing-safe compare that tolerates length mismatch. */
function safeEqual(a: string, b: string): boolean {
  const ab = Buffer.from(a);
  const bb = Buffer.from(b);
  if (ab.length !== bb.length) return false;
  return crypto.timingSafeEqual(ab, bb);
}

export function issueSession(cookies: AstroCookies, sub: string): Session {
  const now = Math.floor(Date.now() / 1000);
  const session: Session = {
    sub,
    sid: crypto.randomBytes(32).toString("base64url"),
    iat: now,
    exp: now + MAX_AGE_SECONDS,
  };
  const payload = b64url(JSON.stringify(session));
  cookies.set(cookieName(), `${payload}.${sign(payload)}`, {
    httpOnly: true,
    secure: isSecure(),
    sameSite: "lax",
    path: "/",
    maxAge: MAX_AGE_SECONDS,
  });
  return session;
}

export function readSession(cookies: AstroCookies): Session | null {
  const raw = cookies.get(cookieName())?.value;
  if (!raw) return null;

  const dot = raw.lastIndexOf(".");
  if (dot < 1) return null;

  const payload = raw.slice(0, dot);
  const mac = raw.slice(dot + 1);
  if (!safeEqual(mac, sign(payload))) return null;

  let session: Session;
  try {
    session = JSON.parse(Buffer.from(payload, "base64url").toString("utf8"));
  } catch {
    return null;
  }

  if (typeof session?.sub !== "string" || typeof session?.exp !== "number") return null;
  if (Math.floor(Date.now() / 1000) >= session.exp) return null;

  return session;
}

export function clearSession(cookies: AstroCookies): void {
  cookies.set(cookieName(), "", {
    httpOnly: true,
    secure: isSecure(),
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });
}

/**
 * CSRF token bound to the session, so no second cookie and no server-side
 * store is needed: the value is recomputable from the session id alone.
 */
export function csrfToken(session: Session): string {
  return crypto
    .createHmac("sha256", secret())
    .update(`${session.sid}:csrf`)
    .digest("base64url");
}

export function verifyCsrf(session: Session | null, submitted: unknown): boolean {
  if (!session || typeof submitted !== "string" || submitted.length === 0) return false;
  return safeEqual(submitted, csrfToken(session));
}
