import { db } from "../../lib/db";


export async function POST({request}){


const formData = await request.formData();


const judul = formData.get("tw_option");
const file = formData.get("file");


// nama file
const fileName = file.name;



await db.query(
`
INSERT INTO twdata
(
tw_option,
file_path
)
VALUES (?,?)
`,
[
judul,
fileName
]
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

export async function DELETE({ params }) {

    const id = Number(params.id);

    if (!id) {

        return new Response(
            JSON.stringify({
                success: false,
                message: "ID tidak valid"
            }),
            {
                status: 400
            }
        );

    }


    const [result] = await db.query(
        `
        DELETE FROM twdata
        WHERE id = ?
        `,
        [id]
    );


    return new Response(
        JSON.stringify({
            success: true,
            message: "Data berhasil dihapus"
        }),
        {
            status: 200
        }
    );

}