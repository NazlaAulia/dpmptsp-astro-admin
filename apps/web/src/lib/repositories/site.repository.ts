// Site chrome data access — now through the Go API rather than SQL.
//
// This file is the seam described in the article repository, and this is what
// it was for: the service above and Header.astro above that are unchanged. Only
// the bodies here moved from db.query to an API call.

import { api } from "@dpmptsp/api-client";
import { FALLBACK_SETTINGS, type MenuRow, type SiteSettings } from "../models/site";

export type ChromeData = {
  settings: SiteSettings;
  navigationMenus: any[];
  contactMenus: any[];
};

/**
 * One call rather than two queries.
 *
 * The menu arrives already nested and sorted: building the tree from flat
 * parent_id rows is data work, and it used to happen inside a component on
 * every render of every page.
 */
export async function fetchChrome(): Promise<ChromeData> {
  const res = await api.siteChrome();

  // The header is on every page. If the API is unreachable the site still
  // renders with its name and no navigation, rather than returning a 500.
  if (!res.ok) {
    console.error("siteChrome failed:", res.status, res.error);
    return { settings: FALLBACK_SETTINGS, navigationMenus: [], contactMenus: [] };
  }

  const { settings, navigation, contact } = res.data;
  return {
    settings: { nama: settings.name || FALLBACK_SETTINGS.nama, logo: settings.logo || FALLBACK_SETTINGS.logo },
    navigationMenus: navigation.map(toLegacyShape),
    contactMenus: contact.map(toLegacyShape),
  };
}

/** Maps the API shape onto the field names Header.astro already renders. */
function toLegacyShape(n: any): any {
  return {
    id: n.id,
    nama: n.name,
    url: n.url,
    tipe: n.type ?? null,
    tombol_kontak: n.contact_button ? 1 : 0,
    children: (n.children ?? []).map(toLegacyShape),
  };
}
