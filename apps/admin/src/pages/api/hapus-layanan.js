export const prerender = false;

import { deleteServiceLocation } from "../../lib/services/content.service";

export async function DELETE({ request }) {
  let id;
  try {
    ({ id } = await request.json());
  } catch {
    return json({ success: false, message: "Body harus JSON." }, 400);
  }

  if (!Number.isInteger(Number(id))) {
    return json({ success: false, message: "ID tidak valid." }, 400);
  }

  const result = await deleteServiceLocation(Number(id));
  return result.ok
    ? json({ success: true }, 200)
    : json({ success: false, message: result.message }, 502);
}

function json(body, status) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
