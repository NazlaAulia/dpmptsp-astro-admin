// Credential verification and session lifecycle, through the Go API.

import { api, type SessionInfo } from "@dpmptsp/api-client";

export type AuthResult =
  | { ok: true; session: SessionInfo }
  | { ok: false; message: string };

export async function verifyCredentials(username: string, password: string): Promise<AuthResult> {
  const res = await api.login(username, password);
  if (res.ok) return { ok: true, session: res.data };

  if (res.status === 401) return { ok: false, message: "Username atau password salah." };
  console.error("login failed:", res.status, res.error);
  return { ok: false, message: "Terjadi kesalahan pada server." };
}

/** Revokes the session in the API so the id cannot be reused. */
export async function revokeSession(sessionId: string): Promise<void> {
  const res = await api.logout(sessionId);
  if (!res.ok) console.error("logout failed:", res.status, res.error);
}
