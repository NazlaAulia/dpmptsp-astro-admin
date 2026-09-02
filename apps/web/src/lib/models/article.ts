// Article shapes for the public site.
//
// Field names follow the legacy `post` columns because the templates already
// read them that way. Renaming them here would mean editing markup as well as
// data access in the same change, and those are much easier to review apart.

export type ArticleListItem = {
  id: number;
  category_id: number;
  category?: string;
  slug: string;
  title: string;
  /** Tag-stripped excerpt. Never the full body — a list of 8 must not ship 8 HTML documents. */
  content: string;
  title_en?: string | null;
  picture: string;
  date: string;
};

export type ArticleDetail = ArticleListItem & {
  /** Rendered body. The article page displays this rather than `content`. */
  ref_content: string;
  hits: number;
};

export type ArticlePage = {
  items: ArticleListItem[];
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
};

/** Category ids the site groups articles by. Hardcoded in the legacy pages. */
export const CATEGORY_BERITA = [1, 8];
export const CATEGORY_ARTIKEL = [2, 3];
