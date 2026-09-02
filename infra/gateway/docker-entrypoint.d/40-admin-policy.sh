#!/bin/sh
# Selects the admin access policy. Numbered 40 so it runs after nginx's own
# 20-envsubst-on-templates.sh.
set -eu

: "${ADMIN_ACCESS:=allowed_ips}"

src="/etc/nginx/policy/${ADMIN_ACCESS}.conf"
if [ ! -f "$src" ]; then
    echo "gateway: unknown ADMIN_ACCESS=${ADMIN_ACCESS}" >&2
    echo "gateway: expected one of public, allowed_ips, vps_only, vpn_only" >&2
    exit 1
fi

# Fail loudly rather than silently falling back to a permissive policy. A typo
# in an access-control value must not become "open to the world".
if [ "$ADMIN_ACCESS" = "allowed_ips" ] && [ ! -s /etc/nginx/policy/allowed_ips.list ]; then
    echo "gateway: ADMIN_ACCESS=allowed_ips but policy/allowed_ips.list is missing or empty." >&2
    echo "gateway: copy allowed_ips.list.example and add at least one entry." >&2
    exit 1
fi

if [ "$ADMIN_ACCESS" = "vpn_only" ] && ! grep -qE '^[[:space:]]*[0-9]' /etc/nginx/policy/vpn_only.conf; then
    echo "gateway: ADMIN_ACCESS=vpn_only but no VPN CIDR is configured." >&2
    echo "gateway: this would deny everyone. Add the CIDR first." >&2
    exit 1
fi

cp "$src" /etc/nginx/conf.d/admin-policy.inc

# 'public' defines no $admin_ip_ok, so give the template one to read.
if [ "$ADMIN_ACCESS" = "public" ]; then
    echo 'set $admin_ip_ok 1;' > /etc/nginx/conf.d/admin-policy-default.inc
    printf 'map $host $admin_ip_ok { default 1; }\n' > /etc/nginx/conf.d/admin-policy.inc
fi

echo "gateway: admin_access policy = ${ADMIN_ACCESS}"
