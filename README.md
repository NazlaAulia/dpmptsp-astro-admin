# DPMPTSP Kota Surabaya

Public site, admin panel and API for the Dinas Penanaman Modal dan Pelayanan
Terpadu Satu Pintu.

```
apps/web        public site        Astro, full SSR
apps/admin      CMS                Astro, full SSR
apps/api        backend            Go, internal network only
packages/ui     shared primitives
packages/config env access
infra/          gateway and container definitions
```

The browser never talks to `apps/api`. It talks to Astro, and Astro talks to Go
over the internal network.

## Running it

```bash
cp .env.example .env      # then fill in the secrets it asks for
make up                   # production-shaped stack
make dev                  # same, with hot reload
make smoke                # check it came up, and that the API stays isolated
```

Nothing beyond Docker, Node 22 and pnpm is required. **Go and golang-migrate are
deliberately not installed on the host** — both run in containers, so `make
test`, `make lint` and `make migrate-*` work on a machine that has neither. The
only cost is that `gopls` will not work against `apps/api` until Go is installed
locally, which is a workstation setup task rather than a repository one.

`make help` lists everything.

## Database

Settings are discrete, as in Laravel, and the driver-specific DSN is built from
them:

```
DB_CONNECTION=mysql        # or postgres
DB_HOST=database
DB_DATABASE=ladpm
DB_USERNAME=root
DB_PASSWORD=…
```

Both engines are supported. Postgres is the target; MySQL is maintained and
verified but not deployed. Two migration sets exist because no single SQL file
is valid on both, so `make schema-check` proves they still agree — it migrates a
throwaway database of each kind from empty and diffs the result.

```bash
make migrate-up
make migrate-fresh        # drop, migrate from empty, seed
make schema-check
```

## Seeding

```bash
SEED_ADMIN_PASSWORD=… make seed   # reference data; re-runs skip what exists
make seed-list                    # show what would run
make seed-fresh                   # clear each seeder's rows, re-insert
```

Seeds are YAML, written once rather than once per dialect. Each declares the
columns that identify a row, which is what makes a second run a no-op.

## Where this is in the migration

The site currently runs on MySQL, queried directly from Astro page frontmatter.
The Go API exists and is deployed alongside, but does not yet own the data. The
end state is no database access anywhere in Astro, and the signal for it is:

```bash
grep -rn "mysql2\|lib/db" apps/     # must return nothing
```

Until then `apps/*/src/lib/db.js` is a deprecated shim. Do not add queries to
it, and do not promote it into `packages/` — that would make an architecture
violation look like shared infrastructure.

Repositories are the seam. Pages call a service, the service calls a repository,
and the repository is the only thing that changes when a resource moves to the
API.

## Two things that need attention from outside this repository

**The legacy article images are on someone else's server.** `post.picture`
references roughly 553 files served from `dpm-ptsp.surabaya.go.id`. They are not
in this repository and there is no copy. When that host is switched off the
images are gone. `make legacy-media` reports what is still reachable and
`make legacy-media-archive` downloads it. This has a deadline set by someone
else and is independent of everything else here.

**Credentials in git history.** Two database dumps and a session cookie file
were committed and later removed. `git rm` clears the working tree, not the
history, so both remain retrievable at earlier commits. Treat their contents as
disclosed and rotate accordingly; purging history requires a force-push and is
a decision for whoever owns the remote.
