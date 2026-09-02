// src/utils/slugify.js
export function slugify(text) {
  if (!text) return "";

  return text
    .toString()
    .toLowerCase()
    .trim()
    .replace(/\s+/g, "-")     // spasi → -
    .replace(/[^\w\-]/g, ""); // hapus karakter aneh
}
