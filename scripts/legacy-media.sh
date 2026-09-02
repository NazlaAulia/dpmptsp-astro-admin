#!/usr/bin/env bash
#
# Inventory and archive the legacy media still hosted on the old PHP site.
#
# WHY THIS EXISTS
#
# The article images are not in this repository. public/uploads holds a handful
# of files; post.picture references roughly 553. The rest are served from
# dpm-ptsp.surabaya.go.id, which is a system nobody here controls and which will
# eventually be switched off. When it is, those articles lose their images and
# there is no copy anywhere.
#
# That makes this the one piece of work whose deadline is set by someone else.
# It is independent of the migration, the Go API and everything else.
#
#   ./scripts/legacy-media.sh check     report which references are reachable
#   ./scripts/legacy-media.sh archive   download everything reachable
#
# Reads database settings from .env, same variables as everything else.

set -euo pipefail

cd "$(dirname "$0")/.."
[ -f .env ] && set -a && . ./.env && set +a

MEDIA_BASE_URL="${MEDIA_BASE_URL:-https://dpm-ptsp.surabaya.go.id}"
OUT_DIR="${LEGACY_MEDIA_DIR:-./legacy-media}"
REPORT="${OUT_DIR}/inventory.csv"
CONCURRENCY="${LEGACY_MEDIA_CONCURRENCY:-4}"

usage() { sed -n '2,30p' "$0" | sed 's/^# \{0,1\}//'; exit 1; }
[ $# -ge 1 ] || usage
MODE="$1"

# --- collect the referenced filenames -----------------------------------------
#
# Runs inside a container so no mysql/psql client is needed on the host.
collect() {
  case "${DB_CONNECTION:-mysql}" in
    postgres)
      docker run --rm --network "${COMPOSE_NETWORK:-dpmptsp_backend}" \
        -e PGPASSWORD="${DB_PASSWORD}" postgres:16-alpine \
        psql -h "${DB_HOST:-database}" -U "${DB_USERNAME}" -d "${DB_DATABASE}" -tAc \
        "SELECT DISTINCT picture FROM post WHERE picture IS NOT NULL AND picture <> '';"
      ;;
    *)
      docker run --rm --network "${COMPOSE_NETWORK:-dpmptsp_backend}" \
        -e MYSQL_PWD="${DB_PASSWORD}" mysql:8.0 \
        mysql -h "${DB_HOST:-database}" -u"${DB_USERNAME}" -N -e \
        "SELECT DISTINCT picture FROM ${DB_DATABASE}.post WHERE picture IS NOT NULL AND picture <> '';"
      ;;
  esac
}

mkdir -p "$OUT_DIR"

echo "collecting referenced filenames from the database..."
mapfile -t FILES < <(collect | sed '/^$/d')
echo "  ${#FILES[@]} distinct filenames referenced"

if [ "${#FILES[@]}" -eq 0 ]; then
  echo "nothing to do — is the database reachable and populated?"
  exit 0
fi

printf 'filename,status,url\n' > "$REPORT"

probe() {
  local name="$1"
  # The stored value is a bare filename; /fileberita is where they live.
  local url="${MEDIA_BASE_URL}/fileberita/${name}"
  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 20 -I "$url" 2>/dev/null || echo 000)
  printf '%s,%s,%s\n' "\"${name}\"" "$code" "\"${url}\"" >> "$REPORT"

  if [ "$MODE" = "archive" ] && [ "$code" = "200" ]; then
    local dest="${OUT_DIR}/fileberita/${name}"
    mkdir -p "$(dirname "$dest")"
    [ -s "$dest" ] || curl -s --max-time 60 -o "$dest" "$url" || true
  fi
}

echo "probing ${#FILES[@]} files (${CONCURRENCY} at a time)..."
running=0
for f in "${FILES[@]}"; do
  probe "$f" &
  running=$((running + 1))
  # Deliberately gentle: this is someone else's production server, not ours.
  if [ "$running" -ge "$CONCURRENCY" ]; then wait -n; running=$((running - 1)); fi
done
wait

ok=$(grep -c ',200,' "$REPORT" || true)
missing=$(($(wc -l < "$REPORT") - 1 - ok))

echo
echo "report: $REPORT"
echo "  reachable:   $ok"
echo "  unreachable: $missing"
if [ "$MODE" = "archive" ]; then
  echo "  archived to: ${OUT_DIR}/fileberita/"
  echo
  echo "Next: upload these through the API's storage layer so they live on a disk"
  echo "we control, then rewrite post.picture to the new keys."
fi
[ "$missing" -eq 0 ] || echo
[ "$missing" -eq 0 ] || echo "NOTE: $missing references already resolve to nothing. Those images are gone."
