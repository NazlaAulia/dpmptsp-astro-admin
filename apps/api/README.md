# apps/api

The Go backend. Reachable only from the internal docker network (SPEC.md §8);
the Astro apps call it server-side and nothing else does.

## Building without Go installed

Go is deliberately **not** installed on the host. The toolchain lives in the
build image:

    docker build -t dpmptsp-api:dev apps/api
    docker run --rm -v "$PWD/apps/api":/src -w /src golang:1.25-alpine go test ./...

The one thing this costs is editor tooling: `gopls` will not work against
`apps/api` until Go is installed locally. That is a workstation setup task, not
a repository one.

Note `go.mod` requires Go 1.25 — `github.com/jackc/pgx/v5` v5.10 does not build
on 1.23.

## The module path is not a URL

The module is `dpmptsp/api`, not a github.com path. It is private, never
published, and never fetched remotely, so the path only has to be unique within
the build. Naming a specific GitHub account here would bake the current remote
into every import line and force a mass rewrite if the repository moves.

## Configuration

Database settings are discrete, the way Laravel does it. The driver-specific
DSN is built from them, so the connection string never appears in an env file
and switching engines is not a DSN rewrite.

| Variable | Required | Default | Notes |
|---|---|---|---|
| `DB_CONNECTION` | no | `postgres` | `postgres` or `mysql`. `DB_ENGINE` is accepted as an alias |
| `DB_HOST` | no | `database` | |
| `DB_PORT` | no | per engine | 5432 for postgres, 3306 for mysql |
| `DB_DATABASE` | **yes** | — | unless `DATABASE_URL` is set |
| `DB_USERNAME` | **yes** | — | unless `DATABASE_URL` is set |
| `DB_PASSWORD` | no | empty | |
| `DB_SSLMODE` | no | `disable` | postgres only |
| `DB_CHARSET` | no | `utf8mb4` | mysql only |
| `DATABASE_URL` | no | — | escape hatch: when set it replaces all of the above |
| `API_SERVICE_KEY` | **yes** | — | shared key for Astro→Go, min 32 chars |
| `REDIS_URL` | no | `redis://redis:6379/0` | cache is optional at boot |
| `API_ADDR` | no | `:8080` | |
| `APP_ENV` | no | `development` | |
| `SHUTDOWN_TIMEOUT_SECONDS` | no | `15` | |

Every problem is reported at once, at startup, rather than one per restart.

The two drivers disagree about what a DSN even is — pgx wants a URL, and
go-sql-driver/mysql wants `user:pass@tcp(host:port)/db` — which is why building
it internally matters: asking an operator to supply one leaks the engine switch
straight into the environment file.

## File storage

Swappable disks, in the shape Laravel's filesystem takes: several disks are
defined and one is selected by `FILESYSTEM_DISK`. Code asks for a disk and never
knows whether it is writing to local disk or an S3 bucket, so moving uploads to
object storage is a configuration change.

| Disk | Where | URL |
|---|---|---|
| `local` | filesystem, private | none — served through an authorizing handler |
| `public` | filesystem | `STORAGE_PUBLIC_URL` prefix |
| `s3` | any S3-compatible bucket | `S3_PUBLIC_URL`, or presigned |

    FILESYSTEM_DISK       local | public | s3          (default local)
    STORAGE_LOCAL_ROOT    /var/lib/dpmptsp/storage
    STORAGE_PUBLIC_ROOT   /var/lib/dpmptsp/public
    STORAGE_PUBLIC_URL    /storage
    S3_ENDPOINT           host:port, no scheme. The s3 disk exists only when
                          this is set, so a local deployment needs no S3 config.
    S3_BUCKET S3_ACCESS_KEY S3_SECRET_KEY S3_REGION S3_USE_SSL S3_PUBLIC_URL

Misconfiguration fails at startup, not on the first upload:

    FILESYSTEM_DISK=gcs   -> unknown disk: only local, public are configured
    S3_BUCKET=missing     -> bucket "missing" does not exist
    unwritable local root -> root ... is not writable

Only the *selected* disk is verified, so an S3 deployment is not forced to
provide writable local paths it never uses.

`storage.Upload` applies the policy, fixing three things the Astro upload routes
got wrong (see docs/uploads.md):

- the content type is **sniffed from the bytes**, never taken from the client's
  Content-Type — a PNG uploaded as `notes.txt` is stored as `.png`
- the size cap is enforced **while streaming**, so an oversized upload does not
  cost its full size in memory first, and the partial object is deleted
- the stored key is **generated** (random, date-sharded). The client's filename
  is used only for a sanity-checked extension when the type is unrecognised.
  The old code passed it to `path.Join`, which is a path traversal, and named
  files by `Date.now()`, which collides.

## Endpoints

    GET /healthz   liveness. Touches nothing, so a database blip cannot cause
                   the orchestrator to restart a healthy process.
    GET /readyz    readiness. Checks the database and Redis.

Both skip service-key auth so container probes can reach them. Everything else
requires `X-Internal-Key`.

## Dual-engine support

Postgres is the deployment target. MySQL is maintained and verified but not
deployed. Most differences were designed away rather than abstracted — `TEXT` +
`CHECK` instead of `ENUM`, `updated_at` set in Go, `TEXT` instead of `JSONB`,
full instead of partial indexes, `time.Now()` instead of `NOW()`.

What genuinely cannot be shared lives behind `internal/infrastructure/database/dialect`:
placeholder style, `RETURNING` vs `LastInsertId`, full-text search, and sequence
reset. That is roughly 200 lines, against several thousand for two parallel sets
of repositories.

## Verified

Both engines and both storage backends were run end-to-end against throwaway
containers, from the same image:

    DB_CONNECTION=postgres  /readyz {"engine":"postgres","storage":"local",...}
    DB_CONNECTION=mysql     /readyz {"engine":"mysql",...}
    FILESYSTEM_DISK=s3      /readyz {"storage":"s3",...}  against real MinIO

Service key: no key 401, wrong key 401, correct key routes normally.
