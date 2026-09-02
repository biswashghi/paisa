#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: PAISA_WEB_DOMAIN=... PAISA_API_DOMAIN=... PAISA_API_IMAGE=... PAISA_WEB_IMAGE=... $0 <deploy-user> <server-ip> [repo-url] [branch]"
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then usage; exit 0; fi
if [[ $# -lt 2 || $# -gt 4 ]]; then usage >&2; exit 1; fi

DEPLOY_USER="$1"
SERVER_IP="$2"
API_DOMAIN="${PAISA_API_DOMAIN:-${APP_API_DOMAIN:-}}"
WEB_DOMAIN="${PAISA_WEB_DOMAIN:-${APP_DOMAIN:-}}"
API_IMAGE="${PAISA_API_IMAGE:-}"
WEB_IMAGE="${PAISA_WEB_IMAGE:-}"
VERIFY_PUBLIC_DEPLOYMENT="${VERIFY_PUBLIC_DEPLOYMENT:-1}"

for domain in "$API_DOMAIN" "$WEB_DOMAIN"; do
  [[ "$domain" =~ ^[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]] || { echo "Paisa domains must be valid DNS names." >&2; exit 1; }
done
for image in "$API_IMAGE" "$WEB_IMAGE"; do
  [[ "$image" =~ ^ghcr\.io/[a-z0-9._/-]+@sha256:[a-f0-9]{64}$ ]] || { echo "Paisa images must be immutable GHCR digests." >&2; exit 1; }
done
[[ -n "${PAISA_POSTGRES_PASSWORD:-}" && -n "${PAISA_INTERNAL_ADMIN_TOKEN:-}" ]] || { echo "PAISA_POSTGRES_PASSWORD and PAISA_INTERNAL_ADMIN_TOKEN are required." >&2; exit 1; }
[[ "$VERIFY_PUBLIC_DEPLOYMENT" =~ ^[01]$ ]] || { echo "VERIFY_PUBLIC_DEPLOYMENT must be 0 or 1." >&2; exit 1; }

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PAYLOAD_FILE="$(mktemp)"
trap 'rm -f "$PAYLOAD_FILE"' EXIT
chmod 0600 "$PAYLOAD_FILE"
printf '%s\n' \
  "PAISA_API_IMAGE=${API_IMAGE}" \
  "PAISA_WEB_IMAGE=${WEB_IMAGE}" \
  "PAISA_API_DOMAIN=${API_DOMAIN}" \
  "PAISA_WEB_DOMAIN=${WEB_DOMAIN}" \
  "PAISA_POSTGRES_DB=${PAISA_POSTGRES_DB:-project}" \
  "PAISA_POSTGRES_USER=${PAISA_POSTGRES_USER:-project}" \
  "PAISA_POSTGRES_PASSWORD=${PAISA_POSTGRES_PASSWORD}" \
  "PAISA_INTERNAL_ADMIN_TOKEN=${PAISA_INTERNAL_ADMIN_TOKEN}" \
  "PAISA_ALLOWED_ORIGINS=https://${WEB_DOMAIN}" \
  "PAISA_PUBLIC_API_URL=https://${API_DOMAIN}" >"$PAYLOAD_FILE"

scp -q "$ROOT_DIR/compose.yml" "${DEPLOY_USER}@${SERVER_IP}:/tmp/paisa-compose.yml"
scp -q "$ROOT_DIR/compose.production.yml" "${DEPLOY_USER}@${SERVER_IP}:/tmp/paisa-compose.production.yml"
scp -q "$ROOT_DIR/deploy/paisa.caddy.template" "${DEPLOY_USER}@${SERVER_IP}:/tmp/paisa.caddy.template"
scp -q "$ROOT_DIR/scripts/remote/deploy-production.sh" "${DEPLOY_USER}@${SERVER_IP}:/tmp/paisa-deploy-production.sh"
scp -q "$PAYLOAD_FILE" "${DEPLOY_USER}@${SERVER_IP}:/tmp/paisa-app.env"

ssh "${DEPLOY_USER}@${SERVER_IP}" \
  API_DOMAIN="$API_DOMAIN" WEB_DOMAIN="$WEB_DOMAIN" \
  VERIFY_PUBLIC_DEPLOYMENT="$VERIFY_PUBLIC_DEPLOYMENT" \
  'bash /tmp/paisa-deploy-production.sh'
