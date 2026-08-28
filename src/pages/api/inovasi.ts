import type { APIRoute } from 'astro';
// ⚠️ GANTI baris di bawah ini sesuai file koneksi DB kamu (misal koneksi mysql2)
import { db } from '../../lib/db'; 

export const POST: APIRoute = async ({ request, redirect }) => {
  try {
    const formData = await request.formData();

    // 1. Ambil input dari form
    const nama = formData.get("nama")?.toString() || "";
    const singkatan = formData.get("singkatan")?.toString() || "";
    const kategori = formData.get("kategori")?.toString() || "";
    const deskripsi = formData.get("deskripsi")?.toString() || "";
    const rancang_bangun = formData.get("rancang_bangun")?.toString() || "";
    const tujuan = formData.get("tujuan")?.toString() || "";
    const manfaat = formData.get("manfaat")?.toString() || "";
    const hasil = formData.get("hasil")?.toString() || "";
    const tahun_usulan = formData.get("tahun_usulan")?.toString() || new Date().getFullYear().toString();
    const tahapan = formData.get("tahapan")?.toString() || "Penerapan Digital";
    const jenis = formData.get("jenis")?.toString() || "Digital";

    // Format slug otomatis (misal: "TAKON SOBAT" -> "takon-sobat")
    const slug = nama.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)+/g, "");

    // 2. Handle Gambar jika ada
    const fileGambar = formData.get("gambar") as File;
    let gambarPath = "";
    if (fileGambar && fileGambar.name && fileGambar.size > 0) {
      gambarPath = `gambar/${fileGambar.name}`;
    }

    // 3. Query eksekusi langsung ke MySQL
    await db.query(
      `INSERT INTO inovasi_layanan 
      (slug, nama, singkatan, kategori, deskripsi, rancang_bangun, tujuan, manfaat, hasil, tahun_usulan, tahapan, jenis, gambar) 
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [slug, nama, singkatan, kategori, deskripsi, rancang_bangun, tujuan, manfaat, hasil, tahun_usulan, tahapan, jenis, gambarPath]
    );

    // Redirect balik ke halaman tabel setelah berhasil insert
    return redirect("/admin/inovasi");

  } catch (error) {
    console.error("Gagal simpan ke DB:", error);
    return new Response("Gagal simpan ke database", { status: 500 });
  }
};