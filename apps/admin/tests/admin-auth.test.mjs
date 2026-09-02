import test from "node:test";
import assert from "node:assert/strict";
import { authenticateAdmin } from "../src/lib/adminAuth.js";

// The previous version of this file asserted that username "admin" with
// password "admin123" logged in successfully with db: null. That test was
// pinning a backdoor in place: any attempt to remove the hardcoded credentials
// would have failed CI. This asserts the opposite.

test("the hardcoded admin/admin123 backdoor is gone", async () => {
  const result = await authenticateAdmin({
    username: "admin",
    password: "admin123",
    db: null,
  });

  assert.equal(result.success, false);
});

test("no credentials are accepted without a database", async () => {
  for (const [username, password] of [
    ["admin", "admin123"],
    ["admin", "admin"],
    ["root", "root"],
  ]) {
    const result = await authenticateAdmin({ username, password, db: null });
    assert.equal(result.success, false, `${username}/${password} was accepted`);
  }
});

test("missing credentials are rejected with a distinct message", async () => {
  const result = await authenticateAdmin({ username: "", password: "", db: null });
  assert.equal(result.success, false);
  assert.equal(result.error, "Username dan password wajib diisi.");
});

test("a matching row authenticates", async () => {
  const db = {
    execute: async () => [[{ id: 1, username: "someone" }]],
  };
  const result = await authenticateAdmin({ username: "someone", password: "pw", db });
  assert.equal(result.success, true);
});

test("no matching row does not authenticate", async () => {
  const db = { execute: async () => [[]] };
  const result = await authenticateAdmin({ username: "someone", password: "wrong", db });
  assert.equal(result.success, false);
});

test("credentials are passed as bound parameters, never interpolated", async () => {
  let capturedSql = "";
  let capturedParams = [];
  const db = {
    execute: async (sql, params) => {
      capturedSql = sql;
      capturedParams = params;
      return [[]];
    },
  };
  await authenticateAdmin({ username: "a' OR '1'='1", password: "x", db });

  assert.ok(!capturedSql.includes("OR '1'='1"), "input reached the SQL string");
  assert.deepEqual(capturedParams, ["a' OR '1'='1", "x"]);
});
