// Admin article data access, through the Go API.

import { currentSessionId } from "../request-context";
import { api, type Article } from "@dpmptsp/api-client";

export type ArticlePage = {
  items: Article[];
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
};

export async function findPage(page: number, perPage: number): Promise<ArticlePage> {
  // include_inactive: the admin table lists drafts as well as published rows.
  const res = await api.listArticles({ page, perPage, includeInactive: true });

  if (!res.ok) {
    console.error("listArticles failed:", res.status, res.error);
    return { items: [], total: 0, page, perPage, totalPages: 1 };
  }

  const meta = res.meta;
  return {
    items: res.data,
    total: meta?.total ?? res.data.length,
    page: meta?.page ?? page,
    perPage: meta?.per_page ?? perPage,
    totalPages: Math.max(1, meta?.pages ?? 1),
  };
}

export async function findById(id: number): Promise<Article | null> {
  const res = await api.articleById(id);
  if (res.ok) return res.data;
  if (res.status !== 404) console.error("articleById failed:", res.status, res.error);
  return null;
}

export async function create(body: Record<string, unknown>) {
  return api.createArticle(body, currentSessionId());
}

export async function update(id: number, body: Record<string, unknown>) {
  return api.updateArticle(id, body, currentSessionId());
}

export async function remove(id: number) {
  return api.deleteArticle(id, currentSessionId());
}
