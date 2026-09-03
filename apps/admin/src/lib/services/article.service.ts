// Admin article use cases.

import { pageWindow } from "@dpmptsp/ui";
import type { ArticlePage } from "../repositories/article.repository";
import { create, findById, findPage, remove, update } from "../repositories/article.repository";

export const DEFAULT_PER_PAGE = 10;

export type ArticleListing = ArticlePage & {
  window: ReturnType<typeof pageWindow>;
};

/** One page of articles, with the page number clamped rather than trusted. */
export async function listArticles(
  rawPage: unknown,
  perPage: number = DEFAULT_PER_PAGE
): Promise<ArticleListing> {
  const parsed = Number.parseInt(String(rawPage ?? "1"), 10);
  const page = Number.isFinite(parsed) && parsed > 0 ? parsed : 1;

  const result = await findPage(page, perPage);
  return {
    ...result,
    window: pageWindow(result.page, result.totalPages, result.total, result.perPage),
  };
}

export async function getArticle(id: number) {
  return findById(id);
}

export async function createArticle(body: Record<string, unknown>) {
  const res = await create(body);
  return res.ok ? { ok: true as const, data: res.data } : { ok: false as const, message: message(res) };
}

export async function updateArticle(id: number, body: Record<string, unknown>) {
  const res = await update(id, body);
  return res.ok ? { ok: true as const, data: res.data } : { ok: false as const, message: message(res) };
}

export async function deleteArticle(id: number) {
  const res = await remove(id);
  return res.ok ? { ok: true as const } : { ok: false as const, message: message(res) };
}

function message(res: { status: number; error: string }): string {
  console.error("article write failed:", res.status, res.error);
  if (res.status === 404) return "Artikel tidak ditemukan.";
  if (res.status === 422) return res.error;
  return "Terjadi kesalahan pada server.";
}
