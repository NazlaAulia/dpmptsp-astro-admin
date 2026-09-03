// Delete one kinerja (twdata) row.
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
