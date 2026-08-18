import { db } from "../../lib/db";
import fs from "node:fs";
import path from "node:path";

export async function POST({ request }) {
    try {
        const body = await request.json();
        let imagePath = body.gambar;

        // Jika gambar dikirim dalam format Base64, ubah jadi file fisik
        if (imagePath && imagePath.startsWith('data:image')) {
            const matches = imagePath.match(/^data:image\/([a-zA-Z0-9]+);base64,(.+)$/);
            
            if (matches && matches.length === 3) {
                const ext = matches[1]; // Ekstensi file (png, jpg, jpeg)
                const base64Data = matches[2];
                
                // Buat nama file unik berdasarkan waktu
                const fileName = `layanan-${Date.now()}.${ext}`;
                const uploadDir = path.join(process.cwd(), 'public/uploads');

                // Pastikan folder public/uploads ada
                if (!fs.existsSync(uploadDir)) {
                    fs.mkdirSync(uploadDir, { recursive: true });
                }

                // Simpan file fisik ke folder public/uploads
                fs.writeFileSync(path.join(uploadDir, fileName), Buffer.from(base64Data, 'base64'));

                // Path relatif yang akan disimpan ke database
                imagePath = `/uploads/${fileName}`;
            }
        }

        await db.query(
        `
        INSERT INTO tempat_layanan
        (
            judul,
            deskripsi,
            gambar,
            alt_text,
            lokasi,
            alamat,
            warna
        )
        VALUES (?,?,?,?,?,?,?)
        `,
        [
            body.judul,
            body.deskripsi || "",
            imagePath, // Path file yang sudah rapi (/uploads/...)
            body.alt_text,
            body.lokasi,
            body.alamat,
            "blue"
        ]);

        return new Response(
            JSON.stringify({
                success: true
            }),
            {
                status: 200,
                headers: {
                    "Content-Type": "application/json"
                }
            }
        );

     } catch (error) { 
        console.log("ERROR DATABASE:", error);

        return new Response(
            JSON.stringify({
                success: false,
                error: error.message
            }),
            {
                status: 500,
                headers: {
                    "Content-Type": "application/json"
                }
            }
        );
    }
}