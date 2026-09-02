// Content use cases for the public site.
//
// Thin on purpose: these are ordered lists with no logic beyond fetching them.
// The layer exists so pages depend on a service rather than on a data source,
// which is what made swapping SQL for the API a change to one file.

export {
  findPerformanceDocs as getPerformanceDocs,
  findPPID as getPPID,
  findPublicApps as getPublicApps,
  findRegulations as getRegulations,
  findServiceLocations as getServiceLocations,
  findVideos as getVideos,
} from "../repositories/content.repository";

import { findPPID } from "../repositories/content.repository";

/**
 * PPID in the two flat lists the page renders.
 *
 * The API returns items nested inside their category, which is the better
 * shape, but flattening here keeps the change to the data source alone — the
 * template already iterates a category list and an item list, and rewriting
 * markup in the same commit as a data-access change makes both harder to review.
 */
export async function getPPIDLists() {
  const categories = await findPPID();

  const kategori = categories.map((c) => ({
    id: c.id,
    nama_kategori: c.nama_kategori,
    slug: c.slug,
  }));

  const items = categories.flatMap((c) =>
    (c.items ?? []).map((i) => ({
      id: i.id,
      kategori_id: c.id,
      judul: i.judul,
      isi: i.isi,
      file_url: i.file_url,
      nama_kategori: c.nama_kategori,
      slug: c.slug,
    }))
  );

  return { kategori, items };
}

export {
  findAbout as getAbout,
  findHome as getHome,
  findInfoSection as getInfoSection,
  findPhotos as getPhotos,
  findServiceStandards as getServiceStandards,
} from "../repositories/content.repository";

import { findHome } from "../repositories/content.repository";

/**
 * The homepage, shaped as the template already expects.
 *
 * Most sections need no adaptation: the API's field names are the legacy column
 * names, so notifikasi, slider, unggul, videos, aplikasi, tentang and renstra
 * pass straight through. Two do not:
 *
 *   berita  the API returns the article DTO (id, published_at, excerpt); the
 *           template reads the post columns (id_post, date, content).
 *   peta    the API returns one object or null, because there is only ever one
 *           row; the template indexes it as peta[0].
 */
export async function getHomeContent() {
  const home = await findHome();

  // A failed call renders an empty homepage rather than a 500. Every section
  // already tolerated an empty array — each had its own try/catch.
  if (!home) {
    return {
      notifikasi: [], tentang: [], regulasi: [], slider: [], berita: [],
      unggul: [], pdfData: [], videos: [], peta: [], renstra: [], aplikasi: [],
    };
  }

  const unggul = home.advantages ?? [];

  return {
    notifikasi: home.notifications ?? [],
    tentang: home.blocks ?? [],
    regulasi: home.regulations ?? [],
    slider: home.sliders ?? [],
    berita: (home.articles ?? []).map((a) => ({
      id_post: a.id,
      id_category: a.category_id,
      category_title: a.category ?? "",
      title: a.title,
      content: a.excerpt ?? "",
      picture: a.picture ?? "",
      date: a.published_at,
      hits: a.hits,
    })),
    unggul,
    pdfData: unggul.map((item) => ({ file: item.link, title: item.judul })),
    videos: home.videos ?? [],
    // Wrapped in an array: the template indexes peta[0].
    peta: home.investment_map ? [home.investment_map] : [],
    renstra: home.renstra ?? [],
    aplikasi: home.public_apps ?? [],
  };
}

import { findAbout, findInfoSection, findPhotos, findServiceStandards } from "../repositories/content.repository";

/**
 * About page content.
 *
 * The tab list is mapped back to the legacy mixed-case column names the
 * template reads (Nama_Konten, Isi_Konten…). The schema renamed them because
 * `TentangKamisWebV5` was queried in two different cases and only worked by
 * MySQL's accident of case folding; the markup still reads the old names, and
 * renaming data access and rewriting markup in one change makes both harder to
 * review.
 */
export async function getAboutContent() {
  const about = await findAbout();
  if (!about) {
    return { tentangdp: [], slider: [], tabs: [], dbError: true };
  }
  return {
    tentangdp: about.blocks ?? [],
    slider: about.sliders ?? [],
    tabs: (about.contents ?? []).map((c) => ({
      Id_Konten: c.id,
      Nama_Konten: c.nama,
      Isi_Konten: c.isi,
      Foto_Konten: c.foto,
      Judul_Konten: c.judul,
      Keterangan: c.keterangan,
      tags: c.tags,
    })),
    dbError: false,
  };
}

/** Service standards, with the JSON `isi` column decoded as the page expects. */
export async function getPelayananContent(sectionId = "info-sec") {
  const [standards, section] = await Promise.all([
    findServiceStandards(),
    findInfoSection(sectionId),
  ]);

  return {
    pelayananData: standards.map((row) => ({
      ...row,
      // The column holds a JSON document. A malformed one must not take the
      // page down, so it degrades to an empty object as the old code did.
      isi: parseJson(row.isi),
    })),
    info: section,
    infoItems: section?.items ?? [],
  };
}

function parseJson(value: string | undefined): unknown {
  try {
    return JSON.parse(value || "{}");
  } catch {
    return {};
  }
}

/** One page of gallery photos. */
export async function getPhotoPage(rawPage: unknown, perPage = 8) {
  const parsed = Number.parseInt(String(rawPage ?? "1"), 10);
  const page = Number.isFinite(parsed) && parsed > 0 ? parsed : 1;
  return findPhotos(page, perPage);
}
