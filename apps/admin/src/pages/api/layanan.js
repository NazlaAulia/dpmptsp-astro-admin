export const prerender = false;

import {
  createServiceLocation,
  updateServiceLocation,
  uploadFile,
} from "../../lib/services/content.service";

/**
 * Stores a base64 data URL through the API and returns its key.
 *
 * The client still sends the image inside the JSON body. That inflates the
 * payload by a third and holds the whole file in a string; converting the form
 * to multipart is a separate change to the page's script.
 */
async function storeDataUrl(value) {
  if (typeof value !== "string" || !value.startsWith("data:image")) {
    return { ok: true, key: value ?? "" };
  }

  const match = value.match(/^data:image\/([a-zA-Z0-9]+);base64,(.+)$/);
  if (!match) return { ok: true, key: "" };

  const [, ext, data] = match;
  const bytes = Buffer.from(data, "base64");
  const file = new File([bytes], `layanan.${ext}`, { type: `image/${ext}` });

  const uploaded = await uploadFile(file, "layanan", "public");
  return uploaded.ok ? { ok: true, key: uploaded.key } : { ok: false, message: uploaded.message };
}

function json(body, status) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

async function readBody(request) {
  try {
    return await request.json();
  } catch {
    return null;
  }
}

export async function POST({ request }) {
  const body = await readBody(request);
  if (!body) return json({ success: false, message: "Body harus JSON." }, 400);

  const image = await storeDataUrl(body.gambar);
  if (!image.ok) return json({ success: false, message: image.message }, 400);

  const result = await createServiceLocation({
    judul: body.judul,
    deskripsi: body.deskripsi || "",
    gambar: image.key,
    alt_text: body.alt_text,
    lokasi: body.lokasi,
    alamat: body.alamat,
    warna: "blue",
  });

  return result.ok
    ? json({ success: true, data: result.data }, 201)
    : json({ success: false, message: result.message }, 502);
}

export async function PUT({ request, url }) {
  const id = Number(url.searchParams.get("id"));
  if (!Number.isInteger(id) || id <= 0) {
    return json({ success: false, message: "ID tidak valid." }, 400);
  }

  const body = await readBody(request);
  if (!body) return json({ success: false, message: "Body harus JSON." }, 400);

  const image = await storeDataUrl(body.gambar);
  if (!image.ok) return json({ success: false, message: image.message }, 400);

  const payload = {
    judul: body.judul,
    deskripsi: body.deskripsi || "",
    alt_text: body.alt_text,
    lokasi: body.lokasi,
    alamat: body.alamat,
  };
  // Only overwrite the image when a new one was actually supplied; an empty
  // value here would blank the existing picture.
  if (image.key) payload.gambar = image.key;

  const result = await updateServiceLocation(id, payload);
  return result.ok
    ? json({ success: true, data: result.data }, 200)
    : json({ success: false, message: result.message }, 502);
}
