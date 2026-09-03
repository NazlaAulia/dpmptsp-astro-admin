// Contact form rules that belong in front of the API.

import { findByTicket, submitMessage } from "../repositories/contact.repository";
import type { ContactStatus } from "@dpmptsp/api-client";
import type { ContactForm } from "../contact-form";

export { readForm, validate } from "../contact-form";
export type { ContactForm } from "../contact-form";

export async function submit(form: ContactForm, clientIp?: string) {
  return submitMessage(form, clientIp);
}

export async function track(tiket: string): Promise<ContactStatus | null> {
  const code = tiket.trim();
  if (!code) return null;
  return findByTicket(code);
}
