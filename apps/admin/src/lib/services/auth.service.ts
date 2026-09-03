// Login use case.
//
// Astro authenticates and holds the session cookie; the API decides whether the
// credentials are valid.

export { verifyCredentials } from "../repositories/auth.repository";
