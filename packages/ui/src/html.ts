// HTML helpers.

/**
 * Escapes text for interpolation into an HTML string.
 *
 * Astro escapes template interpolations automatically, so this is only for the
 * places that still build markup with innerHTML. Those should shrink: the right
 * fix is to render server-side, not to escape more carefully. This exists for
 * the handful of genuinely client-rendered fragments that remain.
 */
export function escapeHtml(value: unknown): string {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

/** Strips tags and truncates on a word boundary. */
export function excerpt(html: string | null | undefined, max = 200): string {
  const text = String(html ?? "")
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
  if (text.length <= max) return text;
  const cut = text.slice(0, max);
  const lastSpace = cut.lastIndexOf(" ");
  return (lastSpace > 0 ? cut.slice(0, lastSpace) : cut) + "…";
}
