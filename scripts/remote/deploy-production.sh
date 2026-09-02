#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/paisa}"
ENV_FILE="${ENV_FILE:-/etc/paisa/app.env}"
PAYLOAD_FILE="${PAYLOAD_FILE:-/tmp/paisa-app.env}"
ROUTE_TEMPLATE="${ROUTE_TEMPLATE:-/tmp/paisa.caddy.template}"
API_DOMAIN="${API_DOMAIN:-}"
WEB_DOMAIN="${WEB_DOMAIN:-}"
VERIFY_PUBLIC_DEPLOYMENT="${VERIFY_PUBLIC_DEPLOYMENT:-1}"
PREVIOUS_API_IMAGE=""
PREVIOUS_WEB_IMAGE=""

compose() {
  sudo docker compose -p paisa --env-file "$ENV_FILE" \
    -f "$APP_DIR/compose.yml" -f "$APP_DIR/compose.production.yml" "$@"
}

validate_inputs() {
  [[ "$API_DOMAIN" =~ ^[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]]
  [[ "$WEB_DOMAIN" =~ ^[A-Za-z0-9.-]+\.[A-Za-z]{2,}$ ]]
  [[ "$VERIFY_PUBLIC_DEPLOYMENT" =~ ^[01]$ ]]
  test -s "$PAYLOAD_FILE"
  test -s "$ROUTE_TEMPLATE"
}

acquire_lock() {
  exec 9>/tmp/vps-deploy-paisa.lock
  flock 9
  sudo /usr/local/bin/vps-platform-check
}

prepare_release() {
  PREVIOUS_API_IMAGE="$(sudo sed -n 's/^PAISA_API_IMAGE=//p' "$ENV_FILE" 2>/dev/null || true)"
  PREVIOUS_WEB_IMAGE="$(sudo sed -n 's/^PAISA_WEB_IMAGE=//p' "$ENV_FILE" 2>/dev/null || true)"
  sudo install -d -m 0755 "$APP_DIR" "$APP_DIR/scripts"
  sudo install -m 0644 /tmp/paisa-compose.yml "$APP_DIR/compose.yml"
  sudo install -m 0644 /tmp/paisa-compose.production.yml "$APP_DIR/compose.production.yml"
  sudo install -m 0755 /tmp/paisa-deploy-production.sh "$APP_DIR/scripts/deploy-production.sh"
  sudo install -d -m 0750 /etc/paisa
  sudo install -m 0600 "$PAYLOAD_FILE" "$ENV_FILE"
}

rollback_release() {
  [[ -n "$PREVIOUS_API_IMAGE" && -n "$PREVIOUS_WEB_IMAGE" ]] || return 0
  echo "Rolling Paisa back to its previous images." >&2
  sudo sed -i "s#^PAISA_API_IMAGE=.*#PAISA_API_IMAGE=${PREVIOUS_API_IMAGE}#" "$ENV_FILE"
  sudo sed -i "s#^PAISA_WEB_IMAGE=.*#PAISA_WEB_IMAGE=${PREVIOUS_WEB_IMAGE}#" "$ENV_FILE"
  compose pull paisa-api paisa-web || true
  compose up -d --no-build || true
}

deploy_release() {
  compose pull paisa-api paisa-web paisa-postgres
  compose up -d --no-build
}

wait_for_health() {
  local attempt
  for attempt in $(seq 1 30); do
    if compose exec -T paisa-api wget -qO- http://127.0.0.1:8080/health </dev/null >/dev/null; then
      return 0
    fi
    if [[ "$attempt" -eq 30 ]]; then
      compose logs --tail=100 paisa-postgres paisa-api paisa-web >&2 || true
      return 1
    fi
    sleep 2
  done
}

install_route() {
  sed -e "s/{{PAISA_WEB_DOMAIN}}/${WEB_DOMAIN}/g" \
    -e "s/{{PAISA_API_DOMAIN}}/${API_DOMAIN}/g" "$ROUTE_TEMPLATE" | \
    sudo /usr/local/bin/vps-route paisa
}

verify_public() {
  if [[ "$VERIFY_PUBLIC_DEPLOYMENT" == "1" ]]; then
    curl -fsS --retry 6 --retry-delay 5 "https://${API_DOMAIN}/health" >/dev/null &&
      curl -fsSI --retry 6 --retry-delay 5 "https://${WEB_DOMAIN}/" >/dev/null
  fi
}

show_status() { compose ps; }
announce_success() { echo "Paisa deployed from immutable images."; }

main() {
  validate_inputs
  acquire_lock
  prepare_release
  if ! deploy_release; then rollback_release; return 1; fi
  if ! wait_for_health; then rollback_release; return 1; fi
  if ! install_route; then rollback_release; return 1; fi
  if ! verify_public; then rollback_release; return 1; fi
  show_status
  announce_success
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  trap 'rm -f /tmp/paisa-app.env /tmp/paisa.caddy.template /tmp/paisa-deploy-production.sh' EXIT
  main "$@"
fi
