// Delete one kinerja (twdata) row.
//
// Replaces src/pages/api/[id].js, which was a catch-all directly under /api:
// any DELETE to /api/<anything-without-a-more-specific-route> ran
// `DELETE FROM twdata WHERE id = ?` with that path segment, unauthenticated
// and unvalidated. Namespacing it removes the accidental blast radius.
//
// Authentication is enforced by src/middleware.ts, which also refuses any
// cross-origin write.
export const prerender = false;

import { db } from "../../../lib/db.js";

export async function DELETE({ params }) {
  const { id } = params;

  if (!/^\d+$/.test(id ?? "")) {
    return new Response(JSON.stringify({ success: false, error: "Invalid id." }), {
      status: 400,
      headers: { "content-type": "application/json" },
    });
  }

  await db.query("DELETE FROM twdata WHERE id = ?", [id]);

  return new Response(JSON.stringify({ success: true }), {
    status: 200,
    headers: { "content-type": "application/json" },
  });
}
