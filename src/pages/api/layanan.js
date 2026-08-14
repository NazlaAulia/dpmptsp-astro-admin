import { db } from "../../lib/db";


export async function POST({request}){

    try{

        const body = await request.json();


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
            body.deskripsi,
            body.gambar,
            body.alt_text,
            body.lokasi,
            body.alamat,
            "blue"
        ]);



        return new Response(
            JSON.stringify({
                success:true
            }),
            {
                status:200,
                headers:{
                    "Content-Type":"application/json"
                }
            }
        );


     } catch(error){ 

    console.log("ERROR DATABASE:", error);


    return new Response(
        JSON.stringify({
            success:false,
            error:error.message
        }),
        {
            status:500,
            headers:{
                "Content-Type":"application/json"
            }
        }
    );

}

}