// Shapes for site-wide chrome: branding and the header menu tree.
//
// Models describe data, nothing else — no fetching, no SQL, no framework
// imports. They are the contract the repository fills and the service returns,
// which is what lets the repository swap from SQL to the Go API without pages
// noticing.

export type SiteSettings = {
  nama: string;
  logo: string;
};

/** One row of header_menu, as stored. */
export type MenuRow = {
  id: number;
  parent_id: number | null;
  nama: string;
  url: string | null;
  urutan: number;
  aktif: number;
  tipe: string | null;
  tombol_kontak: number;
};

/** A menu row after the flat parent_id list has been nested. */
export type MenuNode = Omit<MenuRow, "parent_id"> & {
  parent_id: number | null;
  children: MenuNode[];
};

/** Everything the header and footer need, in one object. */
export type SiteChrome = {
  settings: SiteSettings;
  navigationMenus: MenuNode[];
  contactMenus: MenuNode[];
};

export const FALLBACK_SETTINGS: SiteSettings = {
  nama: "DPM-PTSP Surabaya",
  logo: "https://github.com/arbagasdanangh/pps/blob/main/logo.png?raw=true",
};
