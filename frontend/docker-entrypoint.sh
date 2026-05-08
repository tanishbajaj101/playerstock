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

# Wait for backend to be reachable before nginx starts (prevents DNS_PROBE_FINISHED_NXDOMAIN)
i=0
until wget -qO- "http://${BACKEND_HOST}:8080/healthz" > /dev/null 2>&1; do
  i=$((i+1))
  if [ $i -ge 60 ]; then
    echo "[stakestock] backend failed to start after 60s, aborting." >&2
    exit 1
  fi
  sleep 1
done

exec nginx -g 'daemon off;'
