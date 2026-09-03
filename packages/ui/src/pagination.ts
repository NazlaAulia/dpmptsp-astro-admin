// Pagination arithmetic.

export type PageWindow = {
  /** Page numbers to render, capped to five. */
  visiblePages: number[];
  showingStart: number;
  showingEnd: number;
};

export function pageWindow(
  page: number,
  totalPages: number,
  total: number,
  perPage: number
): PageWindow {
  const offset = (page - 1) * perPage;
  let start = Math.max(1, page - 2);
  const end = Math.min(totalPages, start + 4);
  if (end - start < 4) start = Math.max(1, end - 4);

  return {
    visiblePages: Array.from({ length: end - start + 1 }, (_, i) => start + i),
    showingStart: total === 0 ? 0 : offset + 1,
    // Clamp to the real total: the last page must not claim more rows than exist.
    showingEnd: Math.min(offset + perPage, total),
  };
}

/**
 * Builds a page link that preserves the other query parameters.
 */
export function pageHref(params: URLSearchParams, page: number): string {
  const next = new URLSearchParams(params);
  next.set("page", String(page));
  return `?${next.toString()}`;
}
