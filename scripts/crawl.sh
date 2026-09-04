#!/usr/bin/env bash
# Renders every page and checks the BODY, not just the status.
#
# Astro streams: when frontmatter throws mid-render the headers are already
# sent, so the response is 200 with the error text in the middle of the page.
# Every check that looked only at the status code passed while the homepage
# said "Internal server error".
set -uo pipefail

WEB=${WEB_URL:-http://127.0.0.1:8017}
ADMIN=${ADMIN_URL:-http://127.0.0.1:8018}
ROOT=$(cd "$(dirname "$0")/.." && pwd)
BAD=0

broken() {
  grep -qiE "Internal server error|ReferenceError|is not defined|Cannot read properties of" <<<"$1"
}

visit() {
  local url=$1 label=$2 jar=${3:-}
  local body code
  body=$(curl -s ${jar:+-b "$jar"} "$url")
  code=$(curl -s -o /dev/null -w '%{http_code}' ${jar:+-b "$jar"} "$url")
  if broken "$body"; then
    printf '  BROKEN %s  %s\n' "$code" "$label"
    BAD=$((BAD + 1))
  else
    printf '  ok     %s  %s\n' "$code" "$label"
  fi
}

echo "web:"
for f in "$ROOT"/apps/web/src/pages/*.astro; do
  name=$(basename "$f" .astro)
  [ "$name" = index ] && path=/ || path="/$name"
  visit "$WEB$path" "$path"
done

slug=$(curl -s "$WEB/beritalist" | grep -o 'href="/media/[^"]*"' | head -1 | sed 's|href="/media/||;s|"||')
[ -n "$slug" ] && visit "$WEB/media/$slug" "/media/$slug"

if [ -n "${ADMIN_USER:-}" ] && [ -n "${ADMIN_PASSWORD:-}" ]; then
  echo "admin:"
  jar=$(mktemp)
  trap 'rm -f "$jar"' EXIT
  curl -s -b "$jar" -c "$jar" -o /dev/null -X POST "$ADMIN/admin/login" \
    -H "Origin: $ADMIN" \
    --data-urlencode "username=$ADMIN_USER" --data-urlencode "password=$ADMIN_PASSWORD"

  for f in "$ROOT"/apps/admin/src/pages/admin/*.astro; do
    name=$(basename "$f" .astro)
    case "$name" in login | logout) continue ;; esac
    path="/admin/$name"
    case "$name" in edit-* | hapus-*) path="$path?id=1" ;; esac
    visit "$ADMIN$path" "$path" "$jar"
  done
else
  echo "admin: skipped (set ADMIN_USER and ADMIN_PASSWORD to include it)"
fi

if [ "$BAD" -gt 0 ]; then
  echo "$BAD page(s) rendered an error"
  exit 1
fi
echo "every page rendered"
