#!/bin/sh
set -eu

case "${PAISA_PUBLIC_API_URL:-}" in
  http://*|https://*) ;;
  *)
    echo "PAISA_PUBLIC_API_URL must be an http(s) URL." >&2
    exit 1
    ;;
esac

escaped_url="$(printf '%s' "$PAISA_PUBLIC_API_URL" | sed 's/[\\"]/\\&/g')"
printf 'globalThis.__PAISA_CONFIG__ = { apiBaseUrl: "%s" };\n' "$escaped_url" > /srv/config.js
exec "$@"
