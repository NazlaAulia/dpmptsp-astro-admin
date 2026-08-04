export async function POST({ request }) {
    try {
        const body = await request.json();

        const { nip, unit, jawaban } = body;

        if (!nip || !unit) {
            return new Response(JSON.stringify({
                status: "error",
                message: "NIP dan Unit wajib diisi"
            }), {
                status: 400,
                headers: {
                    "Content-Type": "application/json"
                }
            });
        }

        // simulasi simpan data
        console.log("DATA MASUK:", body);

        return new Response(JSON.stringify({
            status: "success",
            message: "Data berhasil disimpan",
            data: body
        }), {
            status: 200,
            headers: {
                "Content-Type": "application/json"
            }
        });

    } catch (err) {
        return new Response(JSON.stringify({
            status: "error",
            message: "Server error",
            error: err.message
        }), {
            status: 500,
            headers: {
                "Content-Type": "application/json"
            }
        });
    }
}