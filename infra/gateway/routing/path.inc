# ADMIN_ROUTING=path — the CMS is also served at /admin on the public port.
#
# Both Astro apps publish assets under /_astro/ and endpoints under /api/, so
# on a shared hostname those prefixes collide. Each is resolved explicitly
# below; nothing is left to guess.

# The CMS itself. Same IP policy as the admin port — leaving the path open
# would make the port's restriction pointless.
location ^~ /admin {
    if ($admin_ip_ok = 0) { return 403; }

    set $admin_upstream admin:3000;
    proxy_pass http://$admin_upstream;
    include /etc/nginx/includes/proxy.conf;
}

# The admin app's own form endpoints. Listed one by one rather than by prefix:
# /api/kontak is the public contact form and must stay reachable by everyone.
location ~ ^/api/(artikel|layanan|hapus-layanan|kinerja|pemberitahuan)(/|$) {
    if ($admin_ip_ok = 0) { return 403; }

    set $admin_upstream admin:3000;
    proxy_pass http://$admin_upstream;
    include /etc/nginx/includes/proxy.conf;
}

# Build assets. Both apps emit hashed filenames under /_astro/, and a request
# carries nothing that says which app it belongs to. Ask the web app first and
# fall back to the CMS on a miss: the names are content-hashed, so a file that
# exists in one is absent from the other.
#
# Not gated by the IP policy: these are stylesheets and scripts, and gating
# them would only break the CMS for the addresses already allowed to use it.
location ^~ /_astro/ {
    set $web_upstream web:3000;
    proxy_pass http://$web_upstream;
    include /etc/nginx/includes/proxy.conf;

    proxy_intercept_errors on;
    error_page 404 = @admin_assets;
}

location @admin_assets {
    set $admin_upstream admin:3000;
    proxy_pass http://$admin_upstream;
    include /etc/nginx/includes/proxy.conf;
}
