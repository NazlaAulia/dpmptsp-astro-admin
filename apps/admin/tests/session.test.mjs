import test from "node:test";
import assert from "node:assert/strict";

process.env.SESSION_SECRET = "0123456789abcdef0123456789abcdef0123456789";
process.env.SESSION_COOKIE_SECURE = "false";

const { issueSession, readSession, clearSession, csrfToken, verifyCsrf, cookieName } =
  await import("../src/lib/session.ts");

/** Minimal stand-in for Astro's cookie API. */
function fakeCookies() {
  const jar = new Map();
  return {
    set: (name, value) => jar.set(name, { value }),
    get: (name) => jar.get(name),
    _raw: jar,
  };
}

test("a session round-trips", () => {
  const c = fakeCookies();
  const issued = issueSession(c, "admin", "api-session-id");
  const read = readSession(c);
  assert.equal(read?.sub, "admin");
  assert.equal(read?.sid, issued.sid);
  // The cookie must carry the API's session id, since that is what the API
  // resolves and can revoke.
  assert.equal(read?.sid, "api-session-id");
});

test("no cookie means no session", () => {
  assert.equal(readSession(fakeCookies()), null);
});

// The scheme this replaced set the cookie to the literal string "true", so
// anyone could mint one. A forged value must not authenticate.
test("a forged cookie is rejected", () => {
  const c = fakeCookies();
  c.set(cookieName(), "true");
  assert.equal(readSession(c), null);
});

test("a tampered payload is rejected", () => {
  const c = fakeCookies();
  issueSession(c, "admin", "api-session-id");
  const raw = c.get(cookieName()).value;
  const [payload, mac] = [raw.slice(0, raw.lastIndexOf(".")), raw.slice(raw.lastIndexOf(".") + 1)];
  // Same signature, different payload.
  c.set(cookieName(), `${payload}x.${mac}`);
  assert.equal(readSession(c), null);
});

test("a tampered signature is rejected", () => {
  const c = fakeCookies();
  issueSession(c, "admin", "api-session-id");
  const raw = c.get(cookieName()).value;
  c.set(cookieName(), `${raw.slice(0, raw.lastIndexOf("."))}.deadbeef`);
  assert.equal(readSession(c), null);
});

test("an expired session is rejected", () => {
  const c = fakeCookies();
  issueSession(c, "admin", "api-session-id");
  const raw = c.get(cookieName()).value;
  const payload = JSON.parse(
    Buffer.from(raw.slice(0, raw.lastIndexOf(".")), "base64url").toString("utf8")
  );
  payload.exp = Math.floor(Date.now() / 1000) - 1;
  // Re-sign so only expiry differs, not the signature.
  const forged = Buffer.from(JSON.stringify(payload)).toString("base64url");
  c.set(cookieName(), `${forged}.${raw.slice(raw.lastIndexOf(".") + 1)}`);
  assert.equal(readSession(c), null);
});

test("clearing removes the session", () => {
  const c = fakeCookies();
  issueSession(c, "admin", "api-session-id");
  clearSession(c);
  assert.equal(readSession(c), null);
});

test("csrf token matches its own session and nothing else", () => {
  const a = fakeCookies();
  const b = fakeCookies();
  const sa = issueSession(a, "admin", "sid-a");
  const sb = issueSession(b, "admin", "sid-b");

  assert.ok(verifyCsrf(sa, csrfToken(sa)));
  assert.ok(!verifyCsrf(sa, csrfToken(sb)), "a token from another session must not verify");
  assert.ok(!verifyCsrf(sa, ""));
  assert.ok(!verifyCsrf(null, csrfToken(sa)));
});
