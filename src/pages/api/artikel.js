export const prerender = false;

import { db } from "../../lib/db";
import fs from "node:fs/promises";
import path from "node:path";

export async function POST({ request }) {

    const formData = await request.formData();

    const title = formData.get("title");
    const category = formData.get("category");
    const content = formData.get("content");
    const active = formData.get("active");
    const picture = formData.get("picture");

    let fileName = null;

    // =========================
    // UPLOAD GAMBAR
    // =========================
    if (picture instanceof File && picture.size > 0) {

        const uploadDir = path.join(process.cwd(), "public", "uploads");

        // Buat folder uploads kalau belum ada
        await fs.mkdir(uploadDir, { recursive: true });

        // Buat nama file unik
        const ext = path.extname(picture.name);
        fileName = `${Date.now()}${ext}`;

        // Simpan file
        const filePath = path.join(uploadDir, fileName);
        const buffer = Buffer.from(await picture.arrayBuffer());

        await fs.writeFile(filePath, buffer);
    }


    // =========================
    // SIMPAN KE DATABASE
    // =========================
    await db.query(
        `
        INSERT INTO post
        (
            id_category,
            title,
            content,
            date,
            time,
            editor,
            active,
            hits,
            picture
        )
        VALUES (?, ?, ?, CURDATE(), CURTIME(), ?, ?, ?, ?)
        `,
        [
            category,
            title,
            content,
            1,
            active,
            0,
            fileName
        ]
    );


    return Response.redirect(
        new URL("/admin/artikel", request.url),
        302
    );
}