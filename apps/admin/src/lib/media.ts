// Where legacy media is served from.
import { optionalEnv } from "@dpmptsp/config";

export const MEDIA_BASE_URL = optionalEnv(
  "MEDIA_BASE_URL",
  "https://dpm-ptsp.surabaya.go.id"
).replace(/\/+$/, "");

/** Builds a media URL from a root-relative path. */
export function mediaUrl(path: string): string {
  return `${MEDIA_BASE_URL}/${String(path).replace(/^\/+/, "")}`;
}
