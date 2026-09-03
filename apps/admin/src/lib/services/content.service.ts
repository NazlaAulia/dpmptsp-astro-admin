// Admin content use cases.

export * from "../repositories/content.repository";
export { uploadFile, hasFile } from "../repositories/upload.repository";

import type { AboutContent } from "@dpmptsp/api-client";
import { findAboutContent, findAboutContents } from "../repositories/content.repository";

/**
 * The konten screens render the legacy mixed-case column names. Mapping here
 * keeps the change to data access alone rather than also rewriting markup.
 */
export type LegacyKonten = {
  Id_Konten: number;
  Nama_Konten: string;
  Isi_Konten: string;
  Foto_Konten: string;
  Judul_Konten: string;
  Keterangan: string;
  tags: string;
};

function toLegacy(c: AboutContent): LegacyKonten {
  return {
    Id_Konten: c.id ?? 0,
    Nama_Konten: c.nama ?? "",
    Isi_Konten: c.isi ?? "",
    Foto_Konten: c.foto ?? "",
    Judul_Konten: c.judul ?? "",
    Keterangan: c.keterangan ?? "",
    tags: c.tags ?? "",
  };
}

export async function getKontenList(): Promise<LegacyKonten[]> {
  return (await findAboutContents()).map(toLegacy);
}

export async function getKonten(id: number): Promise<LegacyKonten | null> {
  const res = await findAboutContent(id);
  return res.ok ? toLegacy(res.data) : null;
}
