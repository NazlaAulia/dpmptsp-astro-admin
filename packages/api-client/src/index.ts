// Client for the Go API.

import { optionalEnv, requireEnv, intEnv } from "@dpmptsp/config";
import type { components } from "./generated/schema";

// Request and response shapes come from the OpenAPI spec, never hand-written
// . Regenerate with `pnpm --filter
// @dpmptsp/api-client generate` after changing apps/api/openapi.yaml.
type Schemas = components["schemas"];

export type ApiMeta = Schemas["PageMeta"];

export type ApiResult<T> =
  | { ok: true; data: T; meta?: ApiMeta }
  | { ok: false; status: number; error: string };

export type Article = Schemas["Article"];
export type ArticleInput = Schemas["ArticleInput"];
export type Category = Schemas["Category"];
export type MenuNode = Schemas["MenuNode"];
export type SiteChrome = Schemas["SiteChrome"];
export type User = Schemas["User"];
export type Regulation = Schemas["Regulation"];
export type PublicApp = Schemas["PublicApp"];
export type Video = Schemas["Video"];
export type PerformanceDoc = Schemas["PerformanceDoc"];
export type ServiceLocation = Schemas["ServiceLocation"];
export type PPIDCategory = Schemas["PPIDCategory"];
export type Home = Schemas["Home"];
export type About = Schemas["About"];
export type ServiceStandard = Schemas["ServiceStandard"];
export type InfoSection = Schemas["InfoSection"];
export type Photo = Schemas["Photo"];
export type Announcement = Schemas["Announcement"];
export type AnnouncementInput = Schemas["AnnouncementInput"];

export type ListArticlesQuery = {
  page?: number;
  perPage?: number;
  search?: string;
  categoryIds?: number[];
  headlineOnly?: boolean;
  includeInactive?: boolean;
};

function baseUrl(): string {
  return optionalEnv("API_BASE_URL", "http://api:8080").replace(/\/+$/, "");
}

/**
 * Every request is bounded.
 */
const DEFAULT_TIMEOUT_MS = 5_000;

async function request<T>(
  path: string,
  init: RequestInit & { timeoutMs?: number } = {}
): Promise<ApiResult<T>> {
  const { timeoutMs = intEnv("API_TIMEOUT_MS", DEFAULT_TIMEOUT_MS), ...rest } = init;

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const res = await fetch(`${baseUrl()}${path}`, {
      ...rest,
      signal: controller.signal,
      headers: {
        Accept: "application/json",
        // Authenticates this process as one of our own Astro apps.
        "X-Internal-Key": requireEnv("API_SERVICE_KEY"),
        ...(rest.headers ?? {}),
      },
    });

    const text = await res.text();
    const body = text ? JSON.parse(text) : {};

    if (!res.ok) {
      return { ok: false, status: res.status, error: body?.error ?? res.statusText };
    }
    return { ok: true, data: body.data as T, meta: body.meta };
  } catch (err) {
    // Never throw. An uncaught throw in Astro frontmatter turns one failed
    // dependency into a 500 for the whole page; returning a result lets the
    // caller decide whether that section is essential.
    const aborted = err instanceof Error && err.name === "AbortError";
    return {
      ok: false,
      status: aborted ? 504 : 502,
      error: aborted ? `timed out after ${timeoutMs}ms` : String(err),
    };
  } finally {
    clearTimeout(timer);
  }
}

function qs(params: Record<string, string | number | boolean | undefined>): string {
  const out = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== "") out.set(k, String(v));
  }
  const s = out.toString();
  return s ? `?${s}` : "";
}

export const api = {
  listArticles(q: ListArticlesQuery = {}) {
    return request<Article[]>(
      "/v1/articles" +
        qs({
          page: q.page,
          per_page: q.perPage,
          q: q.search,
          category: q.categoryIds?.join(","),
          headline: q.headlineOnly ? "true" : undefined,
          include_inactive: q.includeInactive ? "true" : undefined,
        })
    );
  },

  articleBySlug(slug: string, countView = false) {
    return request<Article>(
      `/v1/articles/by-slug/${encodeURIComponent(slug)}` +
        qs({ count_view: countView ? "true" : undefined })
    );
  },

  articleById(id: number) {
    return request<Article>(`/v1/articles/${id}`);
  },

  categories() {
    return request<Category[]>("/v1/categories");
  },

  login(username: string, password: string) {
    return request<User>("/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });
  },

  siteChrome() {
    return request<SiteChrome>("/v1/site/chrome");
  },

  regulations() {
    return request<Regulation[]>("/v1/regulations");
  },

  publicApps() {
    return request<PublicApp[]>("/v1/public-apps");
  },

  videos() {
    return request<Video[]>("/v1/videos");
  },

  performanceDocs() {
    return request<PerformanceDoc[]>("/v1/performance-docs");
  },

  serviceLocations() {
    return request<ServiceLocation[]>("/v1/service-locations");
  },

  ppid() {
    return request<PPIDCategory[]>("/v1/ppid");
  },

  home() {
    // The homepage is the heaviest payload; give it more room than the default.
    return request<Home>("/v1/home", { timeoutMs: 8_000 });
  },

  about() {
    return request<About>("/v1/about");
  },

  serviceStandards() {
    return request<ServiceStandard[]>("/v1/service-standards");
  },

  infoSection(sectionId: string) {
    return request<InfoSection>(`/v1/info-sections/${encodeURIComponent(sectionId)}`);
  },

  announcements(opts: { tipe?: "notif" | "modal"; includeInactive?: boolean } = {}) {
    return request<Announcement[]>(
      "/v1/announcements" +
        qs({
          tipe: opts.tipe,
          include_inactive: opts.includeInactive ? "true" : undefined,
        })
    );
  },

  announcement(id: number) {
    return request<Announcement>(`/v1/announcements/${id}`);
  },

  createAnnouncement(body: AnnouncementInput) {
    return request<Announcement>("/v1/announcements", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
  },

  updateAnnouncement(id: number, body: AnnouncementInput) {
    return request<Announcement>(`/v1/announcements/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
  },

  deleteAnnouncement(id: number) {
    return request<null>(`/v1/announcements/${id}`, { method: "DELETE" });
  },

  photos(page = 1, perPage = 8) {
    return request<Photo[]>(`/v1/gallery/photos?page=${page}&per_page=${perPage}`);
  },
};
