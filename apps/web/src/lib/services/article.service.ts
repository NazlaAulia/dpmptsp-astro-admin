// Article use cases for the public site.

import { CATEGORY_ARTIKEL, CATEGORY_BERITA, type ArticlePage } from "../models/article";
import { findArticleBySlug, findArticles, findRelated } from "../repositories/article.repository";
import { CATEGORY_ARTIKEL as ARTIKEL_IDS, CATEGORY_BERITA as BERITA_IDS } from "../models/article";

export const PER_PAGE = 8;

export type ListType = "all" | "berita" | "artikel";

function categoriesFor(type: ListType): number[] | undefined {
  if (type === "berita") return CATEGORY_BERITA;
  if (type === "artikel") return CATEGORY_ARTIKEL;
  return undefined;
}

/**
 * One page of the article list.
 *
 * The page number is clamped here rather than trusted. The previous version
 * interpolated an offset derived from the query string straight into SQL, so
 * ?page=abc produced `OFFSET NaN` and a 500.
 */
export async function listArticles(
  rawPage: unknown,
  type: ListType,
  search: string
): Promise<ArticlePage> {
  const parsed = Number.parseInt(String(rawPage ?? "1"), 10);
  const page = Number.isFinite(parsed) && parsed > 0 ? parsed : 1;

  const first = await findArticles({
    page,
    perPage: PER_PAGE,
    search: search || undefined,
    categoryIds: categoriesFor(type),
  });

  // Asking for a page past the end returns nothing at all; show the last real
  // page instead, which is what the old code did after its count query.
  if (page > first.totalPages && first.total > 0) {
    return findArticles({
      page: first.totalPages,
      perPage: PER_PAGE,
      search: search || undefined,
      categoryIds: categoriesFor(type),
    });
  }
  return first;
}

/**
 * Everything the article page needs, fetched concurrently.
 *
 * The page it replaces ran five queries, two of whose results it discarded, and
 * two of which were `SELECT *` over the whole post table — one unbounded, one
 * `LIMIT 1000` — purely to find one article by slug and pick five siblings.
 */
export async function getArticlePage(rawSlug: string) {
  const slug = decodeURIComponent(String(rawSlug ?? "")).toLowerCase().trim();

  const article = await findArticleBySlug(slug);
  if (!article) return null;

  const [berita, artikel] = await Promise.all([
    findRelated(BERITA_IDS, 5),
    findRelated(ARTIKEL_IDS, 5),
  ]);

  return {
    article,
    categoryName: article.category ?? "Tanpa Kategori",
    // Do not show the article you are already reading in its own sidebar.
    berita: berita.filter((b) => b.id !== article.id).slice(0, 5),
    artikel: artikel.filter((b) => b.id !== article.id).slice(0, 5),
  };
}
