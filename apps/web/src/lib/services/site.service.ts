// Site chrome use case.

import type { SiteChrome } from "../models/site";
import { fetchChrome } from "../repositories/site.repository";

export async function getSiteChrome(): Promise<SiteChrome> {
  const { settings, navigationMenus, contactMenus } = await fetchChrome();
  return { settings, navigationMenus, contactMenus } as SiteChrome;
}
