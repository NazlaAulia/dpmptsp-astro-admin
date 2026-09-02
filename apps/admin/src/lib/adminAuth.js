// Admin credential check.
//
// Two things were removed here, and neither should come back:
//
//   1. A hardcoded backdoor. Any request with username "admin" and password
//      "admin123" was accepted before the database was consulted at all, and
//      it worked even with no database configured. A passing test asserted
//      this behaviour, so the test was enforcing the vulnerability; it has
//      been replaced.
//
//   2. Nothing else changed about how passwords are stored, and that remains
//      the outstanding problem: the `user` table holds PLAINTEXT passwords and
//      this function still compares them literally in SQL. Fixing that means
//      hashing (argon2id) and a migration, which belongs with the Go API — it
//      cannot be done from here without locking every existing admin out.
//      Treat every credential in that table as compromised in the meantime.

export async function authenticateAdmin({ username, password, db }) {
  const normalizedUsername = String(username ?? "").trim();
  const normalizedPassword = String(password ?? "").trim();

  if (!normalizedUsername || !normalizedPassword) {
    return {
      success: false,
      error: "Username dan password wajib diisi.",
    };
  }

  if (!db) {
    return {
      success: false,
      error: "Username atau password salah!",
    };
  }

  try {
    const [rows] = await db.execute(
      "SELECT * FROM `user` WHERE username = ? AND password = ?",
      [normalizedUsername, normalizedPassword]
    );

    if (Array.isArray(rows) && rows.length > 0) {
      return { success: true };
    }

    return {
      success: false,
      error: "Username atau password salah!",
    };
  } catch (error) {
    console.error("LOGIN ERROR:", error);
    return {
      success: false,
      error: "Terjadi kesalahan pada server.",
    };
  }
}
