export const prerender = false;
import { db } from "../../lib/db.js";

export async function GET({ url }) {
  const q = url.searchParams.get("q") || "";
  const type = url.searchParams.get("type") || "berita";

  if (q.length < 2) {
    return new Response(JSON.stringify([]), { status: 200 });
  }

  const category =
    type === "berita" ? "(1,8)" : "(2,3)";

  const [rows] = await db.execute(
    `
    SELECT id_post, title, date, picture, id_category
    FROM post
    WHERE id_category IN ${category}
    AND title LIKE ?
    ORDER BY date DESC
    LIMIT 10
    `,
    [`%${q}%`]
  );

  return new Response(JSON.stringify(rows), {
    headers: { "Content-Type": "application/json" }
  });
}
