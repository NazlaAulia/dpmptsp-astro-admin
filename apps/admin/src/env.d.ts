/// <reference types="astro/client" />

import type { Session } from "./lib/session";

declare global {
  namespace App {
    interface Locals {
      /** Set by middleware on every request. Null when unauthenticated. */
      session: Session | null;
      /** CSRF token for the current session, or null when unauthenticated. */
      csrfToken: string | null;
    }
  }
}

export {};
