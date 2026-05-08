#!/bin/sh
set -e

# Prefer IPv4 nameserver; fall back to IPv6 in brackets (nginx requires brackets for IPv6)
NGINX_RESOLVER=$(awk '/^nameserver/ && $2 !~ /:/ {print $2; exit}' /etc/resolv.conf)
if [ -z "$NGINX_RESOLVER" ]; then
  NGINX_RESOLVER=$(awk '/^nameserver/ && $2 ~ /:/ {print "[" $2 "]"; exit}' /etc/resolv.conf)
fi
if [ -z "$NGINX_RESOLVER" ]; then
  echo "ERROR: no nameserver found in /etc/resolv.conf" >&2
  exit 1
fi

export NGINX_RESOLVER
envsubst '${BACKEND_HOST} ${PORT} ${NGINX_RESOLVER}' \
  < /etc/nginx/templates/default.conf.template \
  > /etc/nginx/conf.d/default.conf

exec nginx -g 'daemon off;'
