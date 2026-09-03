// Public contact form: submit a message, and look one up by ticket code.
//
// The browser only ever talks to this route; the Go API is not reachable from
// it. Rate limiting lives in the API, which sees every instance's traffic —
// the check here is a cheap first cut per process.
export const prerender = false;

import type { APIRoute } from "astro";
import { readForm, submit, track, validate } from "../../lib/services/contact.service";

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json", "cache-control": "no-store" },
  });
}

/**
 * cloudflared and the gateway both prepend, so the first entry is the client.
 * Falls back to the socket address, which is the docker bridge in production
 * and therefore shared — a fallback, not a source of truth.
 */
function clientIp(request: Request, address?: string): string {
  const fwd = request.headers.get("x-forwarded-for") ?? request.headers.get("cf-connecting-ip");
  if (fwd) return fwd.split(",")[0]!.trim();
  return address ?? "";
}

// In-process burst guard. Resets with the process, so it is a courtesy check,
// not the real limit.
const BURST_WINDOW_MS = 60_000;
const BURST_MAX = 3;
const recent = new Map<string, number[]>();

function overBurstLimit(ip: string): boolean {
  if (!ip) return false;
  const now = Date.now();
  const hits = (recent.get(ip) ?? []).filter((t) => now - t < BURST_WINDOW_MS);
  hits.push(now);
  recent.set(ip, hits);

  // Keep the map from growing without bound on a long-running process.
  if (recent.size > 5000) {
    for (const [key, times] of recent) {
      if (times.every((t) => now - t >= BURST_WINDOW_MS)) recent.delete(key);
    }
  }
  return hits.length > BURST_MAX;
}

export const POST: APIRoute = async ({ request, clientAddress }) => {
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return json({ success: false, message: "Format permintaan tidak valid." }, 400);
  }

  const form = readForm(body);
  const invalid = validate(form);
  if (invalid) return json({ success: false, message: invalid }, 422);

  const ip = clientIp(request, clientAddress);
  if (overBurstLimit(ip)) {
    return json({ success: false, message: "Terlalu banyak pengiriman. Coba lagi nanti." }, 429);
  }

  const res = await submit(form, ip);
  if (!res.ok) {
    return json({ success: false, message: res.error }, res.status);
  }

  return json({ success: true, tiket: res.ticket.tiket, data: res.ticket }, 201);
};

export const GET: APIRoute = async ({ url }) => {
  const tiket = url.searchParams.get("tiket") ?? "";

  const found = await track(tiket);
  if (!found) {
    return json({ success: false, message: "Nomor tiket tidak ditemukan." }, 404);
  }
  return json({ success: true, data: found });
};
