export const prerender = false;

import { db } from "../../lib/db";


export async function POST({ request }) {

    const formData = await request.formData();

    const title = formData.get("title");
    const category = formData.get("category");
    const content = formData.get("content");
    const active = formData.get("active");


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
            hits
        )
        VALUES (?, ?, ?, CURDATE(), CURTIME(), ?, ?, ?)
        `,
        [
            category,
            title,
            content,
            1,
            active,
            0
        ]
    );


    return Response.redirect(
        new URL("/admin/artikel", request.url),
        302
    );
}