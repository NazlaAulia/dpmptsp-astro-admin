// Delete one announcement. Authentication and cross-origin refusal are handled
// by src/middleware.ts.
export const prerender = false;

import type { APIRoute } from "astro";
import { deleteAnnouncement } from "../../../lib/services/announcement.service";

export const DELETE: APIRoute = async ({ params }) => {
  const id = Number(params.id);
  if (!Number.isInteger(id) || id <= 0) {
    return new Response(JSON.stringify({ success: false, error: "invalid id" }), {
      status: 400,
      headers: { "content-type": "application/json" },
    });
  }

  const res = await deleteAnnouncement(id);
  return new Response(JSON.stringify({ success: res.ok }), {
    status: res.ok ? 200 : res.status,
    headers: { "content-type": "application/json" },
  });
};
