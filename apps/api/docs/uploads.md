# Upload rules to implement in Go

Salvaged from `apps/admin/src/pages/api/konten.js` before that endpoint was
deleted as an orphan. It had no callers, but it held the only real upload
validation in the codebase and the rules are worth keeping.

## What that endpoint got right

**Content-type allowlist** — rejected anything not in:

    image/jpeg  image/jpg  image/png  image/webp
    video/mp4   video/webm video/quicktime

**Size cap** — 50 MB, enforced server-side.

**Generated filename** — the client's filename was never used as the stored
name. It slugified a caller-supplied title and appended a timestamp:

    safeName = title.toLowerCase()
                    .replace(/[^a-z0-9]+/g, "-")
                    .replace(/^-+|-+$/g, "")
    stored   = `${safeName}-${Date.now()}${ext}`

Only the extension came from the upload.

## What it got wrong, and Go must not repeat

- **`file.type` is the client's claim**, not a fact. Sniff the magic bytes and
  derive the extension from what the file actually is. A `.jpg` that is really
  an HTML document is stored and later served.
- **The size check happens after the body is in memory.** Enforce the cap while
  streaming, or a 2 GB upload costs 2 GB of RAM before being rejected.
- **`Date.now()` is not unique.** Two uploads in the same millisecond collide.
  Use a UUID or a content hash.

## What the other upload paths did, which is worse

These are the endpoints that survived, and they are the reason uploads must move
to Go rather than be tidied in place:

| Path | Problem |
|---|---|
| `api/artikel.js` | `path.join(uploadDir, file.name)` with the **raw client filename** — path traversal, and no auth on the route until the middleware landed |
| `admin/tambah-kinerja.astro`, `admin/edit-kinerja.astro` | same raw-filename write, into `public/pdf` |
| `api/layanan.js` | base64 data-URL inside a JSON body, decoded server-side. Inflates the payload ~33% and forces the whole file into a JS string |
| `api/kinerja.js` *(deleted)* | inserted `file.name` into `twdata.file_path` and **never wrote the file** |
| `api/inovasi.ts` *(deleted)* | stored `gambar/<name>` as a path for a file it never wrote |

## Stored-path convention

Three incompatible conventions exist in the data today and a data migration has
to normalise them:

    post.picture              bare filename          "foto.jpg"
    twdata.file_path          bare filename          "laporan.pdf"
    tentang_kami.foto         bare filename          "x-123.jpg"
    tempat_layanan.gambar     rooted path            "/uploads/x.png"
    inovasi_layanan.gambar    phantom prefix         "gambar/x.jpg"  (file absent)

Pick one — an opaque object key — and migrate the existing rows to it.

## Where files should live

Not under `public/`. Writing into the web root means the process needs write
access to the directory it serves, and it does not survive more than one replica.
MinIO already runs on this host and is the correct destination; the interface in
`internal/infrastructure/storage` should keep local disk and object storage
interchangeable.
