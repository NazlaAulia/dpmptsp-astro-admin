#!/bin/sh
# Selects how the CMS is reached. Runs after 40-admin-policy.sh, which is what
# defines $admin_ip_ok.
set -eu

: "${ADMIN_ROUTING:=separate}"

src="/etc/nginx/routing/${ADMIN_ROUTING}.inc"
if [ ! -f "$src" ]; then
    echo "gateway: unknown ADMIN_ROUTING=${ADMIN_ROUTING}" >&2
    echo "gateway: expected one of separate, path" >&2
    exit 1
fi

# The prefix the CMS answers on. A leading slash, no trailing one.
: "${ADMIN_PREFIX:=/admin}"
case "$ADMIN_PREFIX" in
    /*/) echo "gateway: ADMIN_PREFIX must not end in a slash: $ADMIN_PREFIX" >&2; exit 1 ;;
    /*)  ;;
    *)   echo "gateway: ADMIN_PREFIX must start with a slash: $ADMIN_PREFIX" >&2; exit 1 ;;
esac

# These belong to the public site. Taking one would make part of it
# unreachable, which is the kind of mistake that only shows up in production.
case "$ADMIN_PREFIX" in
    /|/api|/api/*|/_astro|/_astro/*|/storage|/storage/*)
        echo "gateway: ADMIN_PREFIX=$ADMIN_PREFIX collides with the public site." >&2
        exit 1 ;;
esac

sed "s|__PREFIX__|${ADMIN_PREFIX}|g" "$src" > /etc/nginx/conf.d/admin-routing.inc

if [ "$ADMIN_ROUTING" = "path" ]; then
    echo "gateway: admin_routing = path, prefix = ${ADMIN_PREFIX}"
    if [ "${ADMIN_ACCESS:-allowed_ips}" = "public" ]; then
        echo "gateway: the CMS login form is reachable from the internet (ADMIN_ACCESS=public)." >&2
        echo "gateway: rate limited to 5 attempts/minute per address." >&2
    fi
else
    echo "gateway: admin_routing = separate"
fi
