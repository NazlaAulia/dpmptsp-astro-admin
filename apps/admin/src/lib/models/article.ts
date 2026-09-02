// Article (legacy table `post`) shapes.

export type Article = {
  id_post: number;
  id_category: number;
  title: string;
  content: string;
  date: string;
  time: string;
  active: "Y" | "N";
  headline: "Y" | "N";
  picture: string;
  hits: number;
  /** Joined from category_berita. Aliased `kategori`, as the table expects. */
  kategori?: string | null;
};

import type { PageWindow } from "@dpmptsp/ui";

export type ArticlePage = {
  items: Article[];
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
  window: PageWindow;
};
