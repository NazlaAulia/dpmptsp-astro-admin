# ADMIN_ROUTING=path — the CMS is served under __PREFIX__ on the public port.
#
# The prefix is a setting, so the CMS can move to /cms or /kelola without
# rebuilding anything. Internally the app's routes are always /admin/*: the
# rewrite maps the chosen prefix onto them, and the response is rewritten on
# the way back so links in the page point at the prefix the visitor used.
#
# Both Astro apps publish under /_astro/ and /api/, and on a shared hostname
# those collide. Each is resolved explicitly below.

# The CMS.
location ^~ __PREFIX__/ {
    if ($admin_ip_ok = 0) { return 403; }

    # Before the rewrite: `break` ends the rewrite phase, and `set` runs in
    # that same phase, so afterwards it would never execute.
    set $admin_upstream admin:3000;
    rewrite ^__PREFIX__(/.*)$ /admin$1 break;

    proxy_pass http://$admin_upstream;
    include /etc/nginx/includes/proxy.conf;

    # Redirects and links come back naming /admin; the visitor knows the
    # prefix instead.
    proxy_redirect /admin/ $public_scheme://$http_host__PREFIX__/;
    proxy_set_header Accept-Encoding "";
    sub_filter '/admin/' '__PREFIX__/';
    sub_filter_once off;
}

# Without the trailing slash, so __PREFIX__ itself reaches the CMS.
location = __PREFIX__ {
    return 302 __PREFIX__/dashboard;
}

# The login form. Same handling, plus a rate limit: this endpoint faces the
# internet, so a guess has to cost more than one HTTP request. 5/minute is far
# above what a person types and far below what a script needs.
location = __PREFIX__/login {
    if ($admin_ip_ok = 0) { return 403; }

    limit_req zone=cms_login burst=5 nodelay;
    limit_req_status 429;

    # Before the rewrite: `break` ends the rewrite phase, and `set` runs in
    # that same phase, so afterwards it would never execute.
    set $admin_upstream admin:3000;
    rewrite ^__PREFIX__(/.*)$ /admin$1 break;

    proxy_pass http://$admin_upstream;
    include /etc/nginx/includes/proxy.conf;

    proxy_redirect /admin/ $public_scheme://$http_host__PREFIX__/;
    proxy_set_header Accept-Encoding "";
    sub_filter '/admin/' '__PREFIX__/';
    sub_filter_once off;
}

# The CMS's own form endpoints. Listed one by one rather than by prefix:
# /api/kontak is the public contact form and must stay reachable by everyone.
location ~ ^/api/(artikel|layanan|hapus-layanan|kinerja|pemberitahuan)(/|$) {
    if ($admin_ip_ok = 0) { return 403; }

    set $admin_upstream admin:3000;
    proxy_pass http://$admin_upstream;
    include /etc/nginx/includes/proxy.conf;
}

# Build assets. Both apps emit content-hashed names under /_astro/, and a
# request carries nothing saying which app it belongs to. Ask the web app
# first and fall back to the CMS on a miss: a file present in one is absent
# from the other.
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
