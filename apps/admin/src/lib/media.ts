// Where legacy media is served from.
//
// The article images, the "tentang" photos and the logo all still live on the
// old PHP deployment at dpm-ptsp.surabaya.go.id, and its hostname was hardcoded
// in about a dozen templates. This is the seam that makes moving them a
// configuration change: point MEDIA_BASE_URL at MinIO, or at a CDN, and every
// reference follows.
//
// The default preserves today's behaviour exactly, so nothing changes until it
// is set.
//
// Note this covers MEDIA only. Links to the legacy *site* — its homepage, the
// interactive potential-investment map — are deliberate external links to
// another system and are left alone.
import { optionalEnv } from "@dpmptsp/config";

export const MEDIA_BASE_URL = optionalEnv(
  "MEDIA_BASE_URL",
  "https://dpm-ptsp.surabaya.go.id"
).replace(/\/+$/, "");

/** Builds a media URL from a root-relative path. */
export function mediaUrl(path: string): string {
  return `${MEDIA_BASE_URL}/${String(path).replace(/^\/+/, "")}`;
}
