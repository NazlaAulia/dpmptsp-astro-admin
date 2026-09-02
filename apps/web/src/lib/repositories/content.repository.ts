// Content data access — through the Go API.
//
// Every one of these was an inline db.execute in a page. They share a shape:
// fetch an ordered list, and on failure render an empty section rather than
// failing the whole page.

import { api } from "@dpmptsp/api-client";
import type {
  PerformanceDoc,
  PPIDCategory,
  PublicApp,
  Regulation,
  ServiceLocation,
  Video,
} from "@dpmptsp/api-client";

/**
 * Unwraps a list response, logging and returning [] on failure.
 *
 * The pages these replaced each wrapped their query in try/catch and fell back
 * to an empty array, which is the right behaviour: a missing regulations list
 * should render an empty section, not a 500 for the whole page.
 */
async function list<T>(
  name: string,
  call: () => Promise<{ ok: true; data: T[] } | { ok: false; status: number; error: string }>
): Promise<T[]> {
  const res = await call();
  if (!res.ok) {
    console.error(`${name} failed:`, res.status, res.error);
    return [];
  }
  return res.data;
}

export const findRegulations = () => list<Regulation>("regulations", () => api.regulations());
export const findPublicApps = () => list<PublicApp>("publicApps", () => api.publicApps());
export const findVideos = () => list<Video>("videos", () => api.videos());
export const findPerformanceDocs = () => list<PerformanceDoc>("performanceDocs", () => api.performanceDocs());
export const findServiceLocations = () => list<ServiceLocation>("serviceLocations", () => api.serviceLocations());
export const findPPID = () => list<PPIDCategory>("ppid", () => api.ppid());

// --- composite page payloads -------------------------------------------------

import type { About, Home, InfoSection, Photo, ServiceStandard } from "@dpmptsp/api-client";

/** The homepage, in one call instead of ten queries. */
export async function findHome(): Promise<Home | null> {
  const res = await api.home();
  if (!res.ok) {
    console.error("home failed:", res.status, res.error);
    return null;
  }
  return res.data;
}

export async function findAbout(): Promise<About | null> {
  const res = await api.about();
  if (!res.ok) {
    console.error("about failed:", res.status, res.error);
    return null;
  }
  return res.data;
}

export const findServiceStandards = () =>
  list<ServiceStandard>("serviceStandards", () => api.serviceStandards());

export async function findInfoSection(sectionId: string): Promise<InfoSection | null> {
  const res = await api.infoSection(sectionId);
  if (!res.ok) {
    // A 404 here just means the section is not configured yet.
    if (res.status !== 404) console.error("infoSection failed:", res.status, res.error);
    return null;
  }
  return res.data;
}

export async function findPhotos(page: number, perPage: number) {
  const res = await api.photos(page, perPage);
  if (!res.ok) {
    console.error("photos failed:", res.status, res.error);
    return { items: [] as Photo[], total: 0, page, perPage, totalPages: 1 };
  }
  return {
    items: res.data,
    total: res.meta?.total ?? res.data.length,
    page: res.meta?.page ?? page,
    perPage: res.meta?.per_page ?? perPage,
    totalPages: Math.max(1, res.meta?.pages ?? 1),
  };
}
