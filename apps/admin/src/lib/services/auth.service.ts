// Login and logout.
//
// Astro holds the cookie; the API owns the session and decides what it may do.

export { verifyCredentials, revokeSession } from "../repositories/auth.repository";
