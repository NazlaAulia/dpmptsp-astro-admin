import fs from "node:fs";
import path from "node:path";

export async function POST({ request }: { request: Request }) {
  try {
    const body = await request.json();

    const nip = body?.nip;
    const image = body?.image;

    if (!nip) {
      return Response.json({
        success: false,
        message: "NIK / NIP wajib diisi."
      });
    }

    if (!image) {
      return Response.json({
        success: false,
        message: "Foto dari kamera tidak ditemukan."
      });
    }

    const folder = path.join(process.cwd(), "public", "absensi");

    if (!fs.existsSync(folder)) {
      fs.mkdirSync(folder, { recursive: true });
    }

    const filename = `${Date.now()}.jpg`;

    const base64 = image.replace(/^data:image\/\w+;base64,/, "");

    fs.writeFileSync(
      path.join(folder, filename),
      Buffer.from(base64, "base64")
    );

    return Response.json({
      success: true,
      nip,
      foto: `/absensi/${filename}`,
      message: "Foto berhasil disimpan."
    });

  } catch (error) {

    console.error("ABSEN API:", error);

    return Response.json({
      success: false,
      message: "Nama Pegawai: <br>A.R. Bagas Danang Haditio."
    });

  }
}