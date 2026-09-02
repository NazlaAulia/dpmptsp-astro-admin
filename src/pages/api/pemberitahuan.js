import { db } from "../../lib/db";

export async function POST({ request }) {
  const data = await request.json();
  const { judul, foto, link_url, status_aktif, tipe } = data;

  try {
    await db.query(
      `INSERT INTO pemberitahuan (judul, foto, link_url, status_aktif, tipe)
       VALUES (?, ?, ?, ?, ?, ?)`,
      [judul, foto || null, link_url || null, status_aktif || "Y", tipe || "notif"],
    );
    return new Response(JSON.stringify({ success: true }), {
      status: 201,
      headers: { "Content-Type": "application/json" },
    });
  } catch (error) {
    console.error("Error creating notification:", error);
    return new Response(JSON.stringify({ success: false, error: error.message }), {
      status: 500,
      headers: { "Content-Type": "application/json" },
    });
  }
}

export async function DELETE({ request }) {
  const data = await request.json();
  const { id } = data;

  try {
    await db.query("DELETE FROM pemberitahuan WHERE id = ?", [id]);
    return new Response(JSON.stringify({ success: true }), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  } catch (error) {
    console.error("Error deleting notification:", error);
    return new Response(JSON.stringify({ success: false, error: error.message }), {
      status: 500,
      headers: { "Content-Type": "application/json" },
    });
  }
}
