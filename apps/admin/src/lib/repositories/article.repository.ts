// Data access for articles.

import { db } from "../db.js";
import type { Article } from "../models/article";

export async function countAll(): Promise<number> {
  const [rows] = await db.query("SELECT COUNT(*) AS total FROM post");
  return Number((rows as { total: number }[])?.[0]?.total ?? 0);
}

export async function findPage(limit: number, offset: number): Promise<Article[]> {
  // LIMIT/OFFSET are bound, not interpolated. Several pages in this codebase
  // splice them into the SQL string; that pattern must not spread.
  const [rows] = await db.query(
    `SELECT post.*, category_berita.title AS kategori
       FROM post
       LEFT JOIN category_berita ON post.id_category = category_berita.id_category
      ORDER BY post.id_post DESC
      LIMIT ? OFFSET ?`,
    [limit, offset]
  );
  return (rows as Article[]) ?? [];
}

export async function deleteById(id: number): Promise<void> {
  await db.query("DELETE FROM post WHERE id_post = ?", [id]);
}
