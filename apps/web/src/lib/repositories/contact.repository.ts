// Public contact form data access — through the Go API.

import { api } from "@dpmptsp/api-client";
import type { ContactMessageInput, ContactStatus, ContactTicket } from "@dpmptsp/api-client";

export type SubmitResult =
  | { ok: true; ticket: ContactTicket }
  | { ok: false; status: number; error: string };

export async function submitMessage(
  input: ContactMessageInput,
  clientIp?: string
): Promise<SubmitResult> {
  const res = await api.submitContact(input, clientIp);
  if (!res.ok) return { ok: false, status: res.status, error: res.error };
  return { ok: true, ticket: res.data };
}

export async function findByTicket(tiket: string): Promise<ContactStatus | null> {
  const res = await api.trackContact(tiket);
  if (!res.ok) {
    // 404 is the normal answer to a wrong code, not something to log.
    if (res.status !== 404) console.error("trackContact failed:", res.status, res.error);
    return null;
  }
  return res.data;
}
