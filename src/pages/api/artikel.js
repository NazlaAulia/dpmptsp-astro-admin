---
export const prerender = false;

import { db } from "../../lib/db";
import fs from "node:fs/promises";
import path from "node:path";

export async function POST({ request }) {

    const formData = await request.formData();

    const title = formData.get("title")?.toString() || "";
    const category = formData.get("category")?.toString() || "";
    const content = formData.get("content")?.toString() || "";
    const active = formData.get("active")?.toString() || "Y";

    // Ambil file gambar
    const picture = formData.get("picture");

    let namaFile = "";

    if (picture instanceof File && picture.size > 0) {

        // Ambil nama file asli
        namaFile = picture.name;

        // Lokasi folder public/uploads
        const uploadDir = path.join(process.cwd(), "public", "uploads");

        // Pastikan folder uploads ada
        await fs.mkdir(uploadDir, { recursive: true });

        // Simpan file
        const buffer = Buffer.from(await picture.arrayBuffer());

        await fs.writeFile(
            path.join(uploadDir, namaFile),
            buffer
        );
    }

    // Simpan ke database
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
            namaFile
        ]
    );

    return Response.redirect(
        new URL("/admin/artikel", request.url),
        302
    );
}
---