// File uploads, through the Go API.
//
// Replaces fs.writeFile into public/: the API owns storage, so uploads survive
// more than one replica and the web process no longer needs write access to the
// directory it serves.

import { currentSessionId } from "../request-context";
import { api } from "@dpmptsp/api-client";

export type UploadResult =
  | { ok: true; key: string; url: string }
  | { ok: false; message: string };

export async function uploadFile(
  file: File,
  prefix: string,
  visibility: "public" | "private" = "public"
): Promise<UploadResult> {
  const res = await api.upload(file, {
    prefix, visibility, filename: file.name, sessionId: currentSessionId(),
  });
  if (res.ok) return { ok: true, key: res.data.key, url: res.data.url ?? res.data.key };

  if (res.status === 415) return { ok: false, message: "Format file tidak didukung." };
  if (res.status === 413) return { ok: false, message: "Ukuran file terlalu besar." };
  console.error("upload failed:", res.status, res.error);
  return { ok: false, message: "Gagal mengunggah berkas." };
}

/** True when a form file part actually carries a file. */
export function hasFile(value: FormDataEntryValue | null): value is File {
  return value instanceof File && value.size > 0;
}
