// @ts-check
import { defineConfig } from 'astro/config';
import node from '@astrojs/node';

// Full SSR (SPEC.md §2), standalone node adapter for self-hosted deployment.
//
// No Tailwind and no global stylesheet: the admin panel styles itself entirely
// with scoped <style> blocks. Importing Tailwind would apply its preflight
// reset and restyle every screen for no benefit.
export default defineConfig({
  output: 'server',

  // Astro's built-in CSRF check compares the Origin header against
  // context.url.origin. Under the node adapter url.origin is always
  // "http://localhost" no matter what the request says, so the comparison can
  // never succeed and every browser form POST — including login — is answered
  // with 403 "Cross-site POST form submissions are forbidden". Verified
  // directly against a built server.
  //
  // src/middleware.ts performs the equivalent check correctly, using
  // Sec-Fetch-Site with an Origin/Host fallback.
  security: { checkOrigin: false },

  adapter: node({ mode: 'standalone' }),
});
