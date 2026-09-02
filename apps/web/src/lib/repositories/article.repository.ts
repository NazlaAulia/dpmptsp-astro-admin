// Article data access — now through the Go API rather than SQL.
//
// This file is the seam. It previously would have held db.query calls; the
// service and the pages above it are unchanged by the switch, which is the
// whole reason the layer exists.

import { api, type Article } from "@dpmptsp/api-client";
import type { ArticleDetail, ArticleListItem, ArticlePage } from "../models/article";

/** Maps the API's wire shape onto the field names the templates already read. */
function toListItem(a: Article): ArticleListItem {
  return {
    id: a.id,
    category_id: a.category_id,
    category: a.category,
    slug: a.slug,
    title: a.title,
    content: a.excerpt ?? a.content ?? "",
    picture: a.picture ?? "",
    date: a.published_at,
    title_en: null,
  };
}

export async function findArticles(opts: {
  page: number;
  perPage: number;
  search?: string;
  categoryIds?: number[];
}): Promise<ArticlePage> {
  const res = await api.listArticles({
    page: opts.page,
    perPage: opts.perPage,
    search: opts.search,
    categoryIds: opts.categoryIds,
  });

  // A failed API call renders an empty list rather than a 500. The page still
  // shows its chrome and its search box; one broken dependency should not take
  // the whole page down.
  if (!res.ok) {
    console.error("listArticles failed:", res.status, res.error);
    return { items: [], total: 0, page: opts.page, perPage: opts.perPage, totalPages: 1 };
  }

  const meta = res.meta;
  return {
    items: res.data.map(toListItem),
    total: meta?.total ?? res.data.length,
    page: meta?.page ?? opts.page,
    perPage: meta?.per_page ?? opts.perPage,
    totalPages: Math.max(1, meta?.pages ?? 1),
  };
}

export async function findArticleBySlug(slug: string): Promise<ArticleDetail | null> {
  // count_view is on: this is the article page, and a view is what it counts.
  const res = await api.articleBySlug(slug, true);
  if (!res.ok) {
    if (res.status !== 404) console.error("articleBySlug failed:", res.status, res.error);
    return null;
  }
  const a = res.data;
  return { ...toListItem(a), content: a.content ?? "", ref_content: a.ref_content ?? "", hits: a.hits };
}

export async function findRelated(categoryIds: number[], limit: number): Promise<ArticleListItem[]> {
  const res = await api.listArticles({ categoryIds, perPage: limit, page: 1 });
  return res.ok ? res.data.map(toListItem) : [];
}
