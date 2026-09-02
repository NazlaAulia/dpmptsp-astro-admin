// Article use cases for the admin panel.

import type { ArticlePage } from "../models/article";
import { pageWindow } from "@dpmptsp/ui";
import { countAll, findPage } from "../repositories/article.repository";

export const DEFAULT_PER_PAGE = 10;

/**
 * One page of articles, with the pagination arithmetic done once here instead
 * of being re-derived in every table component.
 */
export async function listArticles(
  page: number,
  perPage: number = DEFAULT_PER_PAGE
): Promise<ArticlePage> {
  const total = await countAll();
  const totalPages = Math.max(1, Math.ceil(total / perPage));
  // Clamp rather than trust the query string: ?page=0 or ?page=abc otherwise
  // produces a negative OFFSET and a SQL error.
  const current = Math.min(Math.max(Number.isFinite(page) ? page : 1, 1), totalPages);
  const items = await findPage(perPage, (current - 1) * perPage);

  return {
    items,
    total,
    page: current,
    perPage,
    totalPages,
    window: pageWindow(current, totalPages, total, perPage),
  };
}
