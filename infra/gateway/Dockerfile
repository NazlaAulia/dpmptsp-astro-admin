FROM nginx:1.27-alpine

# nginx:alpine already ships /docker-entrypoint.d/20-envsubst-on-templates.sh,
# which renders /etc/nginx/templates/*.template into /etc/nginx/conf.d/.
COPY templates/ /etc/nginx/templates/
COPY includes/  /etc/nginx/includes/
COPY policy/    /etc/nginx/policy/
COPY docker-entrypoint.d/40-admin-policy.sh /docker-entrypoint.d/40-admin-policy.sh
RUN chmod +x /docker-entrypoint.d/40-admin-policy.sh

# $connection_upgrade is needed by includes/proxy.conf for websocket passthrough.
RUN printf 'map $http_upgrade $connection_upgrade {\n  default upgrade;\n  ""      close;\n}\n' \
      > /etc/nginx/conf.d/upgrade-map.conf
