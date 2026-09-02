import { db } from "../../lib/db";

export async function DELETE({ request }) {
  try {
    const { id } = await request.json();

    console.log("ID LAYANAN YANG DIHAPUS:", id);

    await db.query(
      `
      DELETE FROM tempat_layanan
      WHERE id = ?
      `,
      [id]
    );

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
    console.error("Gagal menghapus layanan:", error);

    return new Response(
      JSON.stringify({
        success: false,
        message: "Gagal menghapus layanan"
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