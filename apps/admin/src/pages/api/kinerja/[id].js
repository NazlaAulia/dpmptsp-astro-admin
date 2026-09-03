// Delete one kinerja (twdata) row.
export const prerender = false;

import { deletePerformanceDoc } from "../../../lib/services/content.service";

export async function DELETE({ params }) {
  const { id } = params;

  if (!/^\d+$/.test(id ?? "")) {
    return new Response(JSON.stringify({ success: false, error: "Invalid id." }), {
      status: 400,
      headers: { "content-type": "application/json" },
    });
  }

  const result = await deletePerformanceDoc(Number(id));

  return new Response(
    JSON.stringify(result.ok ? { success: true } : { success: false, error: result.message }),
    { status: result.ok ? 200 : 502, headers: { "content-type": "application/json" } }
  );
}
