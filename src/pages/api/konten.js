import { db } from "../../lib/db.js";
import fs from "node:fs/promises";
import path from "node:path";

export async function POST({ request }) {
  try {
    const formData = await request.formData();

    const namaKonten = formData.get("nama_konten")?.toString().trim();
    const judulKonten = formData.get("judul_konten")?.toString().trim();
    const isiKonten = formData.get("isi_konten")?.toString().trim();
    const keterangan = formData.get("keterangan")?.toString().trim();
    const tags = formData.get("tags")?.toString().trim();

    const file = formData.get("foto_konten");

    // =========================
    // VALIDASI DATA
    // =========================

    if (!namaKonten || !judulKonten || !isiKonten) {
      return new Response(
        JSON.stringify({
          success: false,
          message: "Nama konten, judul, dan isi konten wajib diisi.",
        }),
        {
          status: 400,
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
    }

    // =========================
    // UPLOAD FILE
    // =========================

    let namaFile = null;

    if (file && typeof file !== "string" && file.size > 0) {
      const allowedTypes = [
        "image/jpeg",
        "image/png",
        "image/webp",
        "image/jpg",
        "video/mp4",
        "video/webm",
        "video/quicktime",
      ];

      if (!allowedTypes.includes(file.type)) {
        return new Response(
          JSON.stringify({
            success: false,
            message: "Format file tidak didukung.",
          }),
          {
            status: 400,
            headers: {
              "Content-Type": "application/json",
            },
          }
        );
      }

      // Maksimal 50 MB
      const maxSize = 50 * 1024 * 1024;

      if (file.size > maxSize) {
        return new Response(
          JSON.stringify({
            success: false,
            message: "Ukuran file maksimal 50 MB.",
          }),
          {
            status: 400,
            headers: {
              "Content-Type": "application/json",
            },
          }
        );
      }

      // Ambil ekstensi file
      const originalName = file.name || "file";
      const extension = path.extname(originalName).toLowerCase();

      // Nama file dibuat unik
      const safeName = namaKonten
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");

      namaFile = `${safeName}-${Date.now()}${extension}`;

      // Folder penyimpanan
      const uploadDir = path.join(
        process.cwd(),
        "public",
        "uploads",
        "konten"
      );

      await fs.mkdir(uploadDir, {
        recursive: true,
      });

      // Simpan file
      const buffer = Buffer.from(await file.arrayBuffer());

      await fs.writeFile(
        path.join(uploadDir, namaFile),
        buffer
      );
    }

    // =========================
    // SIMPAN KE DATABASE
    // =========================

    const [result] = await db.execute(
      `INSERT INTO tentangkamiswebv5
      (
        Nama_Konten,
        Isi_Konten,
        Foto_Konten,
        Judul_Konten,
        Keterangan,
        tags
      )
      VALUES (?, ?, ?, ?, ?, ?)`,
      [
        namaKonten,
        isiKonten,
        namaFile,
        judulKonten,
        keterangan || null,
        tags || null,
      ]
    );

    // =========================
    // BERHASIL
    // =========================

    return new Response(
      JSON.stringify({
        success: true,
        message: "Konten berhasil disimpan.",
        id: result.insertId,
        file: namaFile,
      }),
      {
        status: 200,
        headers: {
          "Content-Type": "application/json",
        },
      }
    );
  } catch (error) {
    console.error("ERROR SIMPAN KONTEN:", error);

    return new Response(
      JSON.stringify({
        success: false,
        message: "Terjadi kesalahan saat menyimpan konten.",
        error: error.message,
      }),
      {
        status: 500,
        headers: {
          "Content-Type": "application/json",
        },
      }
    );
  }
}