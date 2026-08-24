import { db } from "../../../lib/db.js";

export async function POST({ request }) {
  const url = new URL(request.url);
  const id = url.searchParams.get("id");

  console.log("ID yang mau dihapus:", id); // Cek terminal server Astro

  if (id) {
    try {
      const [result] = await db.execute("DELETE FROM tentangkamiswebv5 WHERE Id_Konten = ?", [id]);
      console.log("Hasil hapus database:", result);
    } catch (error) {
      console.error("GAGAL HAPUS DI DATABASE:", error); // Cek apakah ada error SQL
    }
  }

  return new Response(JSON.stringify({ success: true }), {
    status: 200,
    headers: { "Content-Type": "application/json" }
  });
}