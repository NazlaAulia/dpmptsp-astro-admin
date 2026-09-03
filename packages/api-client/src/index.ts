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
export type SessionInfo = Schemas["Session"];
export type Innovation = Schemas["Innovation"];
export type AboutContent = Schemas["AboutContent"];
export type Upload = Schemas["Upload"];
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
export type ContactMessageInput = Schemas["ContactMessageInput"];
export type ContactTicket = Schemas["ContactTicket"];
export type ContactStatus = Schemas["ContactStatus"];
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
  init: RequestInit & { timeoutMs?: number; sessionId?: string } = {}
): Promise<ApiResult<T>> {
  const { timeoutMs = intEnv("API_TIMEOUT_MS", DEFAULT_TIMEOUT_MS), sessionId, ...rest } = init;

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
        // Identifies the signed-in administrator. The API resolves it against
        // its session store and authorizes writes on it; the service key alone
        // only proves the caller is one of our own apps.
        ...(sessionId ? { "X-Session-Id": sessionId } : {}),
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

/** JSON-bodied request helper. */
function send<T>(method: string, path: string, body?: unknown, sessionId?: string) {
  return request<T>(path, {
    method,
    sessionId,
    headers: body === undefined ? {} : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
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
    return request<SessionInfo>("/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });
  },

  // --- innovations ---
  innovations() {
    return request<Innovation[]>("/v1/innovations");
  },
  innovation(id: number) {
    return request<Innovation>(`/v1/innovations/${id}`);
  },
  createInnovation(body: Partial<Innovation>, sessionId?: string) {
    return send<Innovation>("POST", "/v1/innovations", body, sessionId);
  },
  updateInnovation(id: number, body: Partial<Innovation>, sessionId?: string) {
    return send<Innovation>("PUT", `/v1/innovations/${id}`, body, sessionId);
  },
  deleteInnovation(id: number, sessionId?: string) {
    return send<null>("DELETE", `/v1/innovations/${id}`, undefined, sessionId);
  },

  // --- performance docs ---
  performanceDoc(id: number) {
    return request<PerformanceDoc>(`/v1/performance-docs/${id}`);
  },
  createPerformanceDoc(body: Partial<PerformanceDoc>, sessionId?: string) {
    return send<PerformanceDoc>("POST", "/v1/performance-docs", body, sessionId);
  },
  updatePerformanceDoc(id: number, body: Partial<PerformanceDoc>, sessionId?: string) {
    return send<PerformanceDoc>("PUT", `/v1/performance-docs/${id}`, body, sessionId);
  },
  deletePerformanceDoc(id: number, sessionId?: string) {
    return send<null>("DELETE", `/v1/performance-docs/${id}`, undefined, sessionId);
  },

  // --- service locations ---
  serviceLocation(id: number) {
    return request<ServiceLocation>(`/v1/service-locations/${id}`);
  },
  createServiceLocation(body: Partial<ServiceLocation>, sessionId?: string) {
    return send<ServiceLocation>("POST", "/v1/service-locations", body, sessionId);
  },
  updateServiceLocation(id: number, body: Partial<ServiceLocation>, sessionId?: string) {
    return send<ServiceLocation>("PUT", `/v1/service-locations/${id}`, body, sessionId);
  },
  deleteServiceLocation(id: number, sessionId?: string) {
    return send<null>("DELETE", `/v1/service-locations/${id}`, undefined, sessionId);
  },

  // --- about contents ---
  aboutContents() {
    return request<AboutContent[]>("/v1/about-contents");
  },
  aboutContent(id: number) {
    return request<AboutContent>(`/v1/about-contents/${id}`);
  },
  createAboutContent(body: Partial<AboutContent>, sessionId?: string) {
    return send<AboutContent>("POST", "/v1/about-contents", body, sessionId);
  },
  updateAboutContent(id: number, body: Partial<AboutContent>, sessionId?: string) {
    return send<AboutContent>("PUT", `/v1/about-contents/${id}`, body, sessionId);
  },
  deleteAboutContent(id: number, sessionId?: string) {
    return send<null>("DELETE", `/v1/about-contents/${id}`, undefined, sessionId);
  },

  // --- articles (admin writes) ---
  createArticle(body: Record<string, unknown>, sessionId?: string) {
    return send<Article>("POST", "/v1/articles", body, sessionId);
  },
  updateArticle(id: number, body: Record<string, unknown>, sessionId?: string) {
    return send<Article>("PUT", `/v1/articles/${id}`, body, sessionId);
  },
  deleteArticle(id: number, sessionId?: string) {
    return send<null>("DELETE", `/v1/articles/${id}`, undefined, sessionId);
  },

  /**
   * Stores a file and returns the key to reference it by.
   *
   * The stored name is generated by the API; the client's filename only
   * influences the extension when the sniffed type is unrecognised.
   */
  upload(file: File | Blob, opts: { prefix?: string; visibility?: "public" | "private"; filename?: string; sessionId?: string } = {}) {
    const form = new FormData();
    form.append("file", file, opts.filename ?? "upload");
    if (opts.prefix) form.append("prefix", opts.prefix);
    if (opts.visibility) form.append("visibility", opts.visibility);
    // No Content-Type header: fetch sets the multipart boundary itself.
    return request<Upload>("/v1/uploads", {
      method: "POST", body: form, timeoutMs: 60_000, sessionId: opts.sessionId,
    });
  },

  currentSession(sessionId: string) {
    return request<SessionInfo>("/v1/auth/session", { sessionId });
  },

  logout(sessionId: string) {
    return send<null>("DELETE", "/v1/auth/session", undefined, sessionId);
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

  createAnnouncement(body: AnnouncementInput, sessionId?: string) {
    return send<Announcement>("POST", "/v1/announcements", body, sessionId);
  },

  updateAnnouncement(id: number, body: AnnouncementInput, sessionId?: string) {
    return send<Announcement>("PUT", `/v1/announcements/${id}`, body, sessionId);
  },

  deleteAnnouncement(id: number, sessionId?: string) {
    return send<null>("DELETE", `/v1/announcements/${id}`, undefined, sessionId);
  },

  photos(page = 1, perPage = 8) {
    return request<Photo[]>(`/v1/gallery/photos?page=${page}&per_page=${perPage}`);
  },

  /**
   * Submit the public contact form. The caller's address is forwarded so the
   * API can rate limit on it; without it every submission looks like it came
   * from Astro.
   */
  submitContact(body: ContactMessageInput, clientIp?: string) {
    return request<ContactTicket>("/v1/contact-messages", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(clientIp ? { "X-Forwarded-For": clientIp } : {}),
      },
      body: JSON.stringify(body),
    });
  },

  trackContact(tiket: string) {
    return request<ContactStatus>(`/v1/contact-messages/track?tiket=${encodeURIComponent(tiket)}`);
  },
};
