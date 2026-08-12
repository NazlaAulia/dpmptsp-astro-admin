import { db } from "../../../lib/db";

export async function DELETE({ params }) {

    console.log("ID YANG DIHAPUS:", params.id);


    await db.query(
        `
        DELETE FROM twdata
        WHERE id = ?
        `,
        [params.id]
    );


    return new Response(
        JSON.stringify({
            success:true
        }),
        {
            status:200
        }
    );

}