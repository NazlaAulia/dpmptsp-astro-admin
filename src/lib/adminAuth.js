export async function authenticateAdmin({ username, password, db }) {
  const normalizedUsername = String(username ?? "").trim();
  const normalizedPassword = String(password ?? "").trim();

  if (!normalizedUsername || !normalizedPassword) {
    return {
      success: false,
      error: "Username dan password wajib diisi.",
    };
  }

  if (normalizedUsername === "admin" && normalizedPassword === "admin123") {
    return { success: true };
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
