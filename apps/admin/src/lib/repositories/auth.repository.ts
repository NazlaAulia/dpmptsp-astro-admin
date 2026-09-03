// Credential verification, through the Go API.

import { api, type User } from "@dpmptsp/api-client";

export type AuthResult =
  | { ok: true; user: User }
  | { ok: false; message: string };

export async function verifyCredentials(username: string, password: string): Promise<AuthResult> {
  const res = await api.login(username, password);
  if (res.ok) return { ok: true, user: res.data };

  if (res.status === 401) {
    return { ok: false, message: "Username atau password salah." };
  }
  console.error("login failed:", res.status, res.error);
  return { ok: false, message: "Terjadi kesalahan pada server." };
}
