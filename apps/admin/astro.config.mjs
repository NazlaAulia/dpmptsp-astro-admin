// @ts-check
import { defineConfig } from 'astro/config';
import node from '@astrojs/node';

// Full SSR, standalone node adapter for self-hosted deployment.
export default defineConfig({
  output: 'server',

  // Astro's built-in CSRF check compares the Origin header against
  // context.url.origin. Under the node adapter url.origin is always
  // "http://localhost" no matter what the request says, so the comparison can
  // never succeed and every browser form POST — including login — is answered
  // with 403 "Cross-site POST form submissions are forbidden". Verified
  // directly against a built server.
  security: { checkOrigin: false },

  adapter: node({ mode: 'standalone' }),
});
