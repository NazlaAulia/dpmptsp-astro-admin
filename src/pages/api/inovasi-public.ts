import type { APIRoute } from "astro";
import { db } from "../../lib/db.js"; // Sesuaikan path db kamu

export const prerender = false;

export const GET: APIRoute = async () => {
  try {
    const [rows]: any = await db.query(
      "SELECT * FROM inovasi_layanan WHERE status = 1 ORDER BY urutan ASC"
    );

    const data = rows.map((item: any) => ({
      id: item.id,
      slug: item.slug,
      nama: item.nama,
      singkatan: item.singkatan || "",
      kategori: item.kategori || "Umum",
      deskripsi: item.deskripsi || "Informasi belum tersedia.",
      tujuan: item.tujuan || "Informasi belum tersedia.",
      manfaat: item.manfaat || "Informasi belum tersedia.",
      hasil: item.hasil || "Informasi belum tersedia.",
      tahun: item.tahun_usulan || "2024",
      jenis: item.jenis || "Digital",
      gambar: item.gambar ? `/${item.gambar}` : null, // Path gambar dibenerin di sini
      url: item.url_layanan || null,
      urlLabel: item.url_label || "Kunjungi Layanan",
      icon: item.icon || "fa-lightbulb",
      warna: item.warna || "blue",
    }));

    return new Response(JSON.stringify(data), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  } catch (error: any) {
    return new Response(JSON.stringify({ error: error.message }), { status: 500 });
  }
};