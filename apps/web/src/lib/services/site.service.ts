// Site chrome use case.
//
// The menu-tree building that used to live here (and before that, inside
// Header.astro) is gone rather than ported: the Go API returns the menu already
// nested, because reshaping relational rows is not the frontend's job.

import type { SiteChrome } from "../models/site";
import { fetchChrome } from "../repositories/site.repository";

export async function getSiteChrome(): Promise<SiteChrome> {
  const { settings, navigationMenus, contactMenus } = await fetchChrome();
  return { settings, navigationMenus, contactMenus } as SiteChrome;
}
