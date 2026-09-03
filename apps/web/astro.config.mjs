// @ts-check
import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';
import node from '@astrojs/node';

// Full SSR. The Vercel adapter was replaced with the standalone
// node adapter: this app opens a MySQL pool at 'localhost' and writes uploaded
export default defineConfig({
  output: 'server',

  vite: {
    plugins: [tailwindcss()],
  },

  adapter: node({ mode: 'standalone' }),
});
