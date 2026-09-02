// @ts-check
import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';
import node from '@astrojs/node';

// Full SSR (SPEC.md §2). The Vercel adapter was replaced with the standalone
// node adapter: this app opens a MySQL pool at 'localhost' and writes uploaded
// files into public/ at request time, neither of which works on Vercel's
// read-only serverless filesystem. It has only ever been able to run on a
// persistent host, which is what SPEC.md §8/§9 target anyway.
export default defineConfig({
  output: 'server',

  vite: {
    plugins: [tailwindcss()],
  },

  adapter: node({ mode: 'standalone' }),
});
