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

# Serving the CMS on the public port with no address restriction puts the
# login form in front of the whole internet. Refuse rather than assume it was
# meant.
if [ "$ADMIN_ROUTING" = "path" ] && [ "${ADMIN_ACCESS:-allowed_ips}" = "public" ]; then
    echo "gateway: ADMIN_ROUTING=path with ADMIN_ACCESS=public would expose the CMS to everyone." >&2
    echo "gateway: pick an ADMIN_ACCESS that restricts by address, or keep ADMIN_ROUTING=separate." >&2
    exit 1
fi

cp "$src" /etc/nginx/conf.d/admin-routing.inc

echo "gateway: admin_routing = ${ADMIN_ROUTING}"
